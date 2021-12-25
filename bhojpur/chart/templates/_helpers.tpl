{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bhojpur.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified Bhojpur.NET application name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bhojpur.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
installation names
*/}}
{{- define "bhojpur.installation.longname" -}}{{- $bp := .bp -}}{{ $bp.installation.stage }}.{{ $bp.installation.tenant }}.{{ $bp.installation.region }}.{{ $bp.installation.cluster }}{{- end -}}
{{- define "bhojpur.installation.shortname" -}}{{- $bp := .bp -}}{{- if $bp.installation.shortname -}}{{ $bp.installation.shortname }}{{- else -}}{{ $bp.installation.region }}-{{ $bp.installation.cluster }}{{- end -}}{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bhojpur.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "bhojpur.container.imagePullPolicy" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $comp := .comp -}}
imagePullPolicy: {{ $comp.imagePullPolicy | default $bp.imagePullPolicy | default "IfNotPresent" }}
{{- end -}}

{{- define "bhojpur.container.resources" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $comp := .comp -}}
resources:
  requests:
    cpu: {{ if $comp.resources }} {{ $comp.resources.cpu | default $bp.resources.default.cpu }}{{ else }}{{ $bp.resources.default.cpu }}{{ end }}
    memory: {{ if $comp.resources }} {{ $comp.resources.memory | default $bp.resources.default.memory }}{{ else }}{{ $bp.resources.default.memory }}{{ end -}}
{{- end -}}

{{- define "bhojpur.container.ports" -}}
{{- $ := .root -}}
{{- $comp := .comp -}}
{{- if $comp.ports }}
ports:
{{- range $key, $val := $comp.ports }}
{{- if $val.containerPort }}
- name: {{ $key | lower }}
  containerPort: {{ $val.containerPort }}
{{- end -}}
{{- end }}
{{- end }}
{{- end -}}

{{- define "bhojpur.pod.dependsOn" -}}
{{- $ := .root -}}
{{- $comp := .comp -}}
{{- if $comp.dependsOn }}
{{- range $path := $comp.dependsOn }}
checksum/{{ $path }}: {{ include (print $.Template.BasePath "/" $path) $ | sha256sum }}
{{- end }}
{{- end }}
{{- end -}}

{{- define "bhojpur.pod.affinity" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $comp := .comp -}}
{{- if $comp.affinity -}}
affinity:
{{ $comp.affinity | toYaml | indent 2 }}
{{- else if $bp.affinity -}}
affinity:
{{ $bp.affinity | toYaml | indent 2 }}
{{- end -}}
{{- end -}}

{{- define "bhojpur.applicationAffinity" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $expr := dict -}}
{{- if $bp.components.application.affinity -}}
{{- if $bp.components.application.affinity.default -}}{{- $_ := set $expr $bp.components.application.affinity.default "" -}}{{- end -}}
{{- if $bp.components.application.affinity.prebuild -}}{{- $_ := set $expr $bp.components.application.affinity.prebuild "" -}}{{- end -}}
{{- if $bp.components.application.affinity.probe -}}{{- $_ := set $expr $bp.components.application.affinity.probe "" -}}{{- end -}}
{{- if $bp.components.application.affinity.regular -}}{{- $_ := set $expr $bp.components.application.affinity.regular "" -}}{{- end -}}
{{- end -}}
{{- /*
  In a previous iteration of the templates the node affinity was part of the workspace pod template.
  In that case we need to extract the affinity from the template and add it to the workspace affinity set.
*/ -}}
{{- if $bp.components.application.template -}}
{{- if $bp.components.application.template.spec -}}
{{- if $bp.components.application.template.spec.affinity -}}
{{- if $bp.components.application.template.spec.affinity.nodeAffinity -}}
{{- if $bp.components.application.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution -}}
{{- range $_, $t := $bp.components.application.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms -}}
{{- range $_, $m := $t.matchExpressions -}}
    {{- $_ := set $expr $m.key "" -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- if not (eq (len $expr) 0) -}}
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      {{- range $key, $val := $expr }}
      - matchExpressions:
        - key: {{ $key }}
          operator: Exists
      {{- end }}
{{- end -}}
{{- end -}}

{{- define "bhojpur.msgbusWaiter.container" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $this := dict "root" $ "bp" $bp "comp" $bp.components.serviceWaiter -}}
- name: msgbus-waiter
  image: {{ template "bhojpur.comp.imageFull" $this }}
  args:
  - -v
  - messagebus
  securityContext:
    privileged: false
    runAsUser: 31001
  env:
{{ include "bhojpur.container.messagebusEnv" . | indent 2 }}
{{- end -}}

{{- define "bhojpur.databaseWaiter.container" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $this := dict "root" $ "bp" $bp "comp" $bp.components.serviceWaiter -}}
- name: database-waiter
  image: {{ template "bhojpur.comp.imageFull" $this }}
  args:
  - -v
  - database
  securityContext:
    privileged: false
    runAsUser: 31001
  env:
{{ include "bhojpur.container.dbEnv" . | indent 2 }}
{{- end -}}

{{- define "bhojpur.container.defaultEnv" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
env:
- name: KUBE_STAGE
  value: "{{ $bp.installation.stage }}"
- name: KUBE_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: KUBE_DOMAIN
  value: "{{ $bp.installation.kubedomain | default "svc.cluster.local" }}"
- name: BHOJPUR_DOMAIN
  value: {{ $bp.hostname | quote }}
- name: HOST_URL
  value: "https://{{ $bp.hostname }}"
- name: BHOJPUR_REGION
  value: {{ $bp.installation.region | quote }}
- name: BHOJPUR_INSTALLATION_LONGNAME
  value: {{ template "bhojpur.installation.longname" . }}
- name: BHOJPUR_INSTALLATION_SHORTNAME
  value: {{ template "bhojpur.installation.shortname" . }}
- name: LOG_LEVEL
  value: {{ template "bhojpur.loglevel" . }}
{{- end -}}

{{- define "bhojpur.loglevel" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{ $bp.log.level | default "info" | lower | quote }}
{{- end -}}

{{- define "bhojpur.container.analyticsEnv" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- if $bp.analytics -}}
- name: BHOJPUR_ANALYTICS_WRITER
  value: {{ $bp.analytics.writer | quote }}
- name: BHOJPUR_ANALYTICS_SEGMENT_KEY
  value: {{ $bp.analytics.segmentKey | quote }}
{{- end }}
{{- end -}}

{{- define "bhojpur.container.dbEnv" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
- name: DB_HOST
  value: "{{ $bp.db.host }}"
- name: DB_USERNAME
  value: "{{ $bp.db.username }}"
- name: DB_PORT
  value: "{{ $bp.db.port }}"
- name: DB_PASSWORD
  value: "{{ $bp.db.password }}"
{{- if $bp.db.disableDeletedEntryGC }}
- name: DB_DELETED_ENTRIES_GC_ENABLED
  value: "false"
{{- end }}
- name: DB_ENCRYPTION_KEYS
{{- if $bp.dbEncryptionKeys.secretName }}
  valueFrom:
    secretKeyRef:
      name: {{ $bp.dbEncryptionKeys.secretName }}
      key: keys
{{- else }}
  value: {{ $.Files.Get $bp.dbEncryptionKeys.file | quote }}
{{- end -}}
{{- end -}}

{{- define "bhojpur.container.messagebusEnv" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
- name: MESSAGEBUS_USERNAME
  value: "{{ $bp.rabbitmq.auth.username }}"
- name: MESSAGEBUS_PASSWORD
  value: "{{ $bp.rabbitmq.auth.password }}"
- name: MESSAGEBUS_CA
  valueFrom:
    secretKeyRef:
        name: {{ $bp.rabbitmq.auth.tls.existingSecret | quote }}
        key: ca.crt
- name: MESSAGEBUS_CERT
  valueFrom:
    secretKeyRef:
        name: {{ $bp.rabbitmq.auth.tls.existingSecret | quote }}
        key: tls.crt
- name: MESSAGEBUS_KEY
  valueFrom:
    secretKeyRef:
        name: {{ $bp.rabbitmq.auth.tls.existingSecret | quote }}
        key: tls.key
{{- end -}}

{{- define "bhojpur.container.tracingEnv" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $comp := .comp -}}
{{- $tracing := $comp.tracing | default $bp.tracing -}}
{{- if $tracing }}
{{- if $tracing.endpoint }}
- name: JAEGER_ENDPOINT
  value: {{ $tracing.endpoint }}
{{- else }}
- name: JAEGER_AGENT_HOST
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
{{- end }}
- name: JAEGER_SAMPLER_TYPE
  value: {{ $tracing.samplerType }}
- name: JAEGER_SAMPLER_PARAM
  value: "{{ $tracing.samplerParam }}"
{{- end }}
{{- end -}}

{{- define "bhojpur.builtinRegistry.name" -}}
{{- if .Values.components.imageBuilder.registry.bypassProxy -}}
{{ include "bhojpur.builtinRegistry.internal_name" . }}
{{- else -}}
registry.{{ .Values.hostname }}
{{- end -}}
{{- end -}}

{{- define "bhojpur.builtinRegistry.internal_name" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{ index .Values "docker-registry" "fullnameOverride" }}.{{ .Release.Namespace }}.{{ $bp.installation.kubedomain | default "svc.cluster.local" }}
{{- end -}}

{{- define "bhojpur.comp.version" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $comp := .comp -}}
{{- required "please specify the Bhojpur.NET Platform version to use in your values.yaml or with the helm flag --set version=x.x.x" ($comp.version | default $bp.version) -}}
{{- end -}}

{{- define "bhojpur.comp.imageRepo" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $comp := .comp -}}
{{- $comp.imagePrefix | default $bp.imagePrefix -}}{{- $comp.imageName | default $comp.name -}}
{{- end -}}

{{- define "bhojpur.comp.imageFull" -}}
{{- $ := .root -}}
{{- $bp := .bp -}}
{{- $comp := .comp -}}
{{ template "bhojpur.comp.imageRepo" . }}:{{- template "bhojpur.comp.version" . -}}
{{- end -}}

{{- define "bhojpur.comp.configMap" -}}
{{- $comp := .comp -}}
{{ $comp.configMapName | default (printf "%s-config" $comp.name) }}
{{- end -}}

{{- define "bhojpur.pull-secret" -}}
{{- $ := .root -}}
{{- if (and .secret .secret.secretName .secret.path (not (eq ($.Files.Get .secret.path) ""))) -}}
{{- $name := .secret.secretName -}}
{{- $path := .secret.path -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}
  labels:
    app: {{ template "bhojpur.fullname" $ }}
    chart: "{{ $.Chart.Name }}-{{ $.Chart.Version }}"
    release: "{{ $.Release.Name }}"
    heritage: "{{ $.Release.Service }}"
  annotations:
    checksum/checksd-config: {{ $.Files.Get $path | b64enc | indent 2 | sha256sum }}
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: "{{ $.Files.Get $path | b64enc }}"
{{- end -}}
{{- end -}}

{{- define "bhojpur.remoteStorage.config" -}}
{{- $ := .root -}}
{{- $remoteStorageMinio := .remoteStorage.minio | default dict -}}
{{- $minio := $.Values.minio | default dict -}}
storage:
{{- if eq .remoteStorage.kind "minio" }}
  kind: minio
  blobQuota: {{ .remoteStorage.blobQuota | default 0 }}
  minio:
    endpoint: {{ $remoteStorageMinio.endpoint | default (printf "minio.%s" $.Values.hostname) }}
    accessKey: {{ required "minio access key is required, please add a value to your values.yaml or with the helm flag --set minio.accessKey=xxxxx" ($remoteStorageMinio.accessKey | default $minio.accessKey) }}
    secretKey: {{ required "minio secret key is required, please add a value to your values.yaml or with the helm flag --set minio.secretKey=xxxxx" ($remoteStorageMinio.secretKey | default $minio.secretKey) }}
    secure: {{ $remoteStorageMinio.secure | default ($minio.enabled | default false) }}
    region: {{ $remoteStorageMinio.region | default "local" }}
    parallelUpload: {{ $remoteStorageMinio.parallelUpload | default "" }}
{{- else }}
{{ toYaml .remoteStorage | indent 2 }}
{{- end -}}
{{- end -}}

{{- define "bhojpur.kube-rbac-proxy" -}}
- name: kube-rbac-proxy
  image: quay.io/brancz/kube-rbac-proxy:v0.11.0
  args:
  - --v=10
  - --logtostderr
  - --insecure-listen-address=[$(IP)]:9500
  - --upstream=http://127.0.0.1:9500/
  env:
  - name: IP
    valueFrom:
      fieldRef:
        fieldPath: status.podIP
  ports:
  - containerPort: 9500
    name: metrics
  resources:
    requests:
      cpu: 1m
      memory: 30Mi
  securityContext:
    runAsGroup: 65532
    runAsNonRoot: true
    runAsUser: 65532
  terminationMessagePolicy: FallbackToLogsOnError
{{- end -}}
