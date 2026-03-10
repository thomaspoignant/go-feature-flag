{{/*
Expand the name of the chart.
*/}}
{{- define "relay-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "relay-proxy.fullname" -}}
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
{{- define "relay-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "relay-proxy.labels" -}}
helm.sh/chart: {{ include "relay-proxy.chart" . }}
{{ include "relay-proxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "relay-proxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "relay-proxy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "relay-proxy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "relay-proxy.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Renders a value that contains template
Usage:
{{ include "common.tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $ ) }}
*/}}
{{- define "relay-proxy.render" -}}
{{- $value := typeIs "string" .value | ternary .value (.value | toYaml) }}
{{- if contains "{{" (toJson .value) }}
  {{- tpl $value .context }}
{{- else }}
    {{- $value }}
{{- end }}
{{- end -}}

{{/*
Extract monitoringPort from relayproxy.config if it exists
Supports both server.monitoringPort (new format) and monitoringPort (old format)
Prioritizes server.monitoringPort if both are present
*/}}
{{- define "relay-proxy.monitoringPort" -}}
{{- if .Values.relayproxy.config }}
{{- $config := .Values.relayproxy.config | toString }}
{{- $port := "" }}
{{- /* First check for server.monitoringPort (new format) */}}
{{- if contains "server.monitoringPort:" $config }}
{{- $serverPortMatch := regexFind "server\\.monitoringPort:\\s*(\\d+)" $config }}
{{- if $serverPortMatch }}
{{- /* Extract just the port number by removing everything before the digits */}}
{{- $port = regexReplaceAll ".*monitoringPort:\\s*" $serverPortMatch "" }}
{{- end }}
{{- end }}
{{- /* If not found, check for top-level monitoringPort (old format) */}}
{{- if not $port }}
{{- /* Replace server.monitoringPort temporarily to avoid matching it */}}
{{- $tempConfig := $config | replace "server.monitoringPort:" "SERVER_MONITORING_PORT_PLACEHOLDER:" }}
{{- if contains "monitoringPort:" $tempConfig }}
{{- $topLevelPortMatch := regexFind "monitoringPort:\\s*(\\d+)" $tempConfig }}
{{- if $topLevelPortMatch }}
{{- /* Extract just the port number by removing everything before the digits */}}
{{- $port = regexReplaceAll ".*monitoringPort:\\s*" $topLevelPortMatch "" }}
{{- end }}
{{- end }}
{{- end }}
{{- if $port }}
{{- $port | trim }}
{{- end }}
{{- end }}
{{- end -}}
