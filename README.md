# Domus

Digitalisasi manajemen paroki. Satu platform untuk melayani banyak keuskupan, dengan hierarki **keuskupan → paroki → lingkungan**.

Monorepo dikelola dengan [Turborepo](https://turbo.build/repo).

## Struktur Repo

```
domus/
├── apps/
│   ├── auth/          # @omed/auth       - Auth utama (ElysiaJS + @omed/better-auth)
│   ├── api/           # @omed/api        - Backend (Go, Fiber v3, Ent, Casbin, Swaggo)
│   └── dash/          # @omed/dash       - Frontend (React Router v8 SPA, Ant Design)
├── packages/
│   ├── better-auth/   # @omed/better-auth - Better Auth reusable (server & client)
│   └── openapi/       # @omed/openapi     - Tipe TypeScript hasil generate dari OpenAPI v3
└── turbo.json
```

| Package | Deskripsi | Port |
| --- | --- | --- |
| `apps/auth` | Service autentikasi utama. Memakai `@omed/better-auth` dengan plugin organization & teams. | - |
| `apps/api` | Backend REST. ORM dengan Ent, otorisasi dengan Casbin, dokumentasi OpenAPI v3 via Swaggo. | 8001 |
| `apps/dash` | Dashboard web (SPA, `ssr: false`) dengan Ant Design. | 3001 |
| `packages/better-auth` | Konfigurasi Better Auth + Drizzle, dipakai bersama sisi server dan client. | - |
| `packages/openapi` | Tipe hasil `openapi-typescript` dari spec `apps/api`. | - |

## Tech Stack

- **Backend:** Go, Fiber v3, Ent, Casbin, Swaggo
- **Auth:** ElysiaJS, Better Auth, Drizzle
- **Frontend:** React Router v8 (SPA mode), Ant Design
- **Infra:** PostgreSQL, Redis, Apache Kafka (KRaft)
- **Tooling:** Bun, Turborepo, Biome, Air, Swag

## Konsep Utama

- **Multi-keuskupan.** Satu instalasi Domus melayani banyak keuskupan.
- **Tiga level admin:** admin keuskupan, admin paroki, admin lingkungan.
- **Super admin** membuat keuskupan baru lewat `apps/dash` (tidak ada pendaftaran mandiri).
- **Auth** ditangani `apps/auth` dengan organization & teams plugin dari Better Auth.
- **Hierarki** keuskupan/paroki/lingkungan disimpan di `apps/api`.

## Prasyarat

- [Bun](https://bun.sh)
- [Go](https://go.dev) (versi sesuai `apps/api/go.mod`)
- Docker & Docker Compose (untuk Postgres, Redis, Kafka)
- Tool Go: [`air`](https://github.com/air-verse/air) dan [`swag`](https://github.com/swaggo/swag)

> Cara paling gampang: pakai **devcontainer** yang sudah menyiapkan semua tool di atas beserta Postgres, Redis, dan Kafka sebagai service Docker Compose.

## Memulai

```bash
# 1. Clone
git clone <repo-url> domus
cd domus

# 2. Install dependency
bun install

# 3. Siapkan environment
cp .env.example .env   # lalu sesuaikan nilainya

# 4. Jalankan infra (kalau tidak pakai devcontainer)
docker compose up -d postgres redis kafka

# 5. Jalankan semua app
bun run dev
```

Jalankan satu app saja:

```bash
bunx turbo run dev --filter=@omed/dash
bunx turbo run dev --filter=@omed/api
bunx turbo run dev --filter=@omed/auth
```

## Generate OpenAPI & Tipe

Spec OpenAPI v3 digenerate dari anotasi Swaggo di `apps/api`, lalu dikonversi ke tipe TypeScript di `packages/openapi`:

```bash
# generate spec dari apps/api
cd apps/api && swag init

# generate tipe TypeScript
bunx turbo run generate --filter=@omed/openapi
```

## Perintah Umum

```bash
bun run dev        # jalankan semua app (watch mode)
bun run build      # build semua package/app
bun run lint       # lint & format check (Biome)
bun run test       # jalankan test
```

## Kontribusi

1. Buat branch dari `main`.
2. Pastikan `bun run lint` dan `bun run test` lolos.
3. Buka pull request dengan deskripsi perubahan yang jelas.

## Lisensi

[MIT](./LICENSE) © Paroki Developer