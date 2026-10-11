# Domus

Digitalisasi manajemen paroki. Satu platform untuk melayani banyak keuskupan, dengan hierarki **keuskupan → paroki → lingkungan**.

Monorepo dikelola dengan [Bun workspaces](https://bun.sh/docs/install/workspaces) dan [Turborepo](https://turbo.build/repo).

## Struktur Repo

```
domus/
├── apps/
│   ├── auth/          # @domus/auth        - Service autentikasi (ElysiaJS + Better Auth)
│   ├── api/           # @domus/api         - Backend REST (Go, Fiber v3, Ent, Casbin, Swaggo)
│   └── dash/          # @domus/dash        - Dashboard web (React Router v8 SPA, Ant Design)
├── packages/
│   ├── better-auth/   # @domus/better-auth - Konfigurasi Better Auth + Drizzle (server & client)
│   ├── openapi/       # @domus/openapi     - Tipe TypeScript hasil generate dari spec OpenAPI `apps/api`
│   └── tsconfig/      # @domus/tsconfig    - tsconfig dasar bersama
├── e2e/               # @domus/e2e         - Test end-to-end (Cucumber + Playwright)
├── .devcontainer/     # Dev environment lengkap (tool + Postgres, Redis, RustFS, Kafka)
├── .github/           # CI, build image, release-please, renovate
├── BRANDING.md        # Panduan visual / design token dashboard
└── turbo.json
```

| Package | Deskripsi | Port |
| --- | --- | --- |
| `apps/auth` | Login, sesi, dan penerbit JWT (plugin `admin` + `jwt` Better Auth). Login sosial Google/GitHub opsional. | 8001 |
| `apps/api` | Backend REST. Memvalidasi JWT lewat JWKS dari `apps/auth`, ORM Ent, otorisasi Casbin, spec OpenAPI v3.1 via Swaggo. | 8002 |
| `apps/dash` | Dashboard SPA (`ssr: false`) bergaya launcher modul, dengan i18n (id/en). | 3001 |
| `packages/better-auth` | Instance Better Auth, skema Drizzle (schema Postgres `auth`), env tervalidasi (zod). | - |
| `packages/openapi` | `src/api.ts` hasil `openapi-typescript`, dipakai `openapi-fetch` di dashboard. | - |

## Arsitektur Singkat

```
 Browser ──► apps/dash ──(login)──► apps/auth ──► Postgres (schema `auth`)
     │                                  │
     │  Bearer JWT                      │ JWKS
     └────────────► apps/api ◄──────────┘
                       │
                       └──► Postgres (Ent)
```

- `apps/auth` menerbitkan JWT. `apps/api` memverifikasinya dengan JWKS, mencocokkan `iss`/`aud` dengan `AUTH_URL`, lalu membuat *snapshot* user lokal pada request pertama.
- Otorisasi di `apps/api` berbasis **membership** per scope (`system`, `diocese`, `parish`, `unit`) dengan role per modul (mis. `finance-writer`) dan `superadmin`.
- Dashboard memakai registry modul (`sys`, `web`, `sacra`, `fin`, `par`, `act`). Menambah modul = satu entri di `apps/dash/app/shared/modules/registry.ts`.

## Tech Stack

- **Backend:** Go, Fiber v3, Ent, Casbin, Swaggo
- **Auth:** ElysiaJS, Better Auth, Drizzle ORM
- **Frontend:** React 19, React Router v8 (SPA), Ant Design 6, Tailwind CSS 4, i18next
- **Data / infra (devcontainer & CI):** PostgreSQL 18, Redis 8, RustFS (S3-compatible), Apache Kafka (KRaft)
- **Tooling:** Bun, Turborepo, Biome, Husky + Commitlint, Air, Swag, Vitest, Cucumber, Playwright

> Redis, S3 (RustFS), dan Kafka saat ini baru disiapkan di devcontainer dan CI; belum ada kode aplikasi yang memakainya.

## Prasyarat

- [Bun](https://bun.sh) `1.4.2` dan Node.js `>= 22`
- [Go](https://go.dev) `1.27` (lihat `apps/api/go.mod`)
- Docker & Docker Compose (Postgres dan layanan pendukung)
- Tool Go: [`air`](https://github.com/air-verse/air) dan [`swag`](https://github.com/swaggo/swag) v2

> Cara paling gampang: buka di **devcontainer** (atau GitHub Codespaces). Semua tool, service, dan environment variable sudah disiapkan, dan `bun install` serta `bun db:push` jalan otomatis.

## Memulai

```bash
git clone https://github.com/kilip/domus.git
cd domus

bun install            # juga memasang git hook (husky)
bun run db:push        # sinkronkan skema Better Auth ke Postgres

bun run dev            # jalankan semua app (turbo)
```

Jalankan satu app saja:

```bash
bunx turbo run dev --filter=@domus/auth
bunx turbo run dev --filter=@domus/api
bunx turbo run dev --filter=@domus/dash
```

Setelah jalan: dashboard di <http://localhost:3001>, auth di <http://localhost:8001> (docs OpenAPI di `/openapi`), API di <http://localhost:8002>.

### Environment

Di devcontainer semua nilai di bawah sudah di-set lewat `.devcontainer/compose.yaml`. Di luar devcontainer, buat `.env` (di-ignore git; `apps/api` juga membaca `.env` lewat godotenv).

| Variable | Dipakai oleh | Keterangan |
| --- | --- | --- |
| `AUTH_URL` | auth, api, e2e | URL publik auth (juga `iss`/`aud` JWT) |
| `AUTH_HOST`, `AUTH_PORT` | auth | Opsional; default dari `AUTH_URL` (port 8001) |
| `AUTH_PATH` | auth | Base path Better Auth (dev: kosong) |
| `AUTH_DB_URL` | auth, better-auth | Koneksi Postgres untuk Better Auth |
| `AUTH_SECRET` | auth | Secret Better Auth (ganti di production) |
| `AUTH_TRUSTED_ORIGINS` | auth, api | Daftar origin CORS, pisahkan dengan koma |
| `AUTH_GOOGLE_ID` / `AUTH_GOOGLE_SECRET` | auth | Opsional, login Google |
| `AUTH_GITHUB_ID` / `AUTH_GITHUB_SECRET` | auth | Opsional, login GitHub |
| `AUTH_JWKS_URL` | api | Endpoint JWKS dari auth. **Belum di-set di compose/CI**, isi manual |
| `API_DB_URL` | api | Koneksi Postgres untuk API (`...?sslmode=disable` di lokal) |
| `API_PORT` | api | Default `8002` |
| `VITE_AUTH_URL`, `VITE_PUBLIC_URL` | dash | URL auth dan URL publik dashboard |
| `VITE_API_URL` | dash | Default `http://localhost:8002` |
| `REDIS_URL`, `S3_*`, `KAFKA_BROKERS` | - | Disiapkan untuk infra, belum dipakai kode |

## OpenAPI & Tipe

Spec OpenAPI v3.1 digenerate dari anotasi Swaggo di `apps/api`, lalu dikonversi jadi tipe TypeScript di `packages/openapi`:

```bash
bun run api:gen
# = turbo run oa:gen (swag init di apps/api)
#   + openapi-typescript apps/api/docs/swagger.json -> packages/openapi/src/api.ts
```

Commit hasilnya (`apps/api/docs/*` dan `packages/openapi/src/api.ts`) setiap kali anotasi atau model response berubah.

## Perintah Umum

```bash
bun run dev          # semua app, watch mode
bun run build        # build package/app yang punya script build
bun run typecheck    # typecheck TypeScript
bun run lint         # Biome check (lint + format)
bun run lint:fix     # Biome auto-fix
bun run test         # test lewat turbo (lihat catatan di bawah)
bun run e2e          # test end-to-end
bun run db:push      # sinkronkan skema Better Auth ke DB (dev)
```

### Test

- **API (Go):** `cd apps/api && go test ./...`. Test API tidak ikut `bun run test` karena `apps/api` tidak punya script `test`; CI menjalankannya langsung.
- **Better Auth:** `bun --filter @domus/better-auth test` (Vitest).
- **E2E:** `bun run e2e`. Runner otomatis menyalakan `auth`, `dash`, dan `api` bila belum jalan (matikan dengan `E2E_AUTO_START=false`). Tambahkan `HEADED=true` untuk melihat browser. Perlu `bunx playwright install chromium` sekali.

## Docker

Semua image dibangun dari root repo (context `.`):

```bash
docker build -f apps/api/Dockerfile  -t domus-api .
docker build -f apps/auth/Dockerfile -t domus-auth .
docker build -f apps/dash/Dockerfile -t domus-dash .
docker build -f Dockerfile.migrator  -t domus-migrator .   # job sekali jalan: db:push skema Better Auth
```

Workflow `build.yml` mem-push image ke `ghcr.io/<owner>/domus/<app>` (tag `nightly`/`sha-*` untuk push ke `main`, semver + `latest` untuk release).

## Kontribusi

1. Buat branch dari `main`.
2. Commit dengan format [Conventional Commits](https://www.conventionalcommits.org). Scope yang diizinkan: `auth`, `dash`, `api`, `e2e`, `deps`, `ci`, `release`, `tools` (divalidasi commitlint).
3. Hook pre-commit menjalankan `biome check --staged`; pastikan `bun run lint` dan test terkait lolos.
4. Buka pull request ke `main`. CI menjalankan test API, test Better Auth, dan E2E.

Rilis dan `CHANGELOG.md` dikelola otomatis oleh release-please dari commit di `main`.

Panduan untuk AI coding agent ada di [AGENTS.md](./AGENTS.md); panduan visual dashboard di [BRANDING.md](./BRANDING.md).

## Lisensi

[MIT](./LICENSE) © Paroki Developer
