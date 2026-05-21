# PRD-001: Authentication & Authorization

**Status:** Accepted
**Created:** 2026-05-21
**Revised:** 2026-05-21
**Deciders:** Anthonius Munthi
**Context:** API (`api/`), Dashboard (`dash/`)

---

## Revision History

| Version | Date | Changes |
|---|---|---|
| 1.0 | 2026-05-21 | Initial acceptance |

---

## 1. Overview

Domus requires a secure, multi-tenant authentication and authorization system that supports parish-scoped identity, role-based access control, and multiple OAuth2 providers. Each parish operates as an isolated organization (tenant). Users authenticate via OAuth2 providers or email/password (development only), are assigned roles within a parish, and permissions are enforced at the API layer.

---

## 2. Goals

- Provide secure, standards-compliant authentication for all Domus users.
- Support multi-tenant isolation: one parish = one organization.
- Enable flexible, role-based permission enforcement using `resource:action` strings.
- Support OAuth2 providers: Google, Facebook, GitHub.
- Support email/password authentication in development environments only.
- Allow parish admins to manage user membership, roles, and session revocation.
- Implement an approval workflow: new registrants default to `PENDING` (guest) state until approved by a parish admin.

---

## 3. Non-Goals

- Social login beyond Google, Facebook, and GitHub is out of scope for v1.
- Fine-grained attribute-based access control (ABAC) is out of scope.
- Cross-parish user sharing or federation is out of scope.
- Self-service role creation by end users is out of scope.

---

## 4. Background & Motivation

Without a formalised auth system, each endpoint is vulnerable to inconsistent access enforcement across tenants. As Domus scales to multiple parishes, a unified auth contract ensures:

- Consistent permission enforcement regardless of parish configuration.
- Support staff can diagnose access issues remotely using structured error codes (`DOMUS-AUTH-*`).
- Parish admins retain full control over their tenant's user base without cross-tenant exposure.

---

## 5. User Personas

| Persona | Description |
|---|---|
| **Super Admin** | System-level operator. Manages organizations (parishes). Not parish-scoped. |
| **Parish Admin** | Manages users, roles, and permissions within their parish. |
| **Staff** | Parish employee assigned a role (e.g., Finance Staff, Secretary). |
| **Member** | Registered parishioner. Read-only access to relevant resources. |
| **Guest** | Registered but not yet approved. Severely restricted access. |

---

## 6. User Stories

### 6.1 Registration & Onboarding

| ID | As a… | I want to… | So that… |
|---|---|---|---|
| US-001 | Visitor | Register via Google, Facebook, or GitHub OAuth2 | I can join a parish without creating a new password |
| US-002 | Visitor | Register via email/password (dev only) | I can test the system without an OAuth2 provider |
| US-003 | Guest | See my pending approval status after registration | I know my account is awaiting review |
| US-004 | Parish Admin | See a list of pending users | I can review and approve or reject new registrants |
| US-005 | Parish Admin | Approve a pending user and assign them a role | The user gains appropriate access to the system |
| US-006 | Parish Admin | Kick (revoke) an active user | Spammers or inactive users are removed from the parish |

### 6.2 Authentication

| ID | As a… | I want to… | So that… |
|---|---|---|---|
| US-007 | User | Log in via OAuth2 | I can authenticate securely without managing a password |
| US-008 | User | Receive an access token and refresh token upon login | My session is maintained securely |
| US-009 | User | Have my access token refreshed automatically | I am not unexpectedly logged out mid-session |
| US-010 | User | Log out from my current session | My session is terminated and tokens are invalidated |
| US-011 | Parish Admin | View all active sessions for a user | I have visibility into concurrent logins |
| US-012 | Parish Admin | Revoke any active session | I can forcibly terminate a compromised or suspicious session |

### 6.3 Authorization

| ID | As a… | I want to… | So that… |
|---|---|---|---|
| US-013 | System | Enforce `resource:action` permissions on every protected endpoint | Unauthorised access is rejected consistently |
| US-014 | Parish Admin | Create and manage roles within my parish | I can define custom access profiles per our operational needs |
| US-015 | Parish Admin | Assign one or more roles to a user | The user inherits the correct set of permissions |
| US-016 | Parish Admin | View the permission set of any role | I can audit what each role can do |

---

## 7. Functional Requirements

### 7.1 Authentication Providers

| Provider | Environment | Protocol |
|---|---|---|
| Email / Password | Development only | N/A (internal) |
| Google | All | OAuth2 / OIDC |
| Facebook | All | OAuth2 |
| GitHub | All | OAuth2 |

All OAuth2 provider credentials (client ID, client secret, redirect URIs) MUST be configured via environment variables. No provider configuration is stored in the database.

```
OAUTH_GOOGLE_CLIENT_ID=
OAUTH_GOOGLE_CLIENT_SECRET=
OAUTH_GOOGLE_REDIRECT_URI=

OAUTH_FACEBOOK_CLIENT_ID=
OAUTH_FACEBOOK_CLIENT_SECRET=
OAUTH_FACEBOOK_REDIRECT_URI=

OAUTH_GITHUB_CLIENT_ID=
OAUTH_GITHUB_CLIENT_SECRET=
OAUTH_GITHUB_REDIRECT_URI=
```

Email/password authentication MUST be disabled in all non-development environments. The application MUST check `APP_ENV` at bootstrap and refuse to register email/password routes unless `APP_ENV=development`.

### 7.2 Session Management

| Property | Value |
|---|---|
| Access Token | JWT, short-lived, 15 minutes |
| Refresh Token | Opaque token, long-lived, 30 days |
| Refresh Token Storage | Database (revocable) |
| Rotation | Refresh token rotated on every use |
| Revocation | Supported — individual session or all sessions for a user |

Access tokens MUST include the following claims:

| Claim | Description |
|---|---|
| `sub` | User UUID |
| `org_id` | Organization (Parish) UUID |
| `roles` | Array of role slugs assigned in this org |
| `permissions` | Flattened array of permission strings |
| `exp` | Expiry timestamp |
| `jti` | Unique token ID |

### 7.3 User Lifecycle

```
[Visitor]
    │
    ▼ Register (OAuth2 or email/password)
[PENDING] ──── Admin Rejects ──▶ [REJECTED]
    │
    ▼ Admin Approves + assigns Role
[ACTIVE]
    │
    ▼ Admin Kicks
[KICKED]
```

| State | Description | Access Level |
|---|---|---|
| `PENDING` | Registered, awaiting admin approval | Guest — read public parish info only |
| `ACTIVE` | Approved and role-assigned | Per assigned role permissions |
| `REJECTED` | Rejected by admin | No access |
| `KICKED` | Removed by admin post-approval | No access, all sessions revoked immediately |

Transitioning a user to `KICKED` MUST immediately invalidate all active refresh tokens for that user.

### 7.4 Role & Permission Model

```
Organization (Parish)
    └── Role (e.g., "finance_staff", "admin")
            └── Permission[] (e.g., "finance:read", "finance:write")

User
    └── OrganizationMembership
            └── Role[] (within that org)
```

Permission strings follow the `resource:action` pattern:

| Resource | Actions |
|---|---|
| `user` | `read`, `write` |
| `org` | `read`, `write` |
| `finance` | `read`, `write` |
| `member` | `read`, `write` |
| `event` | `read`, `write` |
| `document` | `read`, `write` |
| `role` | `read`, `write` |
| `session` | `read`, `write` |

Wildcard permission `*:*` MAY be assigned to the Super Admin role only.

Roles are scoped to an organization. The same role slug (e.g., `admin`) may exist across multiple parishes independently. Role definitions are NOT shared across organizations.

### 7.5 Authorization Enforcement

Permission checks MUST be enforced via middleware at the API layer. Handler functions MUST NOT contain inline permission checks.

Middleware MUST:
1. Validate and parse the JWT access token.
2. Verify the `org_id` claim matches the requested resource's organization.
3. Check that the user's flattened `permissions` array contains the required permission for the endpoint.
4. Return `DOMUS-AUTH-003` (`403 Forbidden`) if the permission check fails.

### 7.6 Admin Approval Workflow

1. User registers → status set to `PENDING` → admin notified (in-app, future: email).
2. Parish Admin reviews pending users via `GET /orgs/{org_id}/members?status=pending`.
3. Admin approves: `PATCH /orgs/{org_id}/members/{user_id}/approve` + assigns role(s).
4. Admin rejects: `PATCH /orgs/{org_id}/members/{user_id}/reject`.
5. Admin kicks active user: `DELETE /orgs/{org_id}/members/{user_id}` → status `KICKED`, all sessions revoked.

---

## 8. Non-Functional Requirements

| Category | Requirement |
|---|---|
| **Security** | Refresh tokens MUST be stored as hashed values (bcrypt or SHA-256). Plain-text storage is prohibited. |
| **Security** | Access tokens MUST be signed with RS256 (asymmetric). HS256 is not permitted in production. |
| **Security** | All auth endpoints MUST be rate-limited to mitigate brute-force attacks. |
| **Isolation** | No API endpoint MAY return data across organization boundaries without explicit Super Admin privilege. |
| **Auditability** | All auth events (login, logout, token refresh, approval, kick) MUST be logged with `request_id`, `user_id`, `org_id`, and timestamp. |
| **Availability** | Token refresh MUST NOT require downtime during deployment. Key rotation MUST support overlap periods. |

---

## 9. API Endpoint Summary

All responses conform to **ADR-001** envelope contract.

| Method | Path | Permission Required | Description |
|---|---|---|---|
| `POST` | `/auth/oauth2/{provider}/redirect` | Public | Initiate OAuth2 flow |
| `GET` | `/auth/oauth2/{provider}/callback` | Public | OAuth2 callback, issue tokens |
| `POST` | `/auth/email/login` | Public (dev only) | Email/password login |
| `POST` | `/auth/refresh` | Public (valid refresh token) | Rotate refresh token, issue new access token |
| `POST` | `/auth/logout` | Authenticated | Revoke current session |
| `GET` | `/auth/sessions` | `session:read` | List active sessions for current user |
| `DELETE` | `/auth/sessions/{session_id}` | `session:write` | Revoke a specific session |
| `GET` | `/orgs/{org_id}/members` | `member:read` | List org members (filterable by status) |
| `PATCH` | `/orgs/{org_id}/members/{user_id}/approve` | `member:write` | Approve pending user + assign role |
| `PATCH` | `/orgs/{org_id}/members/{user_id}/reject` | `member:write` | Reject pending user |
| `DELETE` | `/orgs/{org_id}/members/{user_id}` | `member:write` | Kick active user |
| `GET` | `/orgs/{org_id}/roles` | `role:read` | List roles in org |
| `POST` | `/orgs/{org_id}/roles` | `role:write` | Create role |
| `PUT` | `/orgs/{org_id}/roles/{role_id}` | `role:write` | Update role + permissions |
| `DELETE` | `/orgs/{org_id}/roles/{role_id}` | `role:write` | Delete role |

---

## 10. Error Codes

| Code | HTTP Status | Description |
|---|---|---|
| `DOMUS-AUTH-001` | `401` | Missing or malformed access token |
| `DOMUS-AUTH-002` | `401` | Access token expired |
| `DOMUS-AUTH-003` | `403` | Insufficient permissions |
| `DOMUS-AUTH-004` | `401` | Refresh token invalid or expired |
| `DOMUS-AUTH-005` | `403` | User account is not active (PENDING, KICKED, REJECTED) |
| `DOMUS-AUTH-006` | `403` | Organization mismatch — token org does not match resource org |
| `DOMUS-AUTH-007` | `400` | OAuth2 provider not supported or disabled |
| `DOMUS-AUTH-008` | `409` | User already exists with this identity provider |

---

## 11. Out of Scope (Deferred to Future PRDs)

- Email notification on approval/rejection/kick events.
- Two-factor authentication (2FA / TOTP).
- SSO via SAML for enterprise diocese integrations.
- API key authentication for third-party parish integrations.
- Audit log viewer UI in the dashboard.

---

## 12. Related Documents

| Document | Description |
|---|---|
| ADR-001 | API Response Structure Contract |
| ADR-002 | Auth Architecture (JWT, session, middleware) — to be authored |
| ADR-003 | Multi-tenant & Organization Isolation — to be authored |
| ADR-004 | RBAC & Permission System — to be authored |

---

*Document maintained by the Domus Engineering Team. Review upon any breaking change to the authentication contract.*
