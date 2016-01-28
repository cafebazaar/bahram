package api

import (
	"net/http"

	grest "github.com/ant0ine/go-json-rest/rest"
	"github.com/cafebazaar/bahram/datasource"
	//	"github.com/cafebazaar/blacksmith/logging"
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
	w.WriteJson(user)
}

type GroupInList struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Public      bool   `json:"public"`
	Manager     string `json:"manager"`
	Joined      bool   `json:"joined"`
}

func (r *restServerAPI) ListGroups(w grest.ResponseWriter, req *grest.Request) {
	groups, err := r.ds.Groups()
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userEmail := req.Env["REMOTE_USER"].(string)

	var groupList []*GroupInList
	for _, g := range groups {
		gil := &GroupInList{
			Email:       g.Email,
			Name:        g.Name,
			Description: g.Description,
			Active:      g.Active,
			Public:      g.Public,
			Manager:     g.Manager,
			Joined:      g.IsMemeber(userEmail),
		}
		groupList = append(groupList, gil)
	}
	w.WriteJson(groupList)
}

func (r *restServerAPI) GetGroup(w grest.ResponseWriter, req *grest.Request) {
	g, err := r.ds.GroupByEmail(req.PathParam("email"))
	if err != nil {
		grest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(g)
}

func (r *restServerAPI) CreateGroup(w grest.ResponseWriter, req *grest.Request) {
	var g datasource.Group
	err := req.DecodeJsonPayload(&g)
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO More Validation
	_, err = r.ds.UserByEmail(g.Email)
	if err == nil {
		grest.Error(w, "A user with this email already exists", http.StatusBadRequest)
		return
	}

	_, err = r.ds.GroupByEmail(g.Email)
	if err == nil {
		grest.Error(w, "A group with this email already exists", http.StatusBadRequest)
		return
	}

	r.ds.StoreGroup(&g)
	w.WriteJson(g)
}

func (r *restServerAPI) UpdateGroup(w grest.ResponseWriter, req *grest.Request) {
	g, err := r.ds.GroupByEmail(req.PathParam("email"))
	if err != nil {
		grest.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userEmail := req.Env["REMOTE_USER"].(string)
	action := req.FormValue("action")

	switch action {
	case "join":
		if g.IsMemeber(userEmail) {
			grest.Error(w, "Already joined", http.StatusNotAcceptable)
			return
		}
		g.Members = append(g.Members, userEmail)
	case "leave":
		if g.Manager == userEmail {
			grest.Error(w, "You can't leave a group you which you manage", http.StatusNotAcceptable)
			return
		}
		leaved := false
		for i, m := range g.Members {
			if m == userEmail {
				leaved = true
				n := len(g.Members)
				g.Members[i] = g.Members[n-1]
				g.Members = g.Members[:n-1]
				break
			}
		}
		if !leaved {
			grest.Error(w, "Not joined anyway", http.StatusNotAcceptable)
			return
		}
	default:

	}

	r.ds.StoreGroup(g)
}
