package handlers

/*
	Hello! This file is auto-generated from `docs/src/spec.json`.

	For development:
	In order to update the generated files, edit this file under the location,
	add your struct fields, imports, API definitions and whatever you want, and:

	1. run [spec](https://github.com/titpetric/spec) in the same folder,
	2. run `./_gen.php` in this folder.

	You may edit `search.go`, `search.util.go` or `search_test.go` to
	implement your API calls, helper functions and tests. The file `search.go`
	is only generated the first time, and will not be overwritten if it exists.
*/

import (
	"context"
	"github.com/go-chi/chi"
	"net/http"

	"github.com/titpetric/factory/resputil"

	"github.com/crusttech/crust/sam/rest/request"
)

// Internal API interface
type SearchAPI interface {
	Messages(context.Context, *request.SearchMessages) (interface{}, error)
}

// HTTP API interface
type Search struct {
	Messages func(http.ResponseWriter, *http.Request)
}

func NewSearch(sh SearchAPI) *Search {
	return &Search{
		Messages: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewSearchMessages()
			resputil.JSON(w, params.Fill(r), func() (interface{}, error) {
				return sh.Messages(r.Context(), params)
			})
		},
	}
}

func (sh *Search) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Route("/search", func(r chi.Router) {
			r.Get("/messages", sh.Messages)
		})
	})
}