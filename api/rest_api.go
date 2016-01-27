package api

import (
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"
)

// Returns information about current user
func (r *restServerAPI) Me(w rest.ResponseWriter, req *rest.Request) {
	//	values := map[string]string{
	//		"email":        "reza@cafebazaar.ir",
	//		"uid":          "reza",
	//		"inboxAddress": "remohammadi@gmail.com",
	//	}
	//	user, err := r.ds.CreateUser(true, values)
	user, err := r.ds.UserByEmail("reza@cafebazaar.ir")
	if err != nil {
		w.WriteJson(fmt.Sprintf(`{"error": %q}`, err))
	}
	w.WriteJson(user)
}
