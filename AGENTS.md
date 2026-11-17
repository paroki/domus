# AGENTS.md

Domus: Parish Administration Digitalized. Single-binary Go modular monolith (Fiber v3, Ent, Swag) serving an embedded React Router v8 SPA (Framework Mode, `ssr: false`, TanStack Query + `openapi-fetch`).

## Git

- NEVER commit or push to `main`. Check `git branch --show-current` before editing; on `main`, run `git switch -c <type>/<slug>` first. Push branches only when the user says so.
- Commits follow `commitlint.config.ts`: `<type>(<scope>): <subject>`, header and body lines ≤ 80 chars. Scopes: `api`, `ui`, `deps`, `release`, `tools`, `deploy`, `docs`.
- Default to non-release types (`chore`, `ci`, `build`, `docs`, `refactor`, `test`, `style`). Use `feat`, `fix`, `perf`, `revert` only when the change modifies shipped runtime code (`cmd/`, `internal/`, `ent/`, `ui/app/`), because they cut a release (`.github/release-please.json`).
- Atomic commits: one `<type>(<scope>)` per commit, never `git add -A` across topics.

## Code Ownership

Domain units are called **features** (never "modules").

- `cmd/domus/`: CLI entry (`serve`, subcommands) and Swag root. Subcommands call feature root APIs only.
- `ent/schema/`: all Ent schemas. The `*ent.Client` is shared through `New` / `Register`.
- `internal/app/`: server lifecycle, config, Fiber wiring.
- `internal/platform/httpx/`: `{ data, meta, error }` envelope, sentinel errors, error handler.
- `internal/<feature>/`: Go feature.
- `ui/app/features/<feature>/`: UI feature. `ui/app/shared/`: cross-feature UI (theme, i18n, primitives, API client).
- `api/swagger.json` and `ui/app/shared/api/schema.d.ts`: generated, never hand-edit.

## Backend Features

Layout: `<feature>.go` (wiring only: `New`, `Register`), `domain/` (entities, sentinel errors, ALL DTOs), `service/` (ALL business logic), `handler/` (thin Fiber handlers and Swag annotations). Add `repo/` when storage needs it. Extra subpackages are allowed when the feature's PRD documents them.

- Feature and layer boundaries are enforced by depguard (`.golangci.yml`). Add a rule block when adding a feature.
- Cross-feature calls go through the root package's `api.go` (exported methods and type aliases). Dependency edges must be acyclic and are documented in the PRD that introduces them.
- Handlers parse the request, call a service, and map the result to the `httpx` envelope. Services use `domain` types only and receive the acting user id as an argument.
- Cross-feature atomic work passes a tx-scoped `*ent.Client` as an argument.
- The caller is read with `user.Current(c)` in handlers; guard roles with `user.RequireRole`.
- `/api` is deny-by-default: every route requires a session unless explicitly marked public.
- Never expose `ent` models in responses or Swag annotations. Every handler carries Swag v2 annotations; after changing handlers or DTOs run `mise run generate`.

## Configuration

- `APP_*` for global and platform concerns; `<FEATURE>_*` for one feature; `<FEATURE>_<DRIVER>_*` for a driver.
- A feature declares its variables in its own `config.go` (`env` tags, defaults in the tags). `internal/app.Config` embeds it with go-envconfig's `prefix`. A feature never reads another feature's variables.
- Required variables have no default and fail startup. Invalid values fail fast. No variable for a constant.
- Document every variable in the feature PRD and surface it in `charts/values.yaml`. Secrets come through `extraEnvFrom`, never `extraEnv`.

## Frontend

- Before changing routing, read `./ui/.agents/skills/react-router/SKILL.md` and its `references/framework-mode.md`.
- Never export `loader`, `action`, or `headers` (`ssr: false`).
- A feature MUST NOT import sibling `~/features/*`. Shared code, including route guards and session hooks, lives in `ui/app/shared/`. Cross-feature composition happens only in `ui/app/routes.ts` and `ui/app/apps.ts`.
- API reads and writes use `api` / `unwrap` from `~/shared/api/client` with TanStack Query, inside the feature's `api/`.
- User-facing strings go through `ui/app/shared/i18n` (id and en).

## Lazy Dev

Lazy means efficient, not careless. Stop at the first rung that holds: YAGNI, stdlib, native platform (CSS, Ent constraints, HTML attributes), existing dependency, one-liner, minimum working code.

- No speculative abstraction: no interface with one implementation, no factory for one product, no `repo/` wrapper without real domain complexity. Plain structs and functions.
- Mark deliberate simplifications with `// ponytail: <ceiling>; <upgrade path>`.
- Never simplify away validation at trust boundaries, error handling that prevents data loss, security, accessibility, or requested behavior.
- Leave one focused runnable check for non-trivial logic.

## Verification

Run the narrowest set that covers the change. CI runs the same checks.

- Go: `gofmt -l .`, `go vet ./...`, `golangci-lint run`, `govulncheck ./...`, `go test -race ./...`
- UI: `cd ui && bun run typecheck && bun test`
- OpenAPI: `mise run generate`
- Helm: `helm lint charts --strict && helm template domus charts > /dev/null`
