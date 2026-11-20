# syntax=docker/dockerfile:1

FROM oven/bun:1-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package.json ui/bun.lock ./
RUN bun install --frozen-lockfile
COPY ui/ ./
# Public DSN is baked into the client bundle; auth token is a build secret
# used only for source map upload.
ARG APP_SENTRY_UI_DSN
ARG APP_SENTRY_ENVIRONMENT
ARG APP_SENTRY_RELEASE
RUN --mount=type=secret,id=sentry_auth_token \
    APP_SENTRY_AUTH_TOKEN="$(cat /run/secrets/sentry_auth_token 2>/dev/null)" \
    bun run build

FROM golang:1.27-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/ui/build/client ./ui/build/client
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /domus ./cmd/domus

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-builder /domus /domus
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/domus", "serve"]
