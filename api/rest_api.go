package api

import (
	"github.com/ant0ine/go-json-rest/rest"
)

// Returns information about current user
func (r *restServerAPI) Me(w rest.ResponseWriter, req *rest.Request) {
	user := "Nothing"
	w.WriteJson(user)
}
