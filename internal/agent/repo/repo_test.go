package repo

import (
	"context"
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	client, db, d := testutil.NewEntClientWithDB(t)
	return New(client, db, d)
}

func TestSessionCRUD_ChannelResumeAndTenantIsolation(t *testing.T) {
	ctx := context.Background()
	r := newTestRepo(t)

	parishA := uuid.New()
	parishB := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	s1, err := r.CreateSession(ctx, domain.CreateSessionInput{
		ParishID: parishA,
		UserID:   userA,
		Channel:  domain.ChannelWeb,
		Title:    "Jadwal Misa",
	})
	require.NoError(t, err)
	assert.Equal(t, "Jadwal Misa", s1.Title)

	// Cross-user and cross-parish isolation
	_, err = r.FindSessionByID(ctx, parishA, userB, s1.ID)
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	_, err = r.FindSessionByID(ctx, parishB, userA, s1.ID)
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)

	// Rename session
	renamed, err := r.UpdateSessionTitle(ctx, parishA, userA, s1.ID, "Jadwal Misa Mingguan")
	require.NoError(t, err)
	assert.Equal(t, "Jadwal Misa Mingguan", renamed.Title)

	// Channel auto-resume creates once and resumes same session
	wa1, err := r.FindOrCreateChannelSession(ctx, parishA, userA, domain.ChannelWhatsApp, "WhatsApp")
	require.NoError(t, err)
	wa2, err := r.FindOrCreateChannelSession(ctx, parishA, userA, domain.ChannelWhatsApp, "WhatsApp")
	require.NoError(t, err)
	assert.Equal(t, wa1.ID, wa2.ID)

	// Search sessions with channel filter
	webList, err := r.SearchSessions(ctx, domain.SearchSessionsInput{
		ParishID: parishA,
		UserID:   userA,
		Channel:  string(domain.ChannelWeb),
		Limit:    10,
	})
	require.NoError(t, err)
	require.Len(t, webList, 1)
	assert.Equal(t, s1.ID, webList[0].ID)

	// Soft-delete hides session
	require.NoError(t, r.SoftDeleteSession(ctx, parishA, userA, s1.ID))
	_, err = r.FindSessionByID(ctx, parishA, userA, s1.ID)
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
}

func TestMessages_Tier1RecentAndTier2EpisodicSearch(t *testing.T) {
	ctx := context.Background()
	r := newTestRepo(t)

	parishID := uuid.New()
	userID := uuid.New()

	webSess, err := r.CreateSession(ctx, domain.CreateSessionInput{
		ParishID: parishID,
		UserID:   userID,
		Channel:  domain.ChannelWeb,
		Title:    "Web Chat",
	})
	require.NoError(t, err)

	tgSess, err := r.CreateSession(ctx, domain.CreateSessionInput{
		ParishID: parishID,
		UserID:   userID,
		Channel:  domain.ChannelTelegram,
		Title:    "Telegram Chat",
	})
	require.NoError(t, err)

	m1, err := r.AppendMessage(ctx, domain.CreateMessageInput{
		ParishID:  parishID,
		UserID:    userID,
		SessionID: webSess.ID,
		Role:      domain.RoleUser,
		Content:   "Kapan jadwal baptis anak bulan depan?",
	})
	require.NoError(t, err)

	m2, err := r.AppendMessage(ctx, domain.CreateMessageInput{
		ParishID:  parishID,
		UserID:    userID,
		SessionID: webSess.ID,
		Role:      domain.RoleAssistant,
		Content:   "Baptis anak diadakan setiap Minggu pertama.",
		ToolCalls: []domain.ToolCall{{ID: "tc1", Name: "event_search", Arguments: "{}"}},
	})
	require.NoError(t, err)

	m3, err := r.AppendMessage(ctx, domain.CreateMessageInput{
		ParishID:  parishID,
		UserID:    userID,
		SessionID: tgSess.ID,
		Role:      domain.RoleUser,
		Content:   "Syarat pendaftaran baptis apa saja?",
	})
	require.NoError(t, err)
	_ = m3

	// Tier 1: chronological ASC order
	recent, err := r.ListRecentMessages(ctx, parishID, webSess.ID, 10)
	require.NoError(t, err)
	require.Len(t, recent, 2)
	assert.Equal(t, m1.ID, recent[0].ID)
	assert.Equal(t, m2.ID, recent[1].ID)

	// Tier 2: session search
	sessHits, err := r.SearchSessionMessages(ctx, domain.SearchSessionMessagesInput{
		ParishID:  parishID,
		UserID:    userID,
		SessionID: webSess.ID,
		Q:         "baptis",
		Limit:     10,
	})
	require.NoError(t, err)
	assert.Len(t, sessHits, 2)

	// Tier 2: cross-channel history search
	allHits, err := r.SearchUserHistory(ctx, domain.SearchUserHistoryInput{
		ParishID: parishID,
		UserID:   userID,
		Q:        "baptis",
		Limit:    10,
	})
	require.NoError(t, err)
	assert.Len(t, allHits, 3)

	// Soft-deleting tgSess excludes its messages from SearchUserHistory
	require.NoError(t, r.SoftDeleteSession(ctx, parishID, userID, tgSess.ID))
	afterDeleteHits, err := r.SearchUserHistory(ctx, domain.SearchUserHistoryInput{
		ParishID: parishID,
		UserID:   userID,
		Q:        "baptis",
		Limit:    10,
	})
	require.NoError(t, err)
	assert.Len(t, afterDeleteHits, 2)
}

func TestMemory_Tier1Pinned_Tier3QuotasSupersedeAndRecall(t *testing.T) {
	ctx := context.Background()
	r := newTestRepo(t)

	parishID := uuid.New()
	otherParish := uuid.New()
	userID := uuid.New()
	otherUser := uuid.New()

	// FormatHalfvec validation
	_, err := formatHalfvec([]float32{0.1, 0.2})
	require.Error(t, err)
	vec := make([]float32, domain.EmbeddingDimensions)
	vec[0] = 0.5
	vecStr, err := formatHalfvec(vec)
	require.NoError(t, err)
	assert.Contains(t, vecStr, "[0.5,")

	// Fill 5 pinned memories
	var firstPinnedID uuid.UUID
	for i := range domain.MaxPinnedMemories {
		mem, saveErr := r.SaveMemory(ctx, domain.SaveMemoryInput{
			ParishID:  parishID,
			UserID:    &userID,
			Scope:     domain.MemoryScopeUser,
			Category:  domain.MemoryCategoryPreference,
			Content:   fmt.Sprintf("Pinned fact %d", i),
			IsPinned:  true,
			Embedding: vec,
		})
		require.NoError(t, saveErr)
		if i == 0 {
			firstPinnedID = mem.ID
		}
	}

	pinned, err := r.ListPinnedMemories(ctx, parishID, userID, 0)
	require.NoError(t, err)
	assert.Len(t, pinned, domain.MaxPinnedMemories)

	// 6th pinned memory fails with AGENT_MEMORY_LIMIT_EXCEEDED
	_, err = r.SaveMemory(ctx, domain.SaveMemoryInput{
		ParishID: parishID,
		UserID:   &userID,
		Scope:    domain.MemoryScopeUser,
		Category: domain.MemoryCategoryPreference,
		Content:  "6th pinned fact",
		IsPinned: true,
	})
	assert.ErrorIs(t, err, domain.ErrMemoryLimitExceeded)

	// Superseding an existing pinned memory succeeds even at quota!
	superseded, err := r.SaveMemory(ctx, domain.SaveMemoryInput{
		ParishID:    parishID,
		UserID:      &userID,
		Scope:       domain.MemoryScopeUser,
		Category:    domain.MemoryCategoryMinistry,
		Content:     "Koordinator Lektor Wilayah 3",
		IsPinned:    true,
		SupersedeID: &firstPinnedID,
		Embedding:   vec,
	})
	require.NoError(t, err)
	assert.NotEqual(t, firstPinnedID, superseded.ID)

	// Parish FAQ memory is visible to user recall
	_, err = r.SaveMemory(ctx, domain.SaveMemoryInput{
		ParishID: parishID,
		Scope:    domain.MemoryScopeParish,
		Category: domain.MemoryCategoryParishFAQ,
		Content:  "Sekretariat paroki buka Selasa-Minggu pukul 08:00-15:00.",
	})
	require.NoError(t, err)

	// Other user's private memory must NOT leak
	otherUserMem, err := r.SaveMemory(ctx, domain.SaveMemoryInput{
		ParishID: parishID,
		UserID:   &otherUser,
		Scope:    domain.MemoryScopeUser,
		Category: domain.MemoryCategoryMinistry,
		Content:  "Rahasia Lektor user lain",
	})
	require.NoError(t, err)

	// Other parish's FAQ must NOT leak
	otherParishMem, err := r.SaveMemory(ctx, domain.SaveMemoryInput{
		ParishID: otherParish,
		Scope:    domain.MemoryScopeParish,
		Category: domain.MemoryCategoryParishFAQ,
		Content:  "Sekretariat Lektor paroki lain",
	})
	require.NoError(t, err)

	// Recall with query (hybrid on Postgres, ContainsFold on SQLite)
	var recallEmbedding []float32
	if r.dialect == dialect.Postgres {
		recallEmbedding = vec
	}
	recalled, err := r.RecallMemories(ctx, domain.RecallMemoriesInput{
		ParishID:       parishID,
		UserID:         userID,
		Query:          "Lektor",
		QueryEmbedding: recallEmbedding,
		Limit:          10,
	})
	require.NoError(t, err)
	if r.dialect == dialect.Postgres {
		require.NotEmpty(t, recalled)
		assert.Equal(t, superseded.ID, recalled[0].ID)
		assert.Greater(t, recalled[0].Score, 0.0)
		for _, m := range recalled {
			assert.Equal(t, parishID, m.ParishID)
			if m.Scope == domain.MemoryScopeUser {
				require.NotNil(t, m.UserID)
				assert.Equal(t, userID, *m.UserID)
			}
			assert.NotEqual(t, firstPinnedID, m.ID)
			assert.NotEqual(t, otherUserMem.ID, m.ID)
			assert.NotEqual(t, otherParishMem.ID, m.ID)
		}

		bm25Only, bm25Err := r.RecallMemories(ctx, domain.RecallMemoriesInput{
			ParishID: parishID,
			UserID:   userID,
			Query:    "Lektor",
			Limit:    10,
		})
		require.NoError(t, bm25Err)
		require.Len(t, bm25Only, 1)
		assert.Equal(t, superseded.ID, bm25Only[0].ID)
	} else {
		require.Len(t, recalled, 1)
		assert.Equal(t, superseded.ID, recalled[0].ID)
	}

	// SearchMemories returns user + parish memories for drawer
	drawer, err := r.SearchMemories(ctx, domain.SearchMemoriesInput{
		ParishID: parishID,
		UserID:   userID,
		Limit:    20,
	})
	require.NoError(t, err)
	assert.Len(t, drawer, 6)

	// Soft-delete memory
	require.NoError(t, r.SoftDeleteMemory(ctx, parishID, userID, superseded.ID))
	drawerAfter, err := r.SearchMemories(ctx, domain.SearchMemoriesInput{
		ParishID: parishID,
		UserID:   userID,
		Limit:    20,
	})
	require.NoError(t, err)
	assert.Len(t, drawerAfter, 5)
}
