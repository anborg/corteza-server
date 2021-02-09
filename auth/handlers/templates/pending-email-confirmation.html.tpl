{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Confirm your email</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}
	<div class="alert alert-primary" role="alert">
		You should receive email confirmation link to your inbox in a few moments.
	</div>
	<div class="text-center my-3">
		<a href="{{ links.RequestPasswordReset }}">Request again</a>
	</div>
	<div class="text-center my-3 small">
		<a href="{{ links.Signup }}">Create new account</a>
		|
		<a href="{{ links.Login }}">Login</a>
	</div>
</div>
{{ template "inc_footer.html.tpl" }}
