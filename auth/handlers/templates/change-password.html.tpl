{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Change your password</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}
		<form
		method="POST"
		onsubmit="document.getElementById('submit').disabled=true"
		action="{{ links.ChangePassword }}">
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
				readonly
				placeholder="email@domain.ltd"
				value="{{ .user.Email }}"
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
				name="oldPassword"
				placeholder="Old password"
				aria-label="Old password">
		</div>
		<div class="input-group mb-3">
			<span class="input-group-text">
			  <i class="bi bi-key-fill"></i>
			</span>
			<input
				type="password"
				required
				class="form-control"
				name="newPassword"
				placeholder="New password"
				aria-label="New password">
		</div>
		<div class="text-right">
			<button class="btn btn-primary btn-block" type="submit">Change your password</button>
		</div>
	</form>
</div>
{{ template "inc_footer.html.tpl" }}
