# Contributing to Domus

Thanks for helping out. Read [AGENTS.md](AGENTS.md) first for architecture and conventions.

## Tech Stack

- **Backend:** Go, Fiber v3, Entgo, Cobra, Swag v2 (OpenAPI 3.1)
- **Frontend:** React 19, React Router v8, TypeScript, Tailwind CSS v4, Vite, TanStack Query
- **Tooling:** Bun, mise, Air, Husky, Commitlint
- **Deployment:** Docker, Helm (`./charts`)

## Setup

Prerequisite: [mise](https://mise.jdx.dev) (installs Go, Bun and other tools).

```bash
mise install
bun install
mise run dev        # API (live reload) + UI dev server
```

## Common Tasks

| Task                | Command          |
| :------------------ | :--------------- |
| Regenerate API spec | `mise run api`   |
| Regenerate mocks    | `mise run mock`  |
| Build               | `mise run build` |
| Test                | `mise run test`  |
| Format              | `mise run fmt`   |

## Project Layout

```txt
cmd/domus/        CLI entrypoint
ent/              Entgo schema and generated client
internal/         Go features (app, platform, <feature>/...)
api/              Generated OpenAPI spec (do not edit)
ui/               React Router SPA (embedded into the binary)
charts/           Helm chart
```

## Workflow

1. Never commit to `main`. Branch from it: `git switch -c <type>/<slug>` (e.g. `feat/task-crud`, `docs/readme`).
2. Keep changes small and focused. Make atomic commits: one topic and one scope per commit.
3. Keep history linear: rebase and fast-forward only, no merge commits.
4. Run the narrowest check that covers your change before opening a PR (see below).

## Commit Messages

Format: `<type>(<scope>): <subject>`, enforced by Commitlint (header and body lines ≤ 80 chars).

- **Scopes:** `api`, `ui`, `deps`, `release`, `tools`, `deploy`, `docs`
- **Types:** use `chore`, `ci`, `build`, `docs`, `refactor`, `test`, `style` by default. Use `feat`, `fix` or `perf` only when the change modifies shipped runtime code (`cmd/`, `internal/`, `ent/`, `ui/app/`) and should cut a release.

## Architecture Rules

- Domain units are **features** (never "modules").
- Go: `handler → service → repo → domain`. Other packages import only `internal/<feature>`, never its subpackages. DTOs live in `domain/dto.go`. Handlers stay thin.
- UI: feature code under `ui/app/features/<feature>/` must not import sibling features. Shared code goes in `ui/app/shared/`.
- After changing handler annotations or DTOs, run `mise run api`. Never hand-edit `api/swagger.json` or `ui/app/shared/api/schema.d.ts`.
- After changing mockable interfaces, run `mise run mock` and commit the generated files.

## Verification

| Scope                  | Command                |
| :--------------------- | :--------------------- |
| API QC                 | `mise run qc:api`      |
| UI QC                  | `mise run qc:ui`       |
| Everything             | `mise run qc`          |
| Go tests (race)        | `go test -race ./...`  |
| Helm chart             | `mise run lint:helm`   |

## Pull Requests

- Describe what changed and why.
- Add a focused test for non-trivial logic.
- Make sure QC passes.

## Security

Do not open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).

## License

By contributing, you agree your contributions are licensed under the [MIT License](LICENSE).
