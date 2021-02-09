{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Login</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}
	{{ if .settings.LocalEnabled }}
	<form
		method="POST"
		onsubmit="document.getElementById('submit').disabled=true"
		action="{{ links.Login }}">
		{{ .csrfField }}
		{{ if .form.Error }}
		<div class="alert alert-danger" role="alert">
			{{ .form.Error }}
		</div>
		{{ end }}
		<div class="input-group mb-3">
			<span class="input-group-text">
			  <i class="bi bi-envelope"></i>
			</span>
			<input
				type="email"
				class="form-control"
				name="email"
				required
				placeholder="email@domain.ltd"
				value="{{ if .form }}{{ .form.Email }}{{ end }}"
				aria-label="Email">
		</div>
		<div class="input-group mb-3">
			<span class="input-group-text">
			  <i class="bi bi-key-fill"></i>
			</span>
			<input
				type="password"
				required
				class="form-control"
				name="password"
				placeholder="Password"
				aria-label="Password">
		</div>
		<div class="form-check mb-3">
			<input type="checkbox" class="form-check-input" name="keep-session" id="keep-session" value="1">
			<label class="form-check-label" for="keep-session">Stay logged-in</label>
		</div>
		{{ if .settings.PasswordResetEnabled }}
		<div class="mb-3 text-right">
			<a href="{{ links.RequestPasswordReset }}" class="small">Forgot your password?</a>
		</div>
		{{ end }}
		<div class="text-right">
			<button class="btn btn-primary btn-block" type="submit">Log in</button>
		</div>
	</form>
	{{ if .settings.SignupEnabled }}
	<div class="text-center my-3">
		<a href="{{ links.Signup }}">Create new account</a>
	</div>
	{{ end }}
	<hr>
	{{ end }}
	{{ if .settings.FederatedEnabled }}
	<div>
	{{ range .providers }}
		<a href="{{ links.Federated }}/{{ .Handle }}" class="btn btn-outline-dark btn-block text-left mb-2">
			<i class="bi bi-{{ .Icon }} mr-2"></i>
			<small>Login with {{ coalesce .Label .Handle }}</small>
		</a>
	{{ end }}
	</div>
	{{ end }}
</div>
{{ template "inc_footer.html.tpl" }}
