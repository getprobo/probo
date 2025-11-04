{{/*
Expand the name of the chart.
*/}}
{{- define "probo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "probo.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "probo.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "probo.labels" -}}
helm.sh/chart: {{ include "probo.chart" . }}
{{ include "probo.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "probo.selectorLabels" -}}
app.kubernetes.io/name: {{ include "probo.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "probo.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "probo.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
PostgreSQL hostname
*/}}
{{- define "probo.postgresql.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql-demo" .Release.Name }}
{{- else }}
{{- .Values.postgresql.host | required "postgresql.host is required when postgresql.enabled=false" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL port
*/}}
{{- define "probo.postgresql.port" -}}
{{- if .Values.postgresql.enabled }}
{{- 5432 }}
{{- else }}
{{- .Values.postgresql.port | default 5432 }}
{{- end }}
{{- end }}

{{/*
PostgreSQL database name
*/}}
{{- define "probo.postgresql.database" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.database | default "probod" }}
{{- else }}
{{- .Values.postgresql.database | default "probod" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL username
*/}}
{{- define "probo.postgresql.username" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.postgresUser | default "probod" }}
{{- else }}
{{- .Values.postgresql.username | default "probod" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL password (from subchart or external config)
*/}}
{{- define "probo.postgresql.password" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.postgresPassword | required "postgresql.auth.postgresPassword is required when postgresql.enabled=true" }}
{{- else }}
{{- .Values.postgresql.password | required "postgresql.password is required when postgresql.enabled=false" }}
{{- end }}
{{- end }}

{{/*
S3 endpoint
*/}}
{{- define "probo.s3.endpoint" -}}
{{- if .Values.minio.enabled }}
{{- printf "http://%s-minio:9000" .Release.Name }}
{{- else }}
{{- .Values.s3.endpoint }}
{{- end }}
{{- end }}

{{/*
S3 access key
*/}}
{{- define "probo.s3.accessKeyId" -}}
{{- if .Values.minio.enabled }}
{{- .Values.minio.auth.rootUser | required "minio.auth.rootUser is required when minio.enabled=true" }}
{{- else }}
{{- .Values.s3.accessKeyId | required "s3.accessKeyId is required when minio.enabled=false" }}
{{- end }}
{{- end }}

{{/*
S3 secret key
*/}}
{{- define "probo.s3.secretAccessKey" -}}
{{- if .Values.minio.enabled }}
{{- .Values.minio.auth.rootPassword | required "minio.auth.rootPassword is required when minio.enabled=true" }}
{{- else }}
{{- .Values.s3.secretAccessKey | required "s3.secretAccessKey is required when minio.enabled=false" }}
{{- end }}
{{- end }}

{{/*
Chrome DevTools Protocol address
*/}}
{{- define "probo.chrome.addr" -}}
{{- if .Values.chrome.enabled }}
{{- printf "%s-chrome:9222" (include "probo.fullname" .) }}
{{- else }}
{{- .Values.chrome.external.addr | required "chrome.external.addr is required when chrome.enabled=false" }}
{{- end }}
{{- end }}
