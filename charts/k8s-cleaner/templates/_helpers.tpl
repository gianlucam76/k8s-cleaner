{{/*
Expand the name of the chart.
*/}}
{{- define "k8s-cleaner.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "k8s-cleaner.fullname" -}}
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
{{- define "k8s-cleaner.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "k8s-cleaner.labels" -}}
helm.sh/chart: {{ include "k8s-cleaner.chart" . }}
{{ include "k8s-cleaner.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "k8s-cleaner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k8s-cleaner.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "k8s-cleaner.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "k8s-cleaner.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Resolve a controller port by name, falling back to a default value when absent.
*/}}
{{- define "k8s-cleaner.portByName" -}}
{{- $ports := .ports | default (list) -}}
{{- $port := .default | default 0 -}}
{{- range $ports }}
{{- if eq .name $.name }}
{{- $port = .containerPort -}}
{{- end }}
{{- end }}
{{- printf "%v" $port -}}
{{- end }}

{{/*
Render a value through the template engine. Expects a dict with "tpl" (the
value) and "ctx" (the root context).
*/}}
{{- define "k8s-cleaner.template" -}}
  {{- if typeIs "string" $.tpl }}
    {{- tpl $.tpl $.ctx | replace "+|" "\n" }}
  {{- else }}
    {{- tpl ($.tpl | toYaml) $.ctx | replace "+|" "\n" }}
  {{- end }}
{{- end -}}

{{/*
Render the flag name of a controller argument. The chart README documents the
naming rules. Leading dashes are normalised rather than kept, so the workaround
of writing the whole flag in the key keeps rendering as it did before the
values were rendered correctly.
*/}}
{{- define "k8s-cleaner.argFlag" -}}
{{- $flag := . | mustRegexFind "^[^_]+" -}}
{{- $name := regexReplaceAll "^-+" $flag "" -}}
{{- if or (hasPrefix "-" $flag) (gt (len $name) 1) -}}
{{- printf "--%s" $name -}}
{{- else -}}
{{- printf "-%s" $name -}}
{{- end -}}
{{- end -}}
