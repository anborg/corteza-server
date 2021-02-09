{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Password reset requested</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}
	<div class="alert alert-primary" role="alert">
		If entered email is found in our database you should receive password reset link to your inbox in a few moments.
	</div>
	<div class="text-center my-3 small">
		<a href="{{ links.Signup }}">Create new account</a>
		|
		<a href="{{ links.Login }}">Login</a>
	</div>
</div>
{{ template "inc_footer.html.tpl" }}
