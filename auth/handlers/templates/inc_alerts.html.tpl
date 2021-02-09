{{ range . }}
	<div class="alert alert-{{ .Type }}" role="alert">
		{{ .Text }}
	</div>
{{ end }}
