{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Sign up</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}
	<form
		method="POST"
		onsubmit="document.getElementById('submit').disabled=true"
		action="{{ links.Signup }}">
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
					value="{{ if .user }}{{ .user.Email }}{{ end }}"
					aria-label="Email">
		</div>

		<div class="input-group mb-3">
		<span class="input-group-text">
			<i class="bi bi-key-fill"></i>
		</span>
			<input
					type="password"
					class="form-control"
					name="password"
					required
					placeholder="Password"
					aria-label="Password">
		</div>
		<div class="input-group mb-3">
		<span class="input-group-text">
		  <i class="bi bi-person-fill"></i>
		</span>
			<input
					type="text"
					class="form-control"
					name="name"
					placeholder="Your full name"
					value="{{ if .user }}{{ .user.Name }}{{ end }}"
					aria-label="Full name">
		</div>
		<div class="input-group mb-3">
		<span class="input-group-text">
			<i class="bi bi-emoji-smile"></i>
		</span>
			<input
					type="text"
					class="form-control"
					name="handle"
					placeholder="Short name, nickname or handle"
					value="{{ if .user }}{{ .user.Handle }}{{ end }}"
					aria-label="Handle">
		</div>
		<div class="text-right">
			<button
				id="submit"
				class="btn btn-primary btn-block"
				type="submit"
			>Submit</button>
		</div>
	</form>
	<div class="text-center my-3">Already have an account?
		<a href="{{ links.Login }}">Login</a>
	</div>
</div>
{{ template "inc_footer.html.tpl" }}
