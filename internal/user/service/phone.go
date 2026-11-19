package service

import (
	"regexp"
	"strings"

	"github.com/paroki/domus/internal/user/domain"
)

var e164Re = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

// normalizePhone normalizes a phone number to canonical E.164 (defaulting local ID numbers to +62).
func normalizePhone(raw string) (string, error) {
	cleaned := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" {
		return "", nil
	}
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "+62" + cleaned[1:]
	} else if strings.HasPrefix(cleaned, "62") && !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}
	if !e164Re.MatchString(cleaned) {
		return "", domain.ValidationError("phone", "INVALID_PHONE", "Phone number must be a valid E.164 number.")
	}
	return cleaned, nil
}
