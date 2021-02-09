 {{ template "inc_header.html.tpl" }}
<div class="card-body">
	<h5 class="card-title">Authorize "{{ .client.Name }}"</h5>
	{{ template "inc_alerts.html.tpl" .alerts }}
	<form action="{{ links.OAuth2AuthorizeClient }}" method="POST">
	  {{ .csrfField }}
	  <p>
	  	Hello {{ coalesce .user.Name .user.Handle .user.Email }}.
	  </p>
	  <p>
	  	Application "{{ .client.Name }}" would like to perform actions on your behalf.
	  </p>
	  <p class="text-center">
		<button
		  type="submit"
		  name="deny"
		  class="btn btn-secondary btn-lg m-2"
		  style="width:180px;"
		>
		  Deny
		</button>
		<button
		  type="submit"
		  name="allow"
		  class="btn btn-primary btn-lg m-2"
		  style="width:180px;"
		>
		  Allow
		</button>
	  </p>

	  <hr />

	  This is a mistake, <a href="{{ links.Logout }}">logout</a>.

	</form>
</div>
 {{ template "inc_footer.html.tpl" }}
