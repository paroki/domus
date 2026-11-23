package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/prompts"
)

// UserProfileInfo holds cross-feature user biodata for prompt injection.
type UserProfileInfo struct {
	ID        uuid.UUID
	Name      string
	Gender    *string
	BirthDate *string
}

// ParishInfo holds cross-feature parish metadata for prompt injection.
type ParishInfo struct {
	ID   uuid.UUID
	Code string
	Name string
}

// OrgInfo holds cross-feature organization membership metadata for prompt injection.
type OrgInfo struct {
	ID   uuid.UUID
	Code string
	Name string
	Kind string
	Role string
}

// UserReader reads user biodata without coupling to internal/user.
type UserReader interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*UserProfileInfo, error)
}

// ParishReader reads parish metadata without coupling to internal/ref.
type ParishReader interface {
	GetParish(ctx context.Context, parishID uuid.UUID) (*ParishInfo, error)
	GetParishByCode(ctx context.Context, code string) (*ParishInfo, error)
}

// OrgReader reads a user's organization memberships without coupling to internal/org.
type OrgReader interface {
	ListUserOrgs(ctx context.Context, parishID, userID uuid.UUID) ([]OrgInfo, error)
}

var promptXMLEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
)

// ComputeAgeAndSalutation computes UTC age from birthDate and the Indonesian default salutation.
func ComputeAgeAndSalutation(name string, gender, birthDate *string, now time.Time) (age *int, defaultSalutation string) {
	firstName := "Umat"
	if fields := strings.Fields(strings.TrimSpace(name)); len(fields) > 0 {
		firstName = fields[0]
	}

	if birthDate != nil && strings.TrimSpace(*birthDate) != "" {
		if dob, err := time.Parse("2006-01-02", strings.TrimSpace(*birthDate)); err == nil {
			utcNow := now.UTC()
			years := utcNow.Year() - dob.Year()
			if utcNow.Month() < dob.Month() || (utcNow.Month() == dob.Month() && utcNow.Day() < dob.Day()) {
				years--
			}
			if years >= 0 {
				age = &years
			}
		}
	}

	g := ""
	if gender != nil {
		g = strings.TrimSpace(*gender)
	}
	if age != nil && *age >= 30 {
		switch g {
		case "pria":
			return age, "Bapak " + firstName
		case "wanita":
			return age, "Ibu " + firstName
		}
	}
	return age, "Kak " + firstName
}

// PromptInput holds all context blocks injected into an agent's system prompt.
type PromptInput struct {
	Instruction    string
	IsOrchestrator bool
	Channel        domain.Channel
	Now            time.Time
	Parish         *ParishInfo
	Roles          []string
	Orgs           []OrgInfo
	User           *UserProfileInfo
	PinnedMemories []domain.Memory
}

// BuildSystemPrompt assembles base persona instruction + XML context, pre-computed salutation,
// conditional onboarding, and pinned memory blocks via Genkit Dotprompt.
func BuildSystemPrompt(in PromptInput) string {
	now := in.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	var parishXML string
	if in.Parish != nil {
		parishXML = fmt.Sprintf(
			`<parish id="%s" code="%s">%s</parish>`,
			promptXMLEscaper.Replace(in.Parish.ID.String()),
			promptXMLEscaper.Replace(in.Parish.Code),
			promptXMLEscaper.Replace(in.Parish.Name),
		)
	}

	var rolesB strings.Builder
	for _, role := range in.Roles {
		fmt.Fprintf(&rolesB, "    <role>%s</role>\n", promptXMLEscaper.Replace(role))
	}

	var orgsB strings.Builder
	for _, org := range in.Orgs {
		fmt.Fprintf(
			&orgsB,
			"    <organization id=\"%s\" code=\"%s\" kind=\"%s\" role=\"%s\">%s</organization>\n",
			promptXMLEscaper.Replace(org.ID.String()),
			promptXMLEscaper.Replace(org.Code),
			promptXMLEscaper.Replace(org.Kind),
			promptXMLEscaper.Replace(org.Role),
			promptXMLEscaper.Replace(org.Name),
		)
	}

	userName := "Umat"
	var genderPtr, birthDatePtr *string
	if in.User != nil {
		if strings.TrimSpace(in.User.Name) != "" {
			userName = in.User.Name
		}
		genderPtr = in.User.Gender
		birthDatePtr = in.User.BirthDate
	}
	_, defaultSal := ComputeAgeAndSalutation(userName, genderPtr, birthDatePtr, now)

	hasProfileMem, hasPrefMem := false, false
	for _, m := range in.PinnedMemories {
		if !m.IsPinned {
			continue
		}
		switch m.Category {
		case domain.MemoryCategoryProfile:
			hasProfileMem = true
		case domain.MemoryCategoryPreference:
			hasPrefMem = true
		}
	}
	needsOnboarding := in.IsOrchestrator && (!hasProfileMem || !hasPrefMem)

	var a2ui string
	if in.IsOrchestrator && in.Channel == domain.ChannelWeb {
		a2ui = A2UIWebPromptInstructions()
	}

	rendered, err := prompts.RenderSystemPrompt(context.Background(), prompts.SystemPromptData{
		Instruction:      strings.TrimSpace(in.Instruction),
		Channel:          promptXMLEscaper.Replace(string(in.Channel)),
		CurrentTime:      now.Format(time.RFC3339),
		ParishXML:        parishXML,
		RolesXML:         strings.TrimSuffix(rolesB.String(), "\n"),
		OrgsXML:          strings.TrimSuffix(orgsB.String(), "\n"),
		UserName:         promptXMLEscaper.Replace(userName),
		Salutation:       promptXMLEscaper.Replace(defaultSal),
		NeedsOnboarding:  needsOnboarding,
		UserMemoryXML:    domain.FormatUserMemoryXML(in.PinnedMemories),
		A2UIInstructions: a2ui,
	})
	if err == nil {
		return rendered
	}
	return strings.TrimSpace(in.Instruction)
}
