package api

import (
	"net/http"

	grest "github.com/ant0ine/go-json-rest/rest"
	"github.com/cafebazaar/bahram/datasource"
)

type restServerAPI struct {
	rest *grest.Api
	ds   datasource.DataSource
}

func newRestServerAPI(datasource datasource.DataSource) *restServerAPI {
	rest := grest.NewApi()
	rest.Use(grest.DefaultDevStack...)
	rest.Use(&grest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *grest.Request) bool {
			return true
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	// TODO: auth middleware

	return &restServerAPI{
		rest: rest,
		ds:   datasource,
	}
}

func (r *restServerAPI) MakeHandler() (http.Handler, error) {
	router, err := grest.MakeRouter(
		// Users
		grest.Get("/me", r.Me),
	)
	if err != nil {
		return nil, err
	}

	r.rest.SetApp(router)
	return r.rest.MakeHandler(), nil
}