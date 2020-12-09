{{/*
Helper function to get the proper image prefix
*/}}
{{- define "cray-fas.image-prefix" -}}
    {{ $base := index . "cray-service" }}
    {{- if $base.imagesHost -}}
        {{- printf "%s/" $base.imagesHost -}}
    {{- else -}}
        {{- printf "" -}}
    {{- end -}}
{{- end -}}

{{/*
Helper function to get the proper image tag
*/}}
{{- define "cray-fas.imageTag" -}}
{{- default "latest" .Chart.AppVersion -}}
{{- end -}}