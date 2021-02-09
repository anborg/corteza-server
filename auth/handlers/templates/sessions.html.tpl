{{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Your sessions</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}

	<form method="POST" action="{{ links.Sessions }}">
	{{ .csrfField }}

	{{ range .sessions}}
		<div class="card mb-1 ">
			<div class="card-body {{ if .Current }}bg-secondary text-light{{ end }}">
				{{ if .Current }}
					<span class="badge badge-light float-right">Current session</span>
				{{ end }}

				<dt>
					Created on
				</dt>
				<dd>
					<time datetime="{{ .CreatedAt }}">{{ .CreatedAt | date "Mon, 02 Jan 2006 15:04:05 MST" }}</time>
				</dd>
				<dt>
					Expires at
				</dt>
				<dd>
					<time datetime="{{ .ExpiresAt }}">{{ .ExpiresAt | date "Mon, 02 Jan 2006 15:04:05 MST" }}</time>
					{{ if .Expired }}
					<span class="badge badge-danger">expired</span>
					{{ else if eq .ExpiresIn 0 }}
					<span class="badge badge-warning">today</span>
					{{ else if eq .ExpiresIn 1 }}
					<span class="badge badge-warning">in 1 day</span>
					{{ else }}
					<span class="badge badge-info">in {{ .ExpiresIn }} days</span>
					{{ end }}
				</dd>
				<dt>
					IP address
					{{ if .SameRemoteAddr}}
						<span class="badge badge-secondary">Same as current session</span>
					{{ end }}
				</dt>
				<dd>
					<code>{{ .RemoteAddr }}</code>
				</dd>
				<dt>
					Browser
					{{ if .SameUserAgent}}
						<span class="badge badge-secondary">Same as current session</span>
					{{ end }}
				</dt>
				<dd class="small">
					{{ .UserAgent }}
				</dd>

			{{ if not .Current }}
				<button
					type="submit"
					name="deleteSession"
					value="{{ .ID }}"
					class="btn btn-sm {{ if .Current }}btn-danger{{ else }}btn-warning{{ end }}"
				>
					Delete session
				</button>
			{{ end }}
			</div>
		</div>
	{{ end }}

		<div class="text-center">
			<button
				type="submit"
				name="delete-all-but-current"
				value="true"
				class="btn btn-sm btn-danger"
			>
				Delete all sessions but current
			</button>
		</div>

		<hr />

		<div class="text-center my-3">
			<a href="{{ links.Logout }}">Logout</a>
			<span>  |  </span>
			<a href="{{ links.Profile }}">Profile</a>
		</div>

	</form>
</div>
{{ template "inc_footer.html.tpl" }}
