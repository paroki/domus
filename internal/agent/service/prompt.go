package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
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
	Channel        domain.Channel
	Now            time.Time
	Parish         *ParishInfo
	Roles          []string
	Orgs           []OrgInfo
	User           *UserProfileInfo
	PinnedMemories []domain.Memory
}

// BuildSystemPrompt assembles base persona instruction + XML context, profile, and pinned memory blocks.
func BuildSystemPrompt(in PromptInput) string {
	now := in.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	var b strings.Builder
	b.WriteString(strings.TrimSpace(in.Instruction))
	b.WriteString("\n\n<agent_context>\n")
	fmt.Fprintf(&b, "  <channel>%s</channel>\n", promptXMLEscaper.Replace(string(in.Channel)))
	fmt.Fprintf(&b, "  <current_time>%s</current_time>\n", now.Format(time.RFC3339))
	if in.Parish != nil {
		fmt.Fprintf(
			&b,
			"  <parish id=\"%s\" code=\"%s\">%s</parish>\n",
			promptXMLEscaper.Replace(in.Parish.ID.String()),
			promptXMLEscaper.Replace(in.Parish.Code),
			promptXMLEscaper.Replace(in.Parish.Name),
		)
	}
	b.WriteString("  <user_roles>\n")
	for _, role := range in.Roles {
		fmt.Fprintf(&b, "    <role>%s</role>\n", promptXMLEscaper.Replace(role))
	}
	b.WriteString("  </user_roles>\n")
	b.WriteString("  <user_organizations>\n")
	for _, org := range in.Orgs {
		fmt.Fprintf(
			&b,
			"    <organization id=\"%s\" code=\"%s\" kind=\"%s\" role=\"%s\">%s</organization>\n",
			promptXMLEscaper.Replace(org.ID.String()),
			promptXMLEscaper.Replace(org.Code),
			promptXMLEscaper.Replace(org.Kind),
			promptXMLEscaper.Replace(org.Role),
			promptXMLEscaper.Replace(org.Name),
		)
	}
	b.WriteString("  </user_organizations>\n")
	b.WriteString("</agent_context>\n\n")

	userName := "Umat"
	var genderPtr, birthDatePtr *string
	if in.User != nil {
		if strings.TrimSpace(in.User.Name) != "" {
			userName = in.User.Name
		}
		genderPtr = in.User.Gender
		birthDatePtr = in.User.BirthDate
	}
	age, defaultSal := ComputeAgeAndSalutation(userName, genderPtr, birthDatePtr, now)
	genderStr := "unset"
	if genderPtr != nil && strings.TrimSpace(*genderPtr) != "" {
		genderStr = strings.TrimSpace(*genderPtr)
	}
	birthDateStr := "unset"
	if birthDatePtr != nil && strings.TrimSpace(*birthDatePtr) != "" {
		birthDateStr = strings.TrimSpace(*birthDatePtr)
	}
	ageStr := "unset"
	if age != nil {
		ageStr = strconv.Itoa(*age)
	}

	b.WriteString("<user_profile>\n")
	fmt.Fprintf(&b, "  <name>%s</name>\n", promptXMLEscaper.Replace(userName))
	fmt.Fprintf(&b, "  <gender>%s</gender>\n", promptXMLEscaper.Replace(genderStr))
	fmt.Fprintf(&b, "  <birth_date>%s</birth_date>\n", promptXMLEscaper.Replace(birthDateStr))
	fmt.Fprintf(&b, "  <age>%s</age>\n", ageStr)
	fmt.Fprintf(&b, "  <default_salutation>%s</default_salutation>\n", promptXMLEscaper.Replace(defaultSal))
	b.WriteString("</user_profile>")

	if memXML := domain.FormatUserMemoryXML(in.PinnedMemories); memXML != "" {
		b.WriteString("\n\n")
		b.WriteString(memXML)
	}
	return b.String()
}
