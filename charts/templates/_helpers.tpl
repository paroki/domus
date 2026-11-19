{{/*
Chart name, truncated to DNS label limit.
*/}}
{{- define "domus.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Fully qualified release name used as resource name prefix.
*/}}
{{- define "domus.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "domus.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "domus.labels" -}}
helm.sh/chart: {{ include "domus.chart" . }}
{{ include "domus.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: domus
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "domus.selectorLabels" -}}
app.kubernetes.io/name: {{ include "domus.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "domus.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "domus.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Container image reference.
*/}}
{{- define "domus.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag }}
{{- if .Values.global.imageRegistry }}
{{- printf "%s/%s:%s" .Values.global.imageRegistry .Values.image.repository $tag }}
{{- else }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}
{{- end }}

{{/*
Environment variables from values.config.
*/}}
{{- define "domus.env" -}}
{{- $c := .Values.config -}}
- name: APP_NAME
  value: {{ $c.appName | quote }}
- name: HOST
  value: {{ $c.host | quote }}
- name: PORT
  value: {{ .Values.port | quote }}
- name: LOG_LEVEL
  value: {{ $c.logLevel | quote }}
- name: READ_TIMEOUT
  value: {{ $c.readTimeout | quote }}
- name: WRITE_TIMEOUT
  value: {{ $c.writeTimeout | quote }}
- name: IDLE_TIMEOUT
  value: {{ $c.idleTimeout | quote }}
- name: SHUTDOWN_TIMEOUT
  value: {{ $c.shutdownTimeout | quote }}
- name: BODY_LIMIT
  value: {{ $c.bodyLimit | quote }}
- name: AUTO_MIGRATE
  value: {{ $c.autoMigrate | quote }}
- name: TRUST_PROXY
  value: {{ $c.trustProxy | quote }}
{{- with .Values.config.user }}
{{- with .superAdminEmail }}
- name: USER_SUPER_ADMIN_EMAIL
  value: {{ . | quote }}
{{- end }}
- name: USER_SUPER_ADMIN_NAME
  value: {{ .superAdminName | quote }}
- name: AUTH_OTP_TTL
  value: {{ .otpTtl | quote }}
- name: AUTH_OTP_MAX_ATTEMPTS
  value: {{ .otpMaxAttempts | quote }}
- name: AUTH_OTP_COOLDOWN
  value: {{ .otpCooldown | quote }}
- name: AUTH_SESSION_TTL
  value: {{ .sessionTtl | quote }}
- name: AUTH_SESSION_COOKIE_NAME
  value: {{ .cookieName | quote }}
- name: AUTH_SESSION_COOKIE_SECURE
  value: {{ .cookieSecure | quote }}
{{- end }}
{{- with .Values.existingSecret }}
{{- range $env, $key := dict "APP_DB_URL" "db-url" "AUTH_GOOGLE_ID" "google-id" "AUTH_GOOGLE_SECRET" "google-secret" "AUTH_GITHUB_ID" "github-id" "AUTH_GITHUB_SECRET" "github-secret" }}
- name: {{ $env }}
  valueFrom:
    secretKeyRef:
      name: {{ $.Values.existingSecret }}
      key: {{ $key }}
      optional: true
{{- end }}
{{- end }}
{{- with $c.sentry.dsn }}
- name: APP_SENTRY_DSN
  value: {{ . | quote }}
{{- end }}
- name: APP_SENTRY_ENVIRONMENT
  value: {{ $c.sentry.environment | quote }}
- name: APP_SENTRY_RELEASE
  value: {{ default .Chart.AppVersion $c.sentry.release | quote }}
- name: APP_SENTRY_TRACES_SAMPLE_RATE
  value: {{ $c.sentry.tracesSampleRate | quote }}
{{- with .Values.extraEnv }}
{{ toYaml . }}
{{- end }}
{{- end }}
