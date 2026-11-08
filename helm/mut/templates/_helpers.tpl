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
{{- printf "%s:%s" .Values.image (.Values.tag | default .Chart.AppVersion) -}}
{{- end -}}

{{- define "mut.secretName" -}}
{{- if .Values.secrets.existingSecret -}}
{{- .Values.secrets.existingSecret -}}
{{- else -}}
{{- printf "%s-secrets" (include "mut.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "mut.dbSecretName" -}}
{{- if .Values.database.existingSecret -}}
{{- .Values.database.existingSecret -}}
{{- else -}}
{{- include "mut.secretName" . -}}
{{- end -}}
{{- end -}}

{{- define "mut.natsUrl" -}}
{{- if .Values.nats.url -}}
{{- .Values.nats.url -}}
{{- else -}}
{{- printf "nats://%s-nats:4222" (include "mut.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "mut.env" -}}
- name: MUT_DATABASE_DRIVER
  value: postgres
- name: MUT_DATABASE_HOST
  value: {{ required "database.host is required" .Values.database.host | quote }}
- name: MUT_DATABASE_PORT
  value: {{ .Values.database.port | quote }}
- name: MUT_DATABASE_NAME
  value: {{ .Values.database.name | quote }}
- name: MUT_DATABASE_USERNAME
  value: {{ .Values.database.user | quote }}
- name: MUT_DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.dbSecretName" . }}
      key: MUT_DATABASE_PASSWORD
- name: MUT_BROKER_PROVIDER
  value: nats
- name: MUT_BROKER_URL
  value: {{ include "mut.natsUrl" . | quote }}
- name: MUT_BROKER_QUEUE
  value: mut
- name: MUT_BROKER_NAME
  value: mut
- name: MUT_SERVER_URL
  value: {{ .Values.publicUrl | quote }}
- name: MUT_SERVER_PORT
  value: "8000"
- name: MUT_SERVER_MODE
  value: prod
- name: MUT_SERVER_LOG_LEVEL
  value: info
- name: MUT_SERVER_LOG_FORMAT
  value: json
{{- range $key, $value := .Values.secrets }}
{{- if ne $key "existingSecret" }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ include "mut.secretName" $ }}
      key: {{ $key }}
{{- end }}
{{- end }}
{{- end -}}
