package api

import (
	"net/http"

	grest "github.com/ant0ine/go-json-rest/rest"
)

// Returns information about current user
func (r *restServerAPI) Me(w grest.ResponseWriter, req *grest.Request) {
	//	values := map[string]string{
	//		"email":        "reza@cafebazaar.ir",
	//		"uid":          "reza",
	//		"inboxAddress": "remohammadi@gmail.com",
	//	}
	//	user, err := r.ds.CreateUser(true, values)
	user, err := r.ds.UserByEmail(req.Env["REMOTE_USER"].(string))
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(user.Info())
}
