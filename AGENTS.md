# AGENTS.md

Domus: Parish Administration Digitalized. Single-binary Go Modular Monolith serving an embedded React Router v7 SPA.

## Commands (via `mise`)

Run from repo root unless noted:

- **Dev (full stack)**: `mise run dev` (`air` on `:8080`, Vite on `:5173` proxying `/api` → `127.0.0.1:8080`)
- **Build (UI + Go)**: `mise run build` (builds `ui/build/client` first, then `bin/domus`)
- **Test & Typecheck**: `mise run test` (`go test ./...` + `cd ui && bun run typecheck`)
- **UI Unit Tests**: `cd ui && bun run test` (Vitest)
- **Go Checks (CI parity)**: `gofmt -l .` (must be empty), `go vet ./...`, `go test -v -race ./...`
- **Swagger Codegen**: `mise run generate` (`swag init -g internal/app/fiber.go -o api`)
- **Helm Check**: `helm lint charts --strict && helm template domus charts > /dev/null`

## Modular Monolith Architecture

### Backend (`cmd/`, `internal/`)
- **Runtime**: Go 1.27, Cobra CLI (`cmd/domus`), Fiber v3 (`internal/app`), `log/slog` JSON logger, `sethvargo/go-envconfig`.
- **Composition Root (`internal/app`)**: Wires config, logger, Fiber server, middleware, platform packages, and feature modules.
- **Shared Infrastructure (`internal/platform/<pkg>`)**:
  - `authx`: Authentication middleware, token/session verification.
  - `authz`: Authorization / RBAC enforcement.
  - `db/ent`: Central PostgreSQL + `ent` schemas and generated client.
- **Domain Features (`internal/<feature>`)**:
  - Current/planned features: `internal/identity` (credentials, sessions, OIDC), `internal/user` (user profile & domain entity), etc.
  - **Flat package layout**: Keep `handler.go`, `service.go`, `repository.go`, `model.go` flat inside `internal/<feature>` (no nested sub-packages until genuinely needed).
  - **Strict Isolation**: `internal/<feature>` MUST NOT import sibling `internal/<other-feature>` packages directly. Cross-feature calls use small consumer-defined interfaces inside the caller package, wired in `internal/app`.
  - **Database Ownership**: Central `ent` codegen lives in `internal/platform/db/ent`, but each `internal/<feature>` repository queries and mutates only its own domain entities.
  - **API Docs**: Annotate Fiber `/api/*` handlers with Swag v2 comments and run `mise run generate` when routes/DTOs change.

### Frontend (`ui/`)
- **Stack**: React 19, React Router v7 Framework Mode (`ssr: false` SPA in `ui/react-router.config.ts`), Tailwind CSS v4 (`@tailwindcss/vite`), Bun, Vitest.
- **React Router Skill (MANDATORY)**: Before modifying routes, loaders, actions, fetchers, or navigation, read `ui/.agents/skills/react-router/SKILL.md`, `ui/.agents/skills/react-router/references/framework-mode.md`, and version-matched docs in `ui/node_modules/react-router/docs/`.
- **Feature Modules (`ui/app/features/<feature>/`)**:
  - Mirror backend feature boundaries. Co-locate route views, components, `clientLoader`/`clientAction`, and `*.test.ts(x)` (Vitest) inside `ui/app/features/<feature>/`.
  - Every UI feature defines its own route config in `ui/app/features/<feature>/routes.ts`, composed into root `ui/app/routes.ts`.
- **Embedded SPA (`ui/embed.go`)**:
  - Go embeds `all:build/client` via `ui.DistFS()`.
  - Keep `ui/build/client/.keep` intact so `go test ./...` and `go build` succeed on clean checkouts before `bun run build`.

## Engineering Philosophy (Lazy Senior / Ponytail)

1. **Enforce the Ladder**: Stop at the first rung that holds:
   - Does it need to exist? (YAGNI)
   - Stdlib (`net/http`, `log/slog`, `errors.AsType`, `context`) or native platform feature (CSS, SQL/Ent constraint) covers it? Use it.
   - Already-installed dependency covers it? Use it; never add a package for a few lines of code.
   - Can it be one line? Make it one line.
2. **No Speculative Abstractions**: No 1-implementation interfaces (except cross-feature decoupling), no factories for 1 struct, no config env vars for constants.
3. **`ponytail:` Comments**: Mark deliberate simplifications in code with `// ponytail: <what was skipped>; <when/how to upgrade>` so future maintainers and agents know the ceiling and upgrade trigger.
4. **Trust Boundaries**: Never simplify away input validation at HTTP boundaries, error handling preventing data loss, security (non-root 65532, read-only rootfs in Helm/Docker), or accessibility.

## Git & Commit Rules

- **Branch Policy**: NEVER commit or push directly to `main`. Always work on an existing feature/fix branch or create a new branch (`git checkout -b <type>/<short-name>`) before committing or pushing.
- **Conventional Commits (Commitlint enforced)**: `<type>(<scope>): <subject>`
- **Scope is REQUIRED** (`scope-empty: never`). Allowed scopes in `commitlint.config.ts`:
  - Domain: `api`, `ui`, `db`, `auth`, `core`
  - Infra: `deps`, `release`, `tools`, `deploy`, `docs`
- **Line Limits**: Header ≤ 80 chars, body lines ≤ 80 chars.
- **Gitmoji Hook**: `.husky/commit-msg` runs `.husky/fix-emoji` to auto-prepend Gitmoji (`feat` → ✨, `fix` → 🐛, etc.) before `commitlint`. Write clean `<type>(<scope>): <description>` messages and keep total header length ≤ 78 chars to leave room for the emoji prefix.

## Verification Gates

Before finishing any task:
1. Run the narrowest test covering your change (`go test ./internal/...` or `cd ui && bun run test`).
2. Run `mise run test` (`go test ./...` + `react-router typegen && tsc`) and `gofmt -l .`.
3. If `/api` routes changed, run `mise run generate`.
4. If `charts/` changed, run `helm lint charts --strict`.
