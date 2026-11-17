# Authentication and User Management: Product Requirements Document

| Field    | Value                                                                                                                                  |
| :------- | :------------------------------------------------------------------------------------------------------------------------------------- |
| Document | `docs/auth.md`                                                                                                                         |
| Version  | 1.0.1                                                                                                                                  |
| Status   | Draft                                                                                                                                  |
| Date     | 2026-11-17                                                                                                                             |
| Features | `auth` and `user` (backend `internal/auth`, `internal/user`; UI `ui/app/features/auth`, `ui/app/features/user`)                        |
| Audience | Product, engineering, and AI implementation agents                                                                                     |

## Changelog

| Version | Date       | Notes         |
| :------ | :--------- | :------------ |
| 1.0.0   | 2026-11-17 | Initial draft |
| 1.0.1   | 2026-11-17 | Current user moves to `user` (D-11); OTP-6 covers invite requests; CLI skips channel validation; 90-day session cap; depguard boundaries; trimmed section 17 |

---

## 1. Overview

### 1.1 Problem Statement

Domus digitalizes parish administration and management. It currently has no concept of identity: every route is open and there is no user record. Before any administrative feature (parishioners, sacraments, events, organizations) can ship, Domus must know who is using it, restrict access to people the parish has explicitly invited, and make signing in simple for parishioners who are mostly non-technical and reach Domus from a WhatsApp group.

### 1.2 Goals

1. Provide **invite-only** access. A person who has no invite and no existing account cannot register or sign in.
2. Provide two sign-in methods: **OTP** (delivered by email and WhatsApp) and **OAuth2** (Google and GitHub in v1).
3. Make OAuth providers **drivers**: adding a provider requires one new file, one registry line, and environment variables, with no change to services, handlers, or UI logic.
4. Make OTP delivery channels drivers as well (email and WhatsApp in v1).
5. Introduce the `user` feature: the user record, role, status, and basic user administration.
6. Establish the environment variable conventions that all future features follow.

### 1.3 Non-Goals

- Password-based authentication, and multi-factor authentication on top of OTP or OAuth.
- Open or self-service registration.
- Facebook or any OAuth provider other than Google and GitHub in v1.
- Full role-based access control. Only a minimal `admin` / `member` role exists (see 6.1).
- Linking a `user` to a parishioner record (a later feature).
- Persisted audit trail (structured logs only, see SEC-12).
- Admin UI screens for invites, users, and login-method management (the API and CLI exist in v1, see section 10).
- Renaming the existing unprefixed environment variables (`HOST`, `PORT`, and so on) to `APP_*`. Tracked as a separate issue.
- Database and Ent bootstrap. Tracked as a separate prerequisite issue (see section 14).

### 1.4 Users

| Persona              | Description                                                                              | Primary need                                                |
| :------------------- | :--------------------------------------------------------------------------------------- | :---------------------------------------------------------- |
| Parishioner (member) | Parish member invited through a WhatsApp group. Often non-technical and mobile-first.    | Join with minimum steps and sign in later without friction. |
| Parish administrator | Staff or volunteer who manages Domus. Role `admin`.                                      | Invite people safely and manage who has access.             |
| Operator             | Person who deploys and configures Domus.                                                 | Bootstrap the first admin and configure providers via env.  |

### 1.5 Success Metrics

| Metric                                                                | Target (initial)           |
| :-------------------------------------------------------------------- | :------------------------- |
| Invited users who complete onboarding                                 | 80% or more                |
| Median time from opening an invite link to being signed in            | Under 2 minutes            |
| OTP delivery success (at least one channel delivered)                 | 98% or more                |
| Unauthorized registrations (accounts created without a valid invite)  | 0                          |

---

## 2. Decisions

| ID   | Decision                                                                                                                                                  |
| :--- | :-------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D-01 | Access is invite-only. A generic, multi-use invite link is distributed through a WhatsApp group. Links carry an expiry, a use cap, and a preset role.      |
| D-02 | Registered users sign in at `/login` without an invite. The invite is only needed to create a new account.                                                |
| D-03 | New users are active immediately after onboarding. There is no admin approval step.                                                                       |
| D-04 | Sign-in methods in v1: OTP (email and WhatsApp), Google, GitHub. Facebook is excluded.                                                                    |
| D-05 | OAuth providers and OTP senders are drivers behind small interfaces defined in `domain/`.                                                                 |
| D-06 | WhatsApp OTP is delivered through GoWA, a self-hosted unofficial WhatsApp gateway. Email is the independent second channel.                               |
| D-07 | When a user has both an email and a WhatsApp identity, one OTP is sent to both channels simultaneously and is valid from either.                          |
| D-08 | Sessions are server-side, revocable, and carried in an `HttpOnly` cookie. No JWT.                                                                         |
| D-09 | `identity`, `invite`, `session`, and short-lived flow state belong to `auth`. `user` owns profile, role, and status.                                      |
| D-10 | Dependency direction is `auth → user` only. `user` never imports `auth`.                                                                                  |
| D-11 | The authenticated caller is the `user` feature's `User`, read with `user.Current(c)`. Only `auth.Middleware` sets it, so no feature other than `auth` imports `auth`.                                                                        |
| D-12 | Databases: Postgres for CI and production; SQLite via `modernc.org/sqlite` (pure Go, no CGO) for local testing.                                           |
| D-13 | Environment variables: `APP_*` for global and platform concerns, `<FEATURE>_*` for feature concerns. Mailer and WhatsApp settings use `AUTH_*`.           |
| D-14 | The first administrator is created through a CLI-generated admin invite, not through an environment secret.                                               |

---

## 3. User Flows

### 3.1 New user joins through an invite

1. A member receives the invite link `{APP_BASE_URL}/invite/{token}` in a WhatsApp group and opens it.
2. The UI asks the server to check the token. If it is expired, revoked, exhausted, or unknown, the UI shows a specific message and offers no sign-in options.
3. The member chooses a method: Google, GitHub, or OTP (entering an email address or WhatsApp number).
4. The member **verifies the identity first**: completing the OAuth round trip, or entering the OTP code. No user record exists yet.
5. If the verified identity matches an existing user, the member is signed in (see 3.3) and the invite is **not** consumed.
6. Otherwise the server issues a short-lived onboarding ticket (`domus_onboarding` cookie) and the UI shows the onboarding form, prefilled with the name and verified email from the provider when available.
7. On submit, the server creates the user, the identity (or identities), and consumes one use of the invite in a single transaction, then creates a session.
8. The member lands in the application, signed in.

### 3.2 Returning user signs in

1. The user opens `/login`, with no invite.
2. The user chooses Google, GitHub, or OTP.
3. If the identity is known and the user is active, a session is created. Otherwise the user sees a generic "not registered or not allowed" outcome (OAuth) or an indistinguishable OTP response (see OTP-6).

### 3.3 OAuth account resolution

On a successful OAuth callback with provider profile `P` (`Subject`, `Email`, `EmailVerified`, `Name`):

1. If an identity `(provider, P.Subject)` exists, sign that user in (rejecting disabled users).
2. Else if `P.EmailVerified` is true and a verified `email` identity equal to `P.Email` exists, attach the OAuth identity to that user and sign them in.
3. Else if the flow carries a valid invite, create an onboarding ticket (3.1 step 6).
4. Else redirect to `/login?error=not_registered`.

### 3.4 First administrator

1. The operator deploys Domus with a configured database.
2. The operator runs `domus invite create --role admin --max-uses 1 --expires 24h`. The command prints the invite URL once.
3. The operator opens the URL and completes onboarding. That user is the first `admin`.

---

## 4. Functional Requirements

Keywords MUST, SHOULD, and MAY follow RFC 2119.

### 4.1 Invites (`INV`)

| ID    | Requirement                                                                                                                                                                         |
| :---- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| INV-1 | An invite MUST carry: role, expiry, maximum uses, used count, optional label, creator, and revoked state.                                                                           |
| INV-2 | The invite token MUST be 256 bits from a CSPRNG, encoded base64url. Only its SHA-256 hash is stored. The full URL is shown **once**, at creation. A lost link is revoked and replaced. |
| INV-3 | An invite is valid iff it is not revoked, not expired, and `used_count < max_uses`.                                                                                                 |
| INV-4 | An invite MUST be consumed only when onboarding completes, never when it is opened or checked.                                                                                      |
| INV-5 | Consumption MUST be atomic and race-safe: a single conditional update (`used_count = used_count + 1 WHERE used_count < max_uses AND expires_at > now AND revoked_at IS NULL`) whose affected-row count decides success. |
| INV-6 | Admins MUST be able to create, list, and revoke invites through the API. Creation requires expiry and max uses; the server enforces upper bounds (see section 16).                  |
| INV-7 | Checking a token MUST NOT reveal the role. It returns only validity and, when invalid, the reason.                                                                                  |

### 4.2 Onboarding (`ONB`)

| ID    | Requirement                                                                                                                                                         |
| :---- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ONB-1 | An onboarding ticket MUST be created only after identity verification and only when a valid invite is attached to the flow.                                         |
| ONB-2 | The ticket stores the invite reference and the verified identity payload (type, subject, optional verified email, optional name). It expires after 30 minutes and is single-use. |
| ONB-3 | The ticket token is 256-bit random, stored hashed, and carried in the `domus_onboarding` cookie (`HttpOnly`, `SameSite=Lax`).                                       |
| ONB-4 | The onboarding form requires `name` (trimmed, 2 to 100 characters). All other data comes from the verified identity.                                                |
| ONB-5 | Completing onboarding MUST, in one transaction: re-validate and consume the invite (INV-5), create the user with the invite's role and status `active`, create the identity rows, and delete the ticket. |
| ONB-6 | An OAuth onboarding with a verified email MUST create both the OAuth identity and a verified `email` identity, so email OTP works afterwards.                       |
| ONB-7 | If the identity became linked to another user between verification and submission, the request MUST fail with `identity_conflict` and consume nothing.             |

### 4.3 OTP (`OTP`)

| ID    | Requirement                                                                                                                                                                                                                            |
| :---- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| OTP-1 | A code is 6 decimal digits from a CSPRNG. It is stored as HMAC-SHA256 keyed by `AUTH_SECRET`, never in plaintext or as a bare hash. Comparison MUST be constant-time.                                                                  |
| OTP-2 | A code is single-use, expires after `AUTH_OTP_TTL`, and is invalidated after `AUTH_OTP_MAX_ATTEMPTS` failed verifications. A new request for the same target invalidates the previous code.                                            |
| OTP-3 | The user supplies one identifier in a single field: an email address (contains `@`) or a WhatsApp number. Emails are trimmed and lowercased. Numbers are normalized to E.164 with Indonesia (`+62`) as the default country (`08…`, `628…`, and `+628…` all become `+628…`).            |
| OTP-4 | For a **known user**, the code is sent to **all** of that user's verified OTP identities (`email`, `whatsapp`) concurrently. Success means at least one channel accepted the message. A single code is valid regardless of channel.     |
| OTP-5 | For an **unknown identifier**, a code is sent only when the request carries a valid invite (3.1). Otherwise nothing is sent.                                                                                                           |
| OTP-6 | The login response to an OTP request MUST be identical in body, status, and shape for known, unknown, and unreachable identifiers (`202`). Delivery failures are logged, not reported. This holds for requests that carry an invite too: the response MUST NOT differ between known and unknown identifiers. Only the authenticated add-identity flow MAY report a delivery failure. |
| OTP-7 | Verification outcomes: known identity gives a session; unknown identity with a valid invite gives an onboarding ticket; otherwise a generic `otp_invalid`.                                                                              |
| OTP-8 | Resend cooldown is `AUTH_OTP_RESEND_COOLDOWN`. Sends per target are capped at `AUTH_OTP_MAX_SENDS_PER_HOUR`. Attempts and cooldown are counted **per user** (or per identifier when there is no user), not per channel.                |
| OTP-9 | Rate-limit state MUST live in the database (not process memory), because Domus can run multiple replicas behind the HPA.                                                                                                               |
| OTP-10 | OTP message text is Indonesian, states the code and its validity, and tells the recipient not to share it. It contains no link.                                                                                                       |

### 4.4 OAuth drivers (`OAU`)

| ID    | Requirement                                                                                                                                                                                                      |
| :---- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| OAU-1 | Each provider implements `domain.Provider` (see 7.4) and returns a normalized `Profile`.                                                                                                                         |
| OAU-2 | Enabled providers are listed in `AUTH_OAUTH_PROVIDERS`. An unknown name or missing credentials for an enabled provider MUST fail startup.                                                                        |
| OAU-3 | The flow MUST use a server-stored `state` (single-use, 10 minutes) and PKCE with `S256`. OIDC providers (Google) MUST also validate the nonce and ID token per the spec.                                          |
| OAU-4 | The invite is bound to the flow server-side in the stored `state` record, so the invite token never appears in a URL or in the provider round trip.                                                              |
| OAU-5 | GitHub is not OIDC. Its driver MUST fetch the user's emails and use only an email that is both primary and verified; if none exists, `Profile.Email` is empty and `EmailVerified` is false.                       |
| OAU-6 | Adding a provider MUST require only: a new file in `internal/auth/provider/`, one registry entry, and `AUTH_OAUTH_<NAME>_*` variables.                                                                           |
| OAU-7 | The redirect URI is `{APP_BASE_URL}/api/auth/oauth/{provider}/callback`.                                                                                                                                         |

### 4.5 Identities and login methods (`IDN`)

| ID    | Requirement                                                                                                                                                                      |
| :---- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| IDN-1 | A user has one or more identities. `(type, subject)` is globally unique.                                                                                                         |
| IDN-2 | Identity types: `email`, `whatsapp`, `google`, `github`. `email` and `whatsapp` are OTP-capable.                                                                                 |
| IDN-3 | Automatic linking happens only by rule 3.3 step 2 (verified provider email matching a verified `email` identity). No other automatic linking.                                    |
| IDN-4 | A signed-in user MUST be able to list, add (OTP-verified or OAuth), and remove identities through the API. The last remaining identity cannot be removed (`last_login_method`). |
| IDN-5 | Adding an identity that belongs to another user fails with `identity_conflict`.                                                                                                  |

### 4.6 Sessions (`SES`)

| ID    | Requirement                                                                                                                                                                                                                     |
| :---- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| SES-1 | A session is a 256-bit random token, stored hashed in the database, carried in the `domus_session` cookie: `HttpOnly`, `Path=/`, `SameSite=Lax`, `Secure` when `APP_BASE_URL` is `https`.                                         |
| SES-2 | Expiry is sliding: `AUTH_SESSION_TTL` from last activity. `last_seen_at` updates are throttled (at most once per minute per session). A session also expires 90 days after creation regardless of activity.                                                                                            |
| SES-3 | A new session MUST be created at every sign-in; no session identifier is reused across authentication events.                                                                                                                    |
| SES-4 | Logout deletes the session server-side.                                                                                                                                                                                          |
| SES-5 | Disabling a user deletes all of that user's sessions. A disabled user cannot sign in.                                                                                                                                            |
| SES-6 | State-changing requests (non-`GET`) MUST be rejected when the `Origin` header (or `Sec-Fetch-Site`) does not match `APP_BASE_URL`.                                                                                                |
| SES-7 | All `/api` routes are **deny-by-default**: they require a valid session unless a route is explicitly marked public. Public routes are limited to the auth entry points in section 8.                                             |

### 4.7 User feature (`USR`)

| ID    | Requirement                                                                                                                                                                       |
| :---- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| USR-1 | A user has: id (UUID), name, role (`admin` or `member`), status (`active` or `disabled`), timestamps.                                                                             |
| USR-2 | Any signed-in user can read and update their own name (`/api/users/me`).                                                                                                          |
| USR-3 | Admins can list users, read a user, change a user's role, and enable or disable a user.                                                                                           |
| USR-4 | The system MUST refuse to demote or disable the last active admin (`last_admin`).                                                                                                 |
| USR-5 | The `user` feature exposes a root-package API (see 7.3) to `auth` for lookup and transactional creation. It exposes no session or credential concepts. |
| USR-6 | Any feature reads the authenticated caller with `user.Current` and guards routes with `user.RequireRole`. Only `auth.Middleware` sets the current user (see 7.6).                                   |

### 4.8 CLI (`CLI`)

| ID    | Requirement                                                                                                                                                              |
| :---- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CLI-1 | `domus invite create --role <admin\|member> --expires <duration> --max-uses <n> [--label <text>]` prints the invite URL once.                                           |
| CLI-2 | `domus invite list` and `domus invite revoke <id>` manage existing invites.                                                                                              |
| CLI-3 | The CLI uses the same `auth` root API as the HTTP handlers and requires only `APP_DATABASE_URL` and `APP_BASE_URL`. It does not start the HTTP server and skips the channel and provider startup validation of 9.3.                       |

---

## 5. Security and Privacy Requirements

| ID     | Requirement                                                                                                                                                                                                  |
| :----- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| SEC-1  | All tokens (invite, session, onboarding, OAuth state) are 256-bit CSPRNG values stored as SHA-256 hashes. Raw tokens are never persisted or logged.                                                          |
| SEC-2  | OTP codes are stored per OTP-1 and are never logged, including in error paths.                                                                                                                               |
| SEC-3  | Request logging MUST redact invite tokens, session cookies, onboarding cookies, OTP codes, OAuth `code` and `state` parameters, and provider secrets.                                                       |
| SEC-4  | OTP request responses MUST not enable enumeration of registered emails or numbers (OTP-6).                                                                                                                   |
| SEC-5  | OTP and onboarding endpoints MUST be rate limited per target (database-backed, OTP-8) and additionally per client IP on a best-effort basis (requires `APP_TRUSTED_PROXIES`, see 14.1).                                                                 |
| SEC-6  | Provider-asserted emails are trusted for linking only when marked verified.                                                                                                                                  |
| SEC-7  | `AUTH_SECRET`, OAuth client secrets, the SMTP password, and the GoWA password are secrets: delivered through Kubernetes secrets (`extraEnvFrom`), never committed, never echoed in logs or errors.            |
| SEC-8  | Invite links in a shared group are bearer tokens visible to every member. Mitigations are mandatory expiry, a use cap, and revocation (INV-3, INV-6).                                                       |
| SEC-9  | Expired invites, OTP challenges, onboarding tickets, OAuth states, and sessions MUST be purged by a periodic cleanup so personal data is not retained past its purpose.                                     |
| SEC-10 | Email addresses and phone numbers are personal data. Collection is limited to what sign-in needs. Obligations under Indonesian personal data protection law apply and SHOULD be reviewed before launch.     |
| SEC-11 | Authorization checks (`admin` role) are enforced in handlers via `user.RequireRole`; services receive the acting user id as an argument and never read HTTP context.                                            |
| SEC-12 | Authentication events (sign-in success and failure, onboarding, invite create and revoke, identity added or removed, user disabled) are emitted as structured `slog` entries without sensitive values.       |

---

## 6. Domain Model

All entities are defined in the unified `ent/schema/`. IDs are UUIDs generated application-side for portability. Time columns are UTC.

### 6.1 `user` (owned by `user`)

| Field                       | Type       | Notes                                |
| :-------------------------- | :--------- | :----------------------------------- |
| `id`                        | UUID       | Primary key                          |
| `name`                      | string     | 2 to 100 characters                  |
| `role`                      | enum       | `admin`, `member`                    |
| `status`                    | enum       | `active`, `disabled`                 |
| `created_at`, `updated_at`  | timestamp  |                                      |

### 6.2 `identity` (owned by `auth`)

| Field          | Type      | Notes                                                              |
| :------------- | :-------- | :----------------------------------------------------------------- |
| `id`           | UUID      | Primary key                                                        |
| `user`         | edge      | Required; deleting a user cascades                                 |
| `type`         | enum      | `email`, `whatsapp`, `google`, `github`                            |
| `subject`      | string    | Normalized email, E.164 number, or provider subject                |
| `email`        | string?   | Verified email snapshot for OAuth identities                       |
| `verified_at`  | timestamp |                                                                    |
| `last_used_at` | timestamp |                                                                    |
| `created_at`   | timestamp |                                                                    |

Unique index on `(type, subject)`.

### 6.3 `invite` (owned by `auth`)

| Field        | Type       | Notes                                      |
| :----------- | :--------- | :----------------------------------------- |
| `id`         | UUID       |                                            |
| `token_hash` | bytes      | Unique                                     |
| `role`       | enum       | Role granted at onboarding                 |
| `label`      | string?    | Free text, for example the group name      |
| `expires_at` | timestamp  |                                            |
| `max_uses`   | int        |                                            |
| `used_count` | int        | Default 0                                  |
| `revoked_at` | timestamp? |                                            |
| `created_by` | edge?      | Null when created through the CLI          |
| `created_at` | timestamp  |                                            |

### 6.4 `session` (owned by `auth`)

| Field                       | Type      | Notes                 |
| :-------------------------- | :-------- | :-------------------- |
| `id`                        | UUID      |                       |
| `token_hash`                | bytes     | Unique                |
| `user`                      | edge      | Cascade on delete     |
| `expires_at`                | timestamp |                       |
| `last_seen_at`              | timestamp |                       |
| `user_agent`                | string?   | Truncated             |
| `created_at`                | timestamp |                       |

### 6.5 Short-lived flow state (owned by `auth`)

| Entity              | Purpose                                                                                                   | Key fields                                                                                                        | Lifetime            |
| :------------------ | :-------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------- | :------------------ |
| `otp_challenge`     | One issued code and its attempt and send counters.                                                        | purpose (`login`, `onboarding`, `add_identity`), `user` (nullable), target identifier, `code_hmac`, `attempts`, `expires_at`, `consumed_at`, invite ref (nullable), `created_at` | `AUTH_OTP_TTL`      |
| `onboarding_ticket` | Verified-but-unregistered identity waiting for the onboarding form.                                       | `token_hash`, invite ref, identity payload (type, subject, email, name), `expires_at`                             | 30 minutes          |
| `oauth_state`       | Server-side record of an in-flight OAuth round trip.                                                      | `state_hash`, provider, PKCE verifier, nonce, purpose (`signin`, `link`), invite ref (nullable), `user` (nullable), `expires_at` | 10 minutes          |

---

## 7. Architecture

### 7.1 Package layout

```txt
.golangci.yml                  # depguard: feature and layer boundaries
docs/auth.md
ent/schema/                    # user, identity, invite, session,
                               # otp_challenge, onboarding_ticket, oauth_state
internal/
├── app/                       # wiring: config, router, auth middleware
├── platform/httpx/            # envelope, errors
├── auth/
│   ├── auth.go                # New(cfg, client, users), Register(r): wiring only
│   ├── api.go                 # cross-feature API: Middleware, CreateInvite, ...
│   ├── config.go              # Config (AUTH_*)
│   ├── domain/                # entities, errors.go, dto.go, Provider, Sender, Profile
│   ├── repo/                  # Ent-backed storage
│   ├── service/               # invite, otp, oauth, onboarding, session, identity
│   ├── handler/               # thin Fiber handlers + Swag annotations
│   ├── provider/              # google.go, github.go, registry.go
│   └── sender/                # email.go, whatsapp.go (GoWA)
└── user/
    ├── user.go
    ├── api.go                 # cross-feature API: FindByID, Create, Current, RequireRole
    ├── domain/
    ├── repo/
    ├── service/
    └── handler/
cmd/domus/                     # serve + invite subcommands
ui/app/features/auth/          # login, invite, onboarding
ui/app/features/user/          # renamed from users
```

### 7.2 Dependency rules

- `handler → service → repo → domain` inside each feature. The boundaries in this section are enforced by depguard (`.golangci.yml`).
- `provider/` and `sender/` implement interfaces declared in `auth/domain` and depend only on `domain`. `service` never imports them; `auth.go` wires them from `Config`.
- Between `auth` and `user` the only edge is **`auth → user`**, and only through the root package `internal/user`. `user` MUST NOT import `internal/auth`, which prevents an import cycle. Any other feature may import `internal/user` to read the current user and MUST NOT import `internal/auth`.
- `internal/platform/httpx` MUST NOT import any feature.

### 7.3 Cross-feature API (`api.go`)

Each feature's root package exposes its cross-feature surface in `api.go`: exported methods and type aliases to `domain` types. `<feature>.go` stays wiring only. Indicative surface:

```go
// internal/user/api.go
func (f *Feature) FindByID(ctx context.Context, id uuid.UUID) (User, error)
func (f *Feature) Create(ctx context.Context, db *ent.Client, in NewUser) (User, error)
func Current(c fiber.Ctx) (User, bool)  // set by auth.Middleware
func SetCurrent(c fiber.Ctx, u User)    // auth.Middleware only
func RequireRole(role Role) fiber.Handler

// internal/auth/api.go
func (f *Feature) Middleware() fiber.Handler
func (f *Feature) CreateInvite(ctx context.Context, in NewInvite) (InviteLink, error)
func (f *Feature) ListInvites(ctx context.Context) ([]Invite, error)
func (f *Feature) RevokeInvite(ctx context.Context, id uuid.UUID) error
```

### 7.4 Drivers

```go
// internal/auth/domain (illustrative)
type Profile struct {
    Subject       string
    Email         string
    EmailVerified bool
    Name          string
}

type AuthRequest struct {
    State, CodeChallenge, Nonce, RedirectURL string
}

type Provider interface {
    Name() string
    AuthURL(req AuthRequest) string
    Exchange(ctx context.Context, code, codeVerifier, nonce string) (Profile, error)
}

type Channel string // "email", "whatsapp"

type Sender interface {
    Channel() Channel
    Send(ctx context.Context, to, code string, ttl time.Duration) error
}
```

- `provider.Registry` is built from `Config`: it instantiates only providers named in `AUTH_OAUTH_PROVIDERS`, failing startup on unknown names or missing credentials.
- A `Sender` is enabled iff its configuration is present (`AUTH_EMAIL_SMTP_HOST`, `AUTH_WHATSAPP_GOWA_URL`).
- `Sender` has two implementations from day one, so the interface is justified under the "no single-implementation interface" rule. `Provider` likewise has two (Google, GitHub).
- Adding a provider: create `provider/<name>.go`, add one registry entry, document `AUTH_OAUTH_<NAME>_CLIENT_ID` and `_CLIENT_SECRET`. A test using a fake provider MUST prove services need no change (AC-13).
- When a known user has several OTP-capable identities, `service` calls each matching `Sender` concurrently with a short timeout and treats the send as successful if at least one succeeds.

### 7.5 Transactions

Onboarding spans two features. `auth.service` opens `client.Tx(ctx)`, passes `tx.Client()` into `user.Create(ctx, db, in)`, writes identities and consumes the invite with the same tx client, then commits. Passing a tx-scoped `*ent.Client` as an argument is the sanctioned cross-feature transaction pattern.

### 7.6 Current user

The authenticated caller is the `user` feature's `User`, so no feature other than `auth` needs to import `auth`. `auth.Middleware()` resolves the session cookie, loads the user through `user.FindByID`, rejects disabled users, and calls `user.SetCurrent`. `internal/app/fiber.go` installs it on the `/api` group before feature registration. Handlers call `user.Current(c)`, or guard a route with `user.RequireRole(role)`, and pass the user id to services as an ordinary argument. Services never import `fiber`. `Role` is a typed enum owned by `user`.

### 7.7 Persistence

- All repos use the Ent client; no raw SQL that is not valid on both Postgres and SQLite.
- Migrations and schema bootstrap follow the prerequisite database issue.
- SQLite is for local and unit testing only. CI MUST run the repo and race tests against Postgres, which is the authoritative database.
- A periodic cleanup (simple ticker started with the server) purges expired rows (SEC-9). `// ponytail:` single-process ticker; safe because deletes are idempotent.

---

## 8. HTTP API

All routes live under `/api`, use the `{ data, meta, error }` envelope, and carry Swag annotations; DTOs live in `domain/dto.go`. Access levels: **Public**, **Session**, **Admin**.

| Method | Path                                          | Access  | Purpose                                                                            |
| :----- | :-------------------------------------------- | :------ | :--------------------------------------------------------------------------------- |
| GET    | `/api/auth/providers`                         | Public  | Enabled OAuth providers and OTP channels, for the login UI.                        |
| POST   | `/api/auth/invite/check`                      | Public  | Body `{ token }`. Returns validity and reason only.                                |
| POST   | `/api/auth/otp/request`                       | Public  | Body `{ identifier, inviteToken? }`. Always `202` for login (OTP-6).               |
| POST   | `/api/auth/otp/verify`                        | Public  | Body `{ identifier, code, inviteToken? }`. Sets session or onboarding cookie.      |
| POST   | `/api/auth/oauth/{provider}/start`            | Public  | Body `{ inviteToken? }`. Returns `{ authUrl }`.                                    |
| GET    | `/api/auth/oauth/{provider}/callback`         | Public  | Provider redirect. Sets cookie, then redirects to `/`, `/onboarding`, or `/login?error=…`. |
| GET    | `/api/auth/onboarding`                        | Public  | Reads the pending ticket (prefill data). Requires onboarding cookie.               |
| POST   | `/api/auth/onboarding`                        | Public  | Body `{ name }`. Completes onboarding, sets session cookie.                        |
| POST   | `/api/auth/logout`                            | Session | Deletes the session.                                                               |
| GET    | `/api/auth/session`                           | Session | Current user and role.                                                             |
| GET    | `/api/auth/identities`                        | Session | Lists the caller's login methods.                                                  |
| POST   | `/api/auth/identities/otp/request`            | Session | Starts adding an `email` or `whatsapp` identity.                                   |
| POST   | `/api/auth/identities/otp/verify`             | Session | Completes adding an OTP identity.                                                  |
| POST   | `/api/auth/identities/oauth/{provider}/start` | Session | Starts linking an OAuth identity (shared callback).                                |
| DELETE | `/api/auth/identities/{id}`                   | Session | Removes an identity (IDN-4).                                                       |
| POST   | `/api/auth/invites`                           | Admin   | Body `{ role, expiresAt, maxUses, label? }`. Returns `{ id, url }`; URL shown once. |
| GET    | `/api/auth/invites`                           | Admin   | Lists invites with status and usage, no URLs.                                      |
| DELETE | `/api/auth/invites/{id}`                      | Admin   | Revokes an invite.                                                                 |
| GET    | `/api/users`                                  | Admin   | Lists users.                                                                       |
| GET    | `/api/users/{id}`                             | Admin   | Reads a user.                                                                      |
| PATCH  | `/api/users/{id}`                             | Admin   | Changes role or status (USR-3, USR-4).                                             |
| GET    | `/api/users/me`                               | Session | Reads own profile.                                                                 |
| PATCH  | `/api/users/me`                               | Session | Updates own name.                                                                  |

Invite and OTP tokens travel in request bodies, never in API paths or query strings (SEC-3). The `/invite/:token` path exists only as a UI route.

`/api/users/me` MUST be registered before `/api/users/{id}`.

---

## 9. Configuration

### 9.1 Environment variable conventions

These rules are recorded in `AGENTS.md` (the required change is listed in section 17 of this document) and apply to every feature.

1. `APP_*` is for global and platform concerns (base URL, database, trusted proxies, and so on).
2. `<FEATURE>_*` is for a single feature, for example `AUTH_*`. A driver adds a sub-namespace: `<FEATURE>_<DRIVER>_*`, for example `AUTH_OAUTH_GOOGLE_*`.
3. A feature declares its variables in its own `Config` struct in the feature root (`config.go`) with `env:"..."` tags and defaults. `internal/app.Config` embeds it using go-envconfig's `prefix` option. A feature never reads another feature's variables.
4. Defaults live in struct tags. Required variables have no default and fail startup when absent or invalid. Invalid values fail fast; they are never silently ignored.
5. Every variable is documented in its feature PRD (name, type, default, required, secret) and surfaced through `charts/values.yaml`. Secrets are supplied through `extraEnvFrom` secret references, never through plain `extraEnv`.
6. Do not add a variable for something that is a constant.

### 9.2 Platform variables this PRD depends on

| Variable           | Type   | Required | Notes                                                                                        |
| :----------------- | :----- | :------- | :------------------------------------------------------------------------------------------- |
| `APP_BASE_URL`     | URL    | Yes      | Public origin without trailing slash. Used for invite links, OAuth redirect URIs, `Secure` cookies, and origin checks. |
| `APP_DATABASE_URL` | string | Yes      | Defined by the database bootstrap issue. `postgres://…` in CI and production; a SQLite DSN locally. |

Client IP detection behind the ingress relies on Fiber's trusted proxy configuration, owned by the platform (`APP_TRUSTED_PROXIES`, a prerequisite, see 14.1).

### 9.3 `AUTH_*` variables

| Variable                              | Type     | Default | Required                         | Secret | Description                                                         |
| :------------------------------------ | :------- | :------ | :------------------------------- | :----- | :------------------------------------------------------------------ |
| `AUTH_SECRET`                         | string   | none    | Yes (at least 32 bytes)          | Yes    | HMAC key for OTP code hashing.                                      |
| `AUTH_SESSION_TTL`                    | duration | `720h`  | No                               | No     | Sliding session lifetime.                                           |
| `AUTH_OTP_TTL`                        | duration | `10m`   | No                               | No     | OTP validity.                                                       |
| `AUTH_OTP_MAX_ATTEMPTS`               | int      | `5`     | No                               | No     | Failed verifications before invalidation.                           |
| `AUTH_OTP_RESEND_COOLDOWN`            | duration | `60s`   | No                               | No     | Minimum gap between sends per target.                               |
| `AUTH_OTP_MAX_SENDS_PER_HOUR`         | int      | `5`     | No                               | No     | Send cap per target per hour.                                       |
| `AUTH_OAUTH_PROVIDERS`                | list     | empty   | No                               | No     | Comma-separated driver names, for example `google,github`.          |
| `AUTH_OAUTH_GOOGLE_CLIENT_ID`         | string   | none    | If `google` enabled              | No     |                                                                     |
| `AUTH_OAUTH_GOOGLE_CLIENT_SECRET`     | string   | none    | If `google` enabled              | Yes    |                                                                     |
| `AUTH_OAUTH_GITHUB_CLIENT_ID`         | string   | none    | If `github` enabled              | No     |                                                                     |
| `AUTH_OAUTH_GITHUB_CLIENT_SECRET`     | string   | none    | If `github` enabled              | Yes    |                                                                     |
| `AUTH_EMAIL_SMTP_HOST`                | string   | none    | Enables the email channel        | No     |                                                                     |
| `AUTH_EMAIL_SMTP_PORT`                | int      | `587`   | No                               | No     |                                                                     |
| `AUTH_EMAIL_SMTP_USERNAME`            | string   | none    | If the server requires it        | No     |                                                                     |
| `AUTH_EMAIL_SMTP_PASSWORD`            | string   | none    | If the server requires it        | Yes    |                                                                     |
| `AUTH_EMAIL_FROM`                     | string   | none    | If the email channel is enabled  | No     | Sender address.                                                     |
| `AUTH_WHATSAPP_GOWA_URL`              | URL      | none    | Enables the WhatsApp channel     | No     | Base URL of the GoWA server.                                        |
| `AUTH_WHATSAPP_GOWA_USERNAME`         | string   | none    | If GoWA basic auth is enabled    | No     | Confirm exact GoWA fields against its documentation at build time.  |
| `AUTH_WHATSAPP_GOWA_PASSWORD`         | string   | none    | If GoWA basic auth is enabled    | Yes    |                                                                     |

Server startup validation (not applied to the CLI): `AUTH_SECRET` is required; at least one OTP channel **or** one OAuth provider MUST be enabled; every enabled provider has complete credentials; enabled channels have complete settings.

Constants (not configurable): OTP length 6, OAuth state lifetime 10 minutes, onboarding ticket lifetime 30 minutes, session absolute lifetime 90 days, cookie names `domus_session` and `domus_onboarding`.

GoWA is a separate application that must run on its own with a dedicated WhatsApp number. Its deployment is an operations concern; Domus only needs its address and credentials.

---

## 10. UI Requirements

The UI follows the existing React Router Framework Mode (`ssr: false`) conventions: routes in `features/auth/routes.ts`, data access under `api/` with TanStack Query and `openapi-fetch`, strings through `ui/app/shared/i18n` (Indonesian locale), and no `loader` or `action` exports.

### 10.1 v1 screens

| Route          | Behavior                                                                                                                                                                                                                                             |
| :------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/login`       | One input accepting an email or WhatsApp number plus a "send code" action, then a code step (`inputmode="numeric"`, `autocomplete="one-time-code"`, 6 digits, resend with cooldown timer). Provider buttons come from `GET /api/auth/providers`. Shows `error` query results from the OAuth callback. |
| `/invite/:token` | Calls `invite/check`. Valid: same method choices as login, with the invite carried in request bodies. Invalid: a specific message (expired, revoked, exhausted, unknown) and no sign-in options.                                                      |
| `/onboarding`  | Requires the onboarding cookie. Shows the name field (prefilled when known) and the verified contact. Submitting completes onboarding and enters the app. Expired ticket sends the user back to the invite or login flow with a clear message.        |

### 10.2 Behavior

- A client-side guard on the app layout route calls `GET /api/auth/session`: unauthenticated users go to `/login`; authenticated users visiting `/login` go to `/`.
- All screens MUST be keyboard accessible with labelled inputs and visible error text. The invite and login screens MUST work at phone width, since links arrive through WhatsApp.
- The UI MUST handle OAuth being unavailable inside an in-app browser by keeping OTP visible and prominent on every screen.
- The existing placeholder `ui/app/features/users` is renamed `user` for consistency with `internal/user`.

### 10.3 Deferred UI (API ready in v1)

Invite management, user administration, and login-method management screens are follow-up work. Until then, admins use the CLI and API for invites and users.

---

## 11. Error Codes

Sentinel errors live in `domain/errors.go` and map to HTTP statuses through `httpx`. The UI resolves each code to a message in `features/auth/api/errors.ts`.

| Code                                                       | Meaning                                                              |
| :--------------------------------------------------------- | :------------------------------------------------------------------- |
| `invite_not_found`, `invite_expired`, `invite_revoked`, `invite_exhausted` | Invite is unusable for the stated reason.                |
| `otp_invalid`                                              | Wrong, expired, or exhausted code (indistinguishable by design).     |
| `otp_rate_limited`                                         | Cooldown or hourly cap reached.                                      |
| `otp_delivery_failed`                                      | Add-identity only (authenticated caller): no channel accepted the message.                    |
| `not_registered`                                           | OAuth identity unknown and no valid invite.                          |
| `provider_unavailable`                                     | Provider not enabled or upstream failure.                            |
| `onboarding_expired`                                       | Ticket missing, expired, or already used.                            |
| `identity_conflict`                                        | Identity already belongs to another user.                            |
| `last_login_method`                                        | Cannot remove the last identity.                                     |
| `last_admin`                                               | Cannot demote or disable the last active admin.                      |
| `user_disabled`                                            | Account is disabled.                                                 |
| `unauthenticated`, `forbidden`                             | Missing session or insufficient role.                                |

---

## 12. Acceptance Criteria

| ID    | Scenario                                                                                                                                                              |
| :---- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AC-01 | Opening a valid invite link shows the sign-in options and consumes nothing.                                                                                           |
| AC-02 | Expired, revoked, exhausted, and unknown invites each show their specific message and no sign-in options.                                                             |
| AC-03 | A new user who verifies an email by OTP and submits a name becomes an active user with the invite's role. The invite's `used_count` increases by exactly 1.            |
| AC-04 | The same outcome as AC-03 through Google, including a verified `email` identity alongside the Google identity.                                                        |
| AC-05 | With `max_uses = 1` and two simultaneous onboarding submissions, exactly one succeeds and the other receives `invite_exhausted`. Verified on Postgres in CI.          |
| AC-06 | An existing user signs in by OTP at `/login` with no invite.                                                                                                          |
| AC-07 | An OTP request for an unknown identifier without an invite returns a response identical to a known identifier's, and no message is sent.                              |
| AC-08 | A user with email and WhatsApp identities receives one code on both channels. The code verifies from either and only once.                                            |
| AC-09 | After `AUTH_OTP_MAX_ATTEMPTS` wrong codes the challenge is invalid, even for the correct code.                                                                        |
| AC-10 | A Google sign-in whose verified email matches an existing verified `email` identity attaches to that user and signs them in without an invite.                        |
| AC-11 | A disabled user cannot sign in, and existing sessions stop working immediately.                                                                                       |
| AC-12 | The last active admin cannot be demoted or disabled.                                                                                                                  |
| AC-13 | A test registers a fake OAuth provider with no changes to `service`, `handler`, or UI logic, and the flow completes.                                                  |
| AC-14 | If the WhatsApp sender fails and the email sender succeeds, the user receives the code and the request succeeds.                                                      |
| AC-15 | No test log output contains a raw token, OTP code, OAuth code, or secret.                                                                                             |
| AC-16 | The CLI creates an admin invite whose URL, opened in the UI, produces the first admin.                                                                                |
| AC-17 | Logout invalidates the session server-side; replaying the old cookie returns `unauthenticated`.                                                                       |
| AC-18 | A non-`GET` request with a mismatched `Origin` is rejected. A route without an explicit public marker rejects unauthenticated requests.                               |

---

## 13. Testing and Verification

- **Unit tests** for services use in-memory fakes for `Sender` and `Provider` and SQLite (modernc) for repos.
- **Concurrency test** for INV-5 and AC-05 with many parallel consumers, run with `go test -race`.
- **Integration tests** for repos and the full onboarding transaction run against Postgres in CI.
- **Handler tests** cover envelope, status mapping, cookies, origin checks, and deny-by-default.
- **UI tests** (`bun test`) cover the login code step, invite states, and onboarding prefill.

Run the verification commands listed in `AGENTS.md` before finishing each change.

---

## 14. Delivery Plan

### 14.1 Prerequisites (separate issues, before implementation)

1. **Environment prefix migration**: rename existing unprefixed variables to `APP_*`.
2. **Database bootstrap**: Ent setup under `ent/`, Postgres and SQLite drivers, migration strategy, `mise` codegen task, `APP_DATABASE_URL`, `APP_BASE_URL`, CI Postgres service.
3. **`AGENTS.md` update** with the changes in section 17.
4. **Boundary lint**: `.golangci.yml` with depguard rules, run in CI.
5. **Trusted proxies**: `APP_TRUSTED_PROXIES`, needed for per-IP rate limits (SEC-5).

### 14.2 Phases

| Phase | Scope                                                                                                                                      |
| :---- | :----------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | `user` feature: schema, service, handlers, `api.go`, UI rename. `user.Current`, `user.SetCurrent`, and `user.RequireRole`.                                       |
| 2     | `auth` core: invite, session, middleware (deny-by-default), onboarding ticket, CLI invite commands. First admin can sign up by invite.    |
| 3     | OTP with the email sender: challenge, rate limits, cleanup job.                                                                            |
| 4     | WhatsApp (GoWA) sender and simultaneous multi-channel delivery.                                                                            |
| 5     | OAuth driver framework, Google, then GitHub, plus account resolution (3.3) and identity management endpoints.                              |
| 6     | UI screens: login, invite, onboarding, route guard.                                                                                        |

Each phase is an independently reviewable change following the multi-stage atomic commit rules in `AGENTS.md`.

---

## 15. Risks

| Risk                                                                      | Impact                                         | Mitigation                                                                                                             |
| :------------------------------------------------------------------------ | :--------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------- |
| GoWA is unofficial; the WhatsApp number can be banned                     | WhatsApp OTP stops working                     | Dedicated number, email as independent channel (D-07), sender behind an interface so the official Meta API can replace it. |
| Invite link leaks beyond the group                                        | Unwanted accounts                              | Expiry, use cap, revocation, label per link, structured logs for review.                                               |
| OAuth fails inside WhatsApp's in-app browser                              | Members cannot use Google                      | Keep OTP prominent; test on Android and iOS before launch.                                                             |
| Email deliverability (SMTP reputation)                                    | Delayed or lost codes                          | Authenticated SMTP (SPF, DKIM), WhatsApp as the other channel.                                                         |
| SQLite and Postgres behave differently                                    | Bugs pass locally, fail in production          | Postgres is mandatory in CI; avoid dialect-specific features.                                                          |
| Automatic linking by verified email                                       | Account takeover if a provider mislabels email | Link only on provider-verified email matching a verified identity (IDN-3).                                             |
| Personal data (email, phone) retention                                    | Privacy and compliance exposure                | Minimal collection, cleanup job (SEC-9), review under Indonesian data protection law before launch.                    |

---

## 16. Assumptions

These are defaults chosen to complete the draft. Change them in a later version if needed.

- Roles are limited to `admin` and `member`; full RBAC is a separate PRD.
- Invite creation requires an expiry and a use count. Server-side upper bounds: expiry at most 30 days, maximum uses at most 500.
- Session lifetime is 30 days sliding, with a 90-day absolute maximum.
- Phone numbers without a country code are assumed to be Indonesian.
- A user may have no `email` identity if the OAuth provider supplied no verified email.
- OTP message language is Indonesian, matching the default UI locale.
- The invite URL is shown only once; there is no "copy link again".
- Local development configures OTP through test fakes or a real sender, with no dedicated log-only sender in v1.

---

## 17. Required `AGENTS.md` Changes

To be applied in a separate `docs(docs)` commit. `AGENTS.md` stays lean: it holds rules an agent cannot infer from code or config, and boundaries that depguard enforces are only pointed to, not described.

1. Add the **environment variable conventions** of section 9.1.
2. Allow a root-package **`api.go`** for cross-feature calls (exported methods and aliases), and keep `<feature>.go` wiring-only. State the rule: cross-feature dependency edges must be acyclic and are documented in the PRD that introduces them.
3. Document the **transaction pattern**: cross-feature atomic work passes a tx-scoped `*ent.Client` as an argument.
4. Document the **current user**: handlers read it with `user.Current` and pass the user id to services; services never read HTTP context.
5. Record **deny-by-default** for `/api` routes, with explicit public markers.
6. Point to `.golangci.yml` as the enforcement of feature and layer boundaries.
