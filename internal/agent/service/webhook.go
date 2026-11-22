package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/telegram"
)

var (
	waDigitsMentionRe = regexp.MustCompile(`(?i)@\d+\b`)
	waGabrielRe       = regexp.MustCompile(`(?i)@gabriel\b`)
	multiSpaceRe      = regexp.MustCompile(`[ \t]{2,}`)
)

// webhookTurnTimeout bounds a background channel turn.
const webhookTurnTimeout = 5 * time.Minute

// dedupCache remembers recently seen webhook message keys so provider retries are ignored.
// ponytail: in-memory per-process; move to DB/redis if running multiple replicas.
type dedupCache struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func (d *dedupCache) firstSeen(key string) bool {
	if key == "" {
		return true
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	if d.seen == nil {
		d.seen = make(map[string]time.Time)
	}
	if len(d.seen) > 2048 {
		for k, t := range d.seen {
			if now.Sub(t) > 10*time.Minute {
				delete(d.seen, k)
			}
		}
	}
	if t, ok := d.seen[key]; ok && now.Sub(t) < 10*time.Minute {
		return false
	}
	d.seen[key] = now
	return true
}

// runAsync runs fn detached from the webhook request so the provider gets a fast 200.
func (s *Service) runAsync(fn func(ctx context.Context) error) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), webhookTurnTimeout)
		defer cancel()
		if err := fn(ctx); err != nil {
			s.logger.WarnContext(ctx, "agent channel turn failed", "err", err)
		}
	}()
}

// ChannelUserResolver resolves linked messaging channel identities and clients.
type ChannelUserResolver interface {
	ResolveByChannel(ctx context.Context, provider, subject string, overrideParishID uuid.UUID) (authx.Principal, error)
	SwitchChannelParish(ctx context.Context, provider, subject string, parishID uuid.UUID) error
	TelegramClient() *telegram.Client
	GowaClient() *gowa.Client
}

// WhatsAppInbound is the normalized inbound GOWA WhatsApp webhook message.
type WhatsAppInbound struct {
	Event         string
	From          string
	ChatID        string
	SenderID      string
	Participant   string
	FromMe        bool
	Text          string
	Caption       string
	MentionedJIDs []string
	HasMedia      bool
	MessageID     string
}

// HandleTelegramUpdate routes non-/start Telegram bot updates through the AI agent.
func (s *Service) HandleTelegramUpdate(ctx context.Context, update telegram.Update) error {
	if s.channelResolver == nil || update.Message == nil {
		return nil
	}
	msg := update.Message
	if msg.Chat.ID == 0 || strings.EqualFold(msg.Chat.Type, "channel") || msg.From == nil {
		return nil
	}
	tgClient := s.channelResolver.TelegramClient()
	if tgClient == nil {
		return nil
	}

	isGroup := msg.Chat.Type == "group" || msg.Chat.Type == "supergroup"
	rawText := strings.TrimSpace(msg.Text)
	if rawText == "" {
		rawText = strings.TrimSpace(msg.Caption)
	}
	botUser, _ := tgClient.ResolveBotUsername(ctx)
	if isGroup && !hasTelegramMention(rawText, msg.Entities, msg.CaptionEntities, botUser) {
		return nil
	}
	promptText := stripTelegramBotMention(rawText, botUser)

	senderSubject := strconv.FormatInt(msg.From.ID, 10)
	if msg.From.ID == 0 {
		senderSubject = strconv.FormatInt(msg.Chat.ID, 10)
	}

	if !s.dedup.firstSeen("tg:" + strconv.FormatInt(update.UpdateID, 10)) {
		return nil
	}
	hasMedia := msg.HasMedia()
	s.runAsync(func(ctx context.Context) error {
		sendText := func(txt string) { _ = tgClient.SendMessage(ctx, msg.Chat.ID, txt) }
		sendTyping := func() { _ = tgClient.SendChatAction(ctx, msg.Chat.ID, "typing") }
		return s.handleChannelMessage(ctx, domain.ChannelTelegram, senderSubject, promptText, hasMedia, sendText, sendTyping)
	})
	return nil
}

// HandleWhatsAppInbound routes normalized GOWA WhatsApp webhook messages through the AI agent.
func (s *Service) HandleWhatsAppInbound(ctx context.Context, in WhatsAppInbound) error {
	if s.channelResolver == nil || in.FromMe {
		return nil
	}
	if ev := strings.TrimSpace(in.Event); ev != "" && !strings.EqualFold(ev, "message") {
		return nil
	}
	if isIgnoredWhatsAppAddress(in.ChatID, in.From, in.SenderID, in.Participant) {
		return nil
	}
	gowaClient := s.channelResolver.GowaClient()
	if gowaClient == nil {
		return nil
	}

	chatJID := firstNonEmpty(in.ChatID, in.From)
	isGroup := strings.HasSuffix(strings.ToLower(strings.TrimSpace(chatJID)), "@g.us")

	var replyTarget, senderRaw string
	if isGroup {
		replyTarget = strings.TrimSpace(chatJID)
		senderRaw = firstNonEmpty(in.SenderID, in.Participant, in.From)
	} else {
		e164, err := gowa.NormalizeJID(firstNonEmpty(in.From, in.SenderID, in.ChatID))
		if err != nil {
			return nil
		}
		replyTarget = e164
		senderRaw = e164
	}

	senderSubject, err := gowa.NormalizeJID(senderRaw)
	if err != nil {
		return nil
	}

	rawText := firstNonEmpty(in.Text, in.Caption)
	if isGroup && !isWhatsAppBotMentioned(gowaClient, in.MentionedJIDs, rawText) {
		return nil
	}
	promptText := stripWhatsAppMentions(rawText)

	if in.MessageID != "" && !s.dedup.firstSeen("wa:"+in.MessageID) {
		return nil
	}
	s.runAsync(func(ctx context.Context) error {
		sendText := func(txt string) { _ = gowaClient.SendMessage(ctx, replyTarget, txt) }
		sendTyping := func() { _ = gowaClient.SendChatPresence(ctx, replyTarget, "composing") }
		return s.handleChannelMessage(ctx, domain.ChannelWhatsApp, senderSubject, promptText, in.HasMedia, sendText, sendTyping)
	})
	return nil
}

func (s *Service) handleChannelMessage(
	ctx context.Context,
	ch domain.Channel,
	senderSubject string,
	promptText string,
	hasMedia bool,
	sendText func(string),
	sendTyping func(),
) error {
	if promptText == "" {
		if hasMedia {
			sendText("Maaf, saat ini Gabriel hanya dapat membaca pesan teks 🙏")
		}
		return nil
	}

	principal, err := s.channelResolver.ResolveByChannel(ctx, string(ch), senderSubject, uuid.Nil)
	if err != nil {
		var ex *errx.Error
		if errors.As(err, &ex) {
			switch ex.Code {
			case "USER_NOT_FOUND":
				sendText("Halo! Akun Anda belum terhubung dengan Domus. Silakan buka halaman Profil (/me) di Domus untuk menautkan akun ini.")
				return nil
			case "USER_DISABLED":
				sendText("⚠️ Akun Domus Anda sedang dinonaktifkan.")
				return nil
			}
		}
		return err
	}

	cmd, arg := parseSlashCommand(promptText)
	if cmd == "/parish" {
		return s.handleParishCommand(ctx, ch, senderSubject, principal, arg, sendText)
	}

	if principal.ParishID == uuid.Nil {
		sendText("Anda belum memiliki paroki aktif. Ketik /parish <kode_paroki> atau pilih paroki di aplikasi web Domus.")
		return nil
	}

	if cmd == "/new" || (ch == domain.ChannelTelegram && cmd == "/start") {
		if _, err := s.repo.CreateSession(ctx, domain.CreateSessionInput{
			ParishID: principal.ParishID,
			UserID:   principal.UserID,
			Channel:  ch,
			Title:    "Percakapan Baru",
		}); err != nil {
			return err
		}
		sendText("✅ Sesi percakapan baru telah dimulai. Silakan sampaikan pertanyaan atau kebutuhan Anda!")
		return nil
	}

	sendTyping()
	thinkingSent := false
	_, answer, err := s.RunTurn(ctx, ChatTurnInput{
		Principal: principal,
		Channel:   ch,
		Message:   promptText,
		OnStep: func(ev domain.AgentSSEEvent) error {
			switch ev.Event {
			case "thinking":
				if !thinkingSent {
					thinkingSent = true
					sendText("💭 Sedang menganalisis...")
					sendTyping()
				}
			case "tool_call":
				thinkingSent = false
				if ev.ToolName == "delegate_specialist" {
					sendText(formatDelegateProgress(ev.Arguments))
					sendTyping()
				} else if ev.ToolName != "" {
					sendText(fmt.Sprintf("⏳ Memanggil `%s`...", ev.ToolName))
					sendTyping()
				}
			}
			return nil
		},
	})
	if errors.Is(err, domain.ErrSessionBusy) {
		sendText("⏳ Mohon tunggu, Gabriel masih memproses pesan sebelumnya...")
		return nil
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(answer) != "" {
		sendText(answer)
	}
	return nil
}

func (s *Service) handleParishCommand(
	ctx context.Context,
	ch domain.Channel,
	senderSubject string,
	principal authx.Principal,
	code string,
	sendText func(string),
) error {
	if code == "" {
		sendText("Gunakan format: /parish <kode_paroki>")
		return nil
	}
	if s.parishes == nil {
		return errors.New("agent: parish reader is not configured")
	}
	parish, err := s.parishes.GetParishByCode(ctx, code)
	if err != nil {
		sendText("⚠️ Paroki dengan kode tersebut tidak ditemukan.")
		return nil
	}
	if err := s.channelResolver.SwitchChannelParish(ctx, string(ch), senderSubject, parish.ID); err != nil {
		sendText("⚠️ Anda belum terdaftar sebagai anggota paroki tersebut.")
		return nil
	}
	if _, err := s.channelResolver.ResolveByChannel(ctx, string(ch), senderSubject, parish.ID); err != nil {
		sendText("⚠️ Anda belum terdaftar sebagai anggota paroki tersebut.")
		return nil
	}
	_, _ = s.repo.FindOrCreateChannelSession(ctx, parish.ID, principal.UserID, ch, "Percakapan Baru")
	sendText(fmt.Sprintf("✅ Paroki aktif berhasil diubah ke %s (%s).", parish.Name, parish.Code))
	return nil
}

var tgMentionRes sync.Map // lowercased bot username -> *regexp.Regexp

func telegramMentionRe(botUsername string) *regexp.Regexp {
	key := strings.ToLower(botUsername)
	if v, ok := tgMentionRes.Load(key); ok {
		return v.(*regexp.Regexp)
	}
	re := regexp.MustCompile(`(?i)@` + regexp.QuoteMeta(botUsername) + `\b`)
	tgMentionRes.Store(key, re)
	return re
}

func hasTelegramMention(rawText string, entities, captionEntities []telegram.MessageEntity, botUsername string) bool {
	if botUsername == "" {
		return false
	}
	re := telegramMentionRe(botUsername)
	if re.MatchString(rawText) {
		return true
	}
	// Telegram entity offsets/lengths are UTF-16 code units.
	units := utf16.Encode([]rune(rawText))
	for _, list := range [][]telegram.MessageEntity{entities, captionEntities} {
		for _, e := range list {
			if e.Type != "mention" && e.Type != "bot_command" {
				continue
			}
			if e.Offset >= 0 && e.Length > 0 && e.Offset+e.Length <= len(units) {
				if re.MatchString(string(utf16.Decode(units[e.Offset : e.Offset+e.Length]))) {
					return true
				}
			}
		}
	}
	return false
}

func stripTelegramBotMention(rawText, botUsername string) string {
	if botUsername == "" {
		return strings.TrimSpace(rawText)
	}
	cleaned := telegramMentionRe(botUsername).ReplaceAllString(rawText, "")
	return strings.TrimSpace(multiSpaceRe.ReplaceAllString(cleaned, " "))
}

// isWhatsAppBotMentioned is true for a literal "@gabriel", or (when app bot number is
// configured) a mention JID / "@<digits>" that matches the bot's own number.
func isWhatsAppBotMentioned(c *gowa.Client, jids []string, text string) bool {
	if waGabrielRe.MatchString(text) {
		return true
	}
	for _, j := range jids {
		if c.IsBotMention(j) {
			return true
		}
	}
	for _, m := range waDigitsMentionRe.FindAllString(text, -1) {
		if c.IsBotMention(m) {
			return true
		}
	}
	return false
}

func stripWhatsAppMentions(rawText string) string {
	cleaned := waGabrielRe.ReplaceAllString(rawText, "")
	cleaned = waDigitsMentionRe.ReplaceAllString(cleaned, "")
	return strings.TrimSpace(multiSpaceRe.ReplaceAllString(cleaned, " "))
}

func isIgnoredWhatsAppAddress(addrs ...string) bool {
	for _, a := range addrs {
		lower := strings.ToLower(a)
		if strings.Contains(lower, "status@broadcast") || strings.Contains(lower, "@broadcast") || strings.Contains(lower, "@newsletter") {
			return true
		}
	}
	return false
}

func parseSlashCommand(text string) (cmd, arg string) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "/") {
		return "", ""
	}
	cmd = strings.ToLower(fields[0])
	if len(fields) > 1 {
		arg = strings.TrimSpace(fields[1])
	}
	return cmd, arg
}

func formatDelegateProgress(rawArgs string) string {
	var args struct {
		AgentName  string `json:"agent_name"`
		Specialist string `json:"specialist"`
		Agent      string `json:"agent"`
	}
	_ = json.Unmarshal([]byte(rawArgs), &args)
	target := firstNonEmpty(args.AgentName, args.Specialist, args.Agent)
	if target != "" {
		return fmt.Sprintf("⏳ Menghubungi spesialis %s...", strings.ToUpper(target[:1])+target[1:])
	}
	return "⏳ Menghubungi spesialis..."
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
