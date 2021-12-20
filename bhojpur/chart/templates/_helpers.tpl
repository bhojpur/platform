{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bhojpur.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
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
{{- define "bhojpur.installation.longname" -}}{{- $gp := .gp -}}{{ $gp.installation.stage }}.{{ $gp.installation.tenant }}.{{ $gp.installation.region }}.{{ $gp.installation.cluster }}{{- end -}}
{{- define "bhojpur.installation.shortname" -}}{{- $gp := .gp -}}{{- if $gp.installation.shortname -}}{{ $gp.installation.shortname }}{{- else -}}{{ $gp.installation.region }}-{{ $gp.installation.cluster }}{{- end -}}{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bhojpur.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "bhojpur.container.imagePullPolicy" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{- $comp := .comp -}}
imagePullPolicy: {{ $comp.imagePullPolicy | default $gp.imagePullPolicy | default "IfNotPresent" }}
{{- end -}}

{{- define "bhojpur.container.resources" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{- $comp := .comp -}}
resources:
  requests:
    cpu: {{ if $comp.resources }} {{ $comp.resources.cpu | default $gp.resources.default.cpu }}{{ else }}{{ $gp.resources.default.cpu }}{{ end }}
    memory: {{ if $comp.resources }} {{ $comp.resources.memory | default $gp.resources.default.memory }}{{ else }}{{ $gp.resources.default.memory }}{{ end -}}
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
{{- $gp := .gp -}}
{{- $comp := .comp -}}
{{- if $comp.affinity -}}
affinity:
{{ $comp.affinity | toYaml | indent 2 }}
{{- else if $gp.affinity -}}
affinity:
{{ $gp.affinity | toYaml | indent 2 }}
{{- end -}}
{{- end -}}

{{- define "bhojpur.applicationAffinity" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{- $expr := dict -}}
{{- if $gp.components.application.affinity -}}
{{- if $gp.components.application.affinity.default -}}{{- $_ := set $expr $gp.components.application.affinity.default "" -}}{{- end -}}
{{- if $gp.components.application.affinity.prebuild -}}{{- $_ := set $expr $gp.components.application.affinity.prebuild "" -}}{{- end -}}
{{- if $gp.components.application.affinity.probe -}}{{- $_ := set $expr $gp.components.application.affinity.probe "" -}}{{- end -}}
{{- if $gp.components.application.affinity.regular -}}{{- $_ := set $expr $gp.components.application.affinity.regular "" -}}{{- end -}}
{{- end -}}
{{- /*
  In a previous iteration of the templates the node affinity was part of the workspace pod template.
  In that case we need to extract the affinity from the template and add it to the workspace affinity set.
*/ -}}
{{- if $gp.components.application.template -}}
{{- if $gp.components.application.template.spec -}}
{{- if $gp.components.application.template.spec.affinity -}}
{{- if $gp.components.application.template.spec.affinity.nodeAffinity -}}
{{- if $gp.components.application.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution -}}
{{- range $_, $t := $gp.components.application.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms -}}
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
{{- $gp := .gp -}}
{{- $this := dict "root" $ "gp" $gp "comp" $gp.components.serviceWaiter -}}
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
{{- $gp := .gp -}}
{{- $this := dict "root" $ "gp" $gp "comp" $gp.components.serviceWaiter -}}
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
{{- $gp := .gp -}}
env:
- name: KUBE_STAGE
  value: "{{ $gp.installation.stage }}"
- name: KUBE_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: KUBE_DOMAIN
  value: "{{ $gp.installation.kubedomain | default "svc.cluster.local" }}"
- name: BHOJPUR_DOMAIN
  value: {{ $gp.hostname | quote }}
- name: HOST_URL
  value: "https://{{ $gp.hostname }}"
- name: BHOJPUR_REGION
  value: {{ $gp.installation.region | quote }}
- name: BHOJPUR_INSTALLATION_LONGNAME
  value: {{ template "bhojpur.installation.longname" . }}
- name: BHOJPUR_INSTALLATION_SHORTNAME
  value: {{ template "bhojpur.installation.shortname" . }}
- name: LOG_LEVEL
  value: {{ template "bhojpur.loglevel" . }}
{{- end -}}

{{- define "bhojpur.loglevel" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{ $gp.log.level | default "info" | lower | quote }}
{{- end -}}

{{- define "bhojpur.container.analyticsEnv" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{- if $gp.analytics -}}
- name: BHOJPUR_ANALYTICS_WRITER
  value: {{ $gp.analytics.writer | quote }}
- name: BHOJPUR_ANALYTICS_SEGMENT_KEY
  value: {{ $gp.analytics.segmentKey | quote }}
{{- end }}
{{- end -}}

{{- define "bhojpur.container.dbEnv" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
- name: DB_HOST
  value: "{{ $gp.db.host }}"
- name: DB_USERNAME
  value: "{{ $gp.db.username }}"
- name: DB_PORT
  value: "{{ $gp.db.port }}"
- name: DB_PASSWORD
  value: "{{ $gp.db.password }}"
{{- if $gp.db.disableDeletedEntryGC }}
- name: DB_DELETED_ENTRIES_GC_ENABLED
  value: "false"
{{- end }}
- name: DB_ENCRYPTION_KEYS
{{- if $gp.dbEncryptionKeys.secretName }}
  valueFrom:
    secretKeyRef:
      name: {{ $gp.dbEncryptionKeys.secretName }}
      key: keys
{{- else }}
  value: {{ $.Files.Get $gp.dbEncryptionKeys.file | quote }}
{{- end -}}
{{- end -}}

{{- define "bhojpur.container.messagebusEnv" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
- name: MESSAGEBUS_USERNAME
  value: "{{ $gp.rabbitmq.auth.username }}"
- name: MESSAGEBUS_PASSWORD
  value: "{{ $gp.rabbitmq.auth.password }}"
- name: MESSAGEBUS_CA
  valueFrom:
    secretKeyRef:
        name: {{ $gp.rabbitmq.auth.tls.existingSecret | quote }}
        key: ca.crt
- name: MESSAGEBUS_CERT
  valueFrom:
    secretKeyRef:
        name: {{ $gp.rabbitmq.auth.tls.existingSecret | quote }}
        key: tls.crt
- name: MESSAGEBUS_KEY
  valueFrom:
    secretKeyRef:
        name: {{ $gp.rabbitmq.auth.tls.existingSecret | quote }}
        key: tls.key
{{- end -}}

{{- define "bhojpur.container.tracingEnv" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{- $comp := .comp -}}
{{- $tracing := $comp.tracing | default $gp.tracing -}}
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
{{- $gp := .gp -}}
{{ index .Values "docker-registry" "fullnameOverride" }}.{{ .Release.Namespace }}.{{ $gp.installation.kubedomain | default "svc.cluster.local" }}
{{- end -}}

{{- define "bhojpur.comp.version" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{- $comp := .comp -}}
{{- required "please specify the Bhojpur.NET Platform version to use in your values.yaml or with the helm flag --set version=x.x.x" ($comp.version | default $gp.version) -}}
{{- end -}}

{{- define "bhojpur.comp.imageRepo" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
{{- $comp := .comp -}}
{{- $comp.imagePrefix | default $gp.imagePrefix -}}{{- $comp.imageName | default $comp.name -}}
{{- end -}}

{{- define "bhojpur.comp.imageFull" -}}
{{- $ := .root -}}
{{- $gp := .gp -}}
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
