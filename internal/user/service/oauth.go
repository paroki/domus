package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// EnabledProviders returns the configured OAuth2 provider names in deterministic order.
func (s *Service) EnabledProviders() []string {
	ordered := make([]string, 0, len(s.oauthProviders))
	seen := make(map[string]bool, len(s.oauthProviders))
	for _, name := range []string{domain.ProviderGoogle, domain.ProviderGitHub} {
		if _, ok := s.oauthProviders[name]; ok {
			ordered = append(ordered, name)
			seen[name] = true
		}
	}
	var extras []string
	for name := range s.oauthProviders {
		if !seen[name] {
			extras = append(extras, name)
		}
	}
	sort.Strings(extras)
	return append(ordered, extras...)
}

// OAuthAuthURL returns the authorization redirect URL for providerName.
func (s *Service) OAuthAuthURL(providerName, state, redirectURI string) (string, error) {
	p, ok := s.oauthProviders[providerName]
	if !ok {
		return "", domain.ErrOAuthProviderUnknown
	}
	return p.AuthCodeURL(state, redirectURI), nil
}

// OAuthCallback exchanges the authorization code, verifies (or invite-onboards) the user, and creates a session.
func (s *Service) OAuthCallback(
	ctx context.Context,
	providerName, code, redirectURI, inviteCode, userAgent, ipAddress string,
) (string, *domain.AuthSessionResponse, error) {
	p, ok := s.oauthProviders[providerName]
	if !ok {
		return "", nil, domain.ErrOAuthProviderUnknown
	}

	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode != "" && s.inviteJoiner != nil {
		if err := s.inviteJoiner.ValidateInviteCode(ctx, inviteCode); err != nil {
			return "", nil, err
		}
	}

	info, err := p.Exchange(ctx, code, redirectURI)
	if err != nil {
		return "", nil, err
	}
	email := normalizeEmail(info.Email)
	if !info.EmailVerified || email == "" {
		return "", nil, domain.ErrOAuthEmailUnverified
	}

	now := time.Now().UTC()
	var rawToken string
	var sessionID uuid.UUID
	var resp *domain.AuthSessionResponse
	var updatedUser *ent.User

	var preEnsuredUserID uuid.UUID
	if inviteCode != "" && s.inviteJoiner != nil {
		acc, accErr := s.repo.FindAccountByProviderSubject(ctx, providerName, info.Subject)
		switch {
		case accErr == nil:
			preEnsuredUserID = acc.UserID
		case ent.IsNotFound(accErr):
			existing, findErr := s.repo.FindUserByEmail(ctx, email)
			switch {
			case findErr == nil:
				preEnsuredUserID = existing.ID
			case ent.IsNotFound(findErr):
				ensured, ensureErr := s.EnsureUser(ctx, email, info.Name, "", providerName, info.Subject)
				if ensureErr != nil {
					return "", nil, ensureErr
				}
				preEnsuredUserID = ensured.ID
			default:
				return "", nil, findErr
			}
		default:
			return "", nil, accErr
		}
	}

	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		var u *ent.User
		if preEnsuredUserID != uuid.Nil {
			u, err = guardActiveUser(txRepo.FindUserByID(ctx, preEnsuredUserID))
		} else {
			acc, accErr := txRepo.FindAccountByProviderSubject(ctx, providerName, info.Subject)
			if accErr == nil {
				u, err = guardActiveUser(txRepo.FindUserByID(ctx, acc.UserID))
			} else if ent.IsNotFound(accErr) {
				u, err = guardActiveUser(txRepo.FindUserByEmail(ctx, email))
			} else {
				return accErr
			}
		}
		if err != nil {
			return err
		}

		var displayName *string
		if info.Name != "" {
			displayName = &info.Name
		}
		var avatarURL *string
		if info.AvatarURL != "" {
			avatarURL = &info.AvatarURL
		}

		if _, err := txRepo.UpsertAccount(ctx, u.ID, providerName, info.Subject, &email, displayName, avatarURL, now); err != nil {
			return err
		}
		updatedUser, err = txRepo.UpdateUserLastLogin(ctx, u.ID, now)
		if err != nil {
			return err
		}
		if avatarURL != nil && (updatedUser.AvatarURL == nil || *updatedUser.AvatarURL == "") {
			updatedUser, err = txRepo.UpdateUser(ctx, u.ID, nil, nil, nil, nil, avatarURL, nil, nil, nil, nil, nil, nil, nil, nil)
			if err != nil {
				return err
			}
		}
		rawToken, sessionID, resp, err = s.createSessionWithID(ctx, txRepo, updatedUser, userAgent, ipAddress)
		return err
	})
	if err != nil {
		return "", nil, err
	}

	if inviteCode != "" && s.inviteJoiner != nil {
		parokiID, err := s.inviteJoiner.RedeemInviteForUser(ctx, inviteCode, updatedUser.ID, sessionID)
		if err != nil {
			return "", nil, err
		}
		reloaded, err := s.repo.FindUserByID(ctx, updatedUser.ID)
		if err == nil {
			updatedUser = reloaded
		}
		resp, err = s.buildAuthSessionResponse(ctx, s.repo, updatedUser, &parokiID, resp.ExpiresAt)
		if err != nil {
			return "", nil, err
		}
	}
	return rawToken, resp, nil
}

type oauth2Driver struct {
	name     string
	cfg      *oauth2.Config
	userInfo func(ctx context.Context, client *http.Client) (domain.OAuthUserInfo, error)
}

func (d *oauth2Driver) Name() string { return d.name }

func (d *oauth2Driver) AuthCodeURL(state, redirectURI string) string {
	return d.cfg.AuthCodeURL(state, oauth2.SetAuthURLParam("redirect_uri", redirectURI))
}

func (d *oauth2Driver) Exchange(ctx context.Context, code, redirectURI string) (domain.OAuthUserInfo, error) {
	tok, err := d.cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("redirect_uri", redirectURI))
	if err != nil {
		return domain.OAuthUserInfo{}, err
	}
	return d.userInfo(ctx, d.cfg.Client(ctx, tok))
}

func newGoogleProvider(clientID, clientSecret string) domain.OAuthProvider {
	return &oauth2Driver{
		name: domain.ProviderGoogle,
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		userInfo: fetchGoogleUserInfo,
	}
}

func newGitHubProvider(clientID, clientSecret string) domain.OAuthProvider {
	return &oauth2Driver{
		name: domain.ProviderGitHub,
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		},
		userInfo: fetchGitHubUserInfo,
	}
}

func fetchGoogleUserInfo(ctx context.Context, client *http.Client) (domain.OAuthUserInfo, error) {
	var payload struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := getJSON(ctx, client, "https://www.googleapis.com/oauth2/v3/userinfo", &payload); err != nil {
		return domain.OAuthUserInfo{}, err
	}
	return domain.OAuthUserInfo{
		Subject:       payload.Sub,
		Email:         payload.Email,
		EmailVerified: payload.EmailVerified,
		Name:          payload.Name,
		AvatarURL:     payload.Picture,
	}, nil
}

func fetchGitHubUserInfo(ctx context.Context, client *http.Client) (domain.OAuthUserInfo, error) {
	var profile struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	if err := getJSON(ctx, client, "https://api.github.com/user", &profile); err != nil {
		return domain.OAuthUserInfo{}, err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getJSON(ctx, client, "https://api.github.com/user/emails", &emails); err != nil {
		return domain.OAuthUserInfo{}, err
	}

	var resolvedEmail string
	var verified bool
	for _, e := range emails {
		if e.Primary && e.Verified {
			resolvedEmail = e.Email
			verified = true
			break
		}
		if e.Verified && resolvedEmail == "" {
			resolvedEmail = e.Email
			verified = true
		}
	}

	name := profile.Name
	if name == "" {
		name = profile.Login
	}
	return domain.OAuthUserInfo{
		Subject:       strconv.FormatInt(profile.ID, 10),
		Email:         resolvedEmail,
		EmailVerified: verified,
		Name:          name,
		AvatarURL:     profile.AvatarURL,
	}, nil
}

func getJSON(ctx context.Context, client *http.Client, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("oauth userinfo %s returned %d: %s", url, resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
