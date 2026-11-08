{{- define "mut.fullname" -}}
{{- printf "%s-mut" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "mut.labels" -}}
app.kubernetes.io/name: mut
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "mut.selectorLabels" -}}
app.kubernetes.io/name: mut
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "mut.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}

{{- define "mut.secretName" -}}
{{- if .Values.secrets.existingSecret -}}
{{- .Values.secrets.existingSecret -}}
{{- else -}}
{{- printf "%s-secrets" (include "mut.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "mut.dbSecretName" -}}
{{- if .Values.postgres.existingSecret -}}
{{- .Values.postgres.existingSecret -}}
{{- else -}}
{{- include "mut.secretName" . -}}
{{- end -}}
{{- end -}}

{{- define "mut.natsUrl" -}}
{{- if .Values.nats.enabled -}}
{{- printf "nats://%s-nats:4222" (include "mut.fullname" .) -}}
{{- else -}}
{{- .Values.nats.url -}}
{{- end -}}
{{- end -}}

{{- define "mut.env" -}}
- name: MUT_DATABASE_DRIVER
  value: postgres
- name: MUT_DATABASE_HOST
  value: {{ required "postgres.host is required" .Values.postgres.host | quote }}
- name: MUT_DATABASE_PORT
  value: {{ .Values.postgres.port | quote }}
- name: MUT_DATABASE_NAME
  value: {{ .Values.postgres.database | quote }}
- name: MUT_DATABASE_USERNAME
  value: {{ .Values.postgres.username | quote }}
- name: MUT_DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.dbSecretName" . }}
      key: {{ ternary .Values.postgres.existingSecretKey "MUT_DATABASE_PASSWORD" (ne .Values.postgres.existingSecret "") }}
- name: MUT_BROKER_PROVIDER
  value: nats
- name: MUT_BROKER_URL
  value: {{ include "mut.natsUrl" . | quote }}
- name: MUT_BROKER_QUEUE
  value: {{ .Values.nats.queue | quote }}
- name: MUT_BROKER_NAME
  value: mut
- name: MUT_SERVER_URL
  value: {{ .Values.server.publicUrl | quote }}
- name: MUT_SERVER_PORT
  value: {{ .Values.server.port | quote }}
- name: MUT_ENCRYPTION_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.secretName" . }}
      key: MUT_ENCRYPTION_KEY
- name: MUT_SERVER_PROM_METRICS_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.secretName" . }}
      key: MUT_SERVER_PROM_METRICS_SECRET
- name: MUT_AI_API_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.secretName" . }}
      key: MUT_AI_API_KEY
- name: MUT_RESEND_API_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.secretName" . }}
      key: MUT_RESEND_API_KEY
- name: MUT_OAUTH_GITHUB_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.secretName" . }}
      key: MUT_OAUTH_GITHUB_CLIENT_SECRET
- name: MUT_OAUTH_GOOGLE_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.secretName" . }}
      key: MUT_OAUTH_GOOGLE_CLIENT_SECRET
{{- range $key, $value := .Values.config }}
- name: {{ $key }}
  value: {{ $value | quote }}
{{- end }}
{{- end -}}
