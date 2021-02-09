{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Your profile</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}
	<div>
		<dt>Email:</dt>
		<dd>
			{{ .user.Email }}
		</dd>

		{{ if .emailConfirmationRequired }}
		<div class="alert alert-danger" role="alert">
			{{ .user.Email }} is not verified, <a href="{{ links.PendingEmailConfirmation }}?resend">resend confirmation link.</a>
		</div>
		{{ end }}



		{{ if .user.Name }}
		<dt>Full Name:</dt>
		<dd>{{ .user.Name }}</dd>
		{{ end }}
		{{ if .user.Name }}
		<dt>Handle:</dt>
		<dd>{{ .user.Handle }}</dd>
		{{ end }}
	</div>

	<hr />

	<div class="text-center my-3">
		<a href="{{ links.Logout }}">Logout</a>
		<span>  |  </span>
		<a href="{{ links.Sessions }}">Sessions</a>
		<span>  |  </span>
		<a href="{{ links.ChangePassword }}">Change your password</a>
	</div>
</div>
{{ template "inc_footer.html.tpl" }}
