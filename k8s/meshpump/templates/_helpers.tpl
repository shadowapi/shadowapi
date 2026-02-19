{{/*
Expand the name of the chart.
*/}}
{{- define "meshpump.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "meshpump.fullname" -}}
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
{{- define "meshpump.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "meshpump.labels" -}}
helm.sh/chart: {{ include "meshpump.chart" . }}
{{ include "meshpump.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "meshpump.selectorLabels" -}}
app.kubernetes.io/name: {{ include "meshpump.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "meshpump.backendSelectorLabels" -}}
app.kubernetes.io/name: {{ include "meshpump.name" . }}-backend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
SSR selector labels
*/}}
{{- define "meshpump.ssrSelectorLabels" -}}
app.kubernetes.io/name: {{ include "meshpump.name" . }}-ssr
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
grpc2nats selector labels
*/}}
{{- define "meshpump.grpc2natsSelectorLabels" -}}
app.kubernetes.io/name: {{ include "meshpump.name" . }}-grpc2nats
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Worker selector labels
*/}}
{{- define "meshpump.workerSelectorLabels" -}}
app.kubernetes.io/name: {{ include "meshpump.name" . }}-worker
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "meshpump.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "meshpump.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
