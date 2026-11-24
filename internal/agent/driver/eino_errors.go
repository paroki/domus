package driver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/paroki/domus/internal/agent/domain"
	"google.golang.org/genai"
)

// WrapProviderError maps upstream LLM SDK/HTTP errors to 4xx domain sentinels
// (429 for rate limit, 400 for auth/request/overloaded/unavailable) and redacts apiKey.
func WrapProviderError(err error, apiKey string) error {
	if err == nil {
		return nil
	}
	msg := redactAPIKey(err.Error(), apiKey)

	if status, ok := extractHTTPStatus(err); ok {
		return fmt.Errorf("%w: provider HTTP %d: %s", providerStatusSentinel(status), status, msg)
	}

	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "429") || strings.Contains(lower, "rate limit") || strings.Contains(lower, "quota"):
		return fmt.Errorf("%w: %s", domain.ErrProviderRateLimited, msg)
	case strings.Contains(lower, "401") || strings.Contains(lower, "403") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "invalid api key") || strings.Contains(lower, "authentication"):
		return fmt.Errorf("%w: %s", domain.ErrProviderAuthFailed, msg)
	case strings.Contains(lower, "400") || strings.Contains(lower, "422") || strings.Contains(lower, "bad request"):
		return fmt.Errorf("%w: %s", domain.ErrProviderRequestInvalid, msg)
	case strings.Contains(lower, "500") || strings.Contains(lower, "502") || strings.Contains(lower, "503") || strings.Contains(lower, "504") || strings.Contains(lower, "overloaded"):
		return fmt.Errorf("%w: %s", domain.ErrProviderOverloaded, msg)
	default:
		return fmt.Errorf("%w: %s", domain.ErrProviderUnavailable, msg)
	}
}

func extractHTTPStatus(err error) (int, bool) {
	var genaiErr *genai.APIError
	if errors.As(err, &genaiErr) && genaiErr.Code >= http.StatusBadRequest {
		return genaiErr.Code, true
	}
	var antErr *anthropic.Error
	if errors.As(err, &antErr) && antErr.StatusCode >= http.StatusBadRequest {
		return antErr.StatusCode, true
	}
	return 0, false
}
