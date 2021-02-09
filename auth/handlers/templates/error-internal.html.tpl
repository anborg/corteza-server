{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Internal error</h5>
	<div class="alert alert-danger" role="alert">
		{{ .error }}
	</div>
</div>
{{ template "inc_footer.html.tpl" }}
