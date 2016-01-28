package api

import (
	"net/http"

	grest "github.com/ant0ine/go-json-rest/rest"
)

// Returns information about current user
func (r *restServerAPI) Me(w grest.ResponseWriter, req *grest.Request) {
	//	user, err := r.ds.CreateUser("reza@cafebazaar.ir", "reza", "remohammadi@gmail.com")
	//	if err == nil {
	//		user.SetActive(true)
	//		user.SetAdmin(true)
	//		user.SetPassword("testy")
	//		r.ds.StoreUser(user)
	//	}
	user, err := r.ds.UserByEmail(req.Env["REMOTE_USER"].(string))
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(user.Info())
}

func (r *restServerAPI) CreateGroup(w grest.ResponseWriter, req *grest.Request) {
}
