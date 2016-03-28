package api

import (
	"fmt"
	"net/http"

	grest "github.com/ant0ine/go-json-rest/rest"
	"github.com/cafebazaar/bahram/datasource"
)

///////////////
// Users //////

//func sendResetPassword()

// Returns information about current user
func (r *restServerAPI) Me(w grest.ResponseWriter, req *grest.Request) {
	w.WriteJson(req.Env["REMOTE_USER_OBJECT"])
}

type UserInList struct {
	Email       string `json:"email"`
	UIDStr      string `json:"uid"`
	InboxAddr   string `json:"inboxAddress"`
	Active      bool   `json:"active"`
	Admin       bool   `json:"admin"`
	EnFirstName string `json:"enFirstName"`
	EnLastName  string `json:"enLastName"`
	FaFirstName string `json:"faFirstName"`
	FaLastName  string `json:"faLastName"`
}
type ChangePassword struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (r *restServerAPI) ListUsers(w grest.ResponseWriter, req *grest.Request) {
	currentUser := req.Env["REMOTE_USER_OBJECT"].(*datasource.User)
	if !currentUser.Admin {
		grest.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	users, err := r.ds.Users()
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userList []*UserInList
	for _, u := range users {
		uil := &UserInList{
			Email:       u.Email,
			UIDStr:      u.UIDStr,
			InboxAddr:   u.InboxAddr,
			Active:      u.Active,
			Admin:       u.Admin,
			EnFirstName: u.EnFirstName,
			EnLastName:  u.EnLastName,
			FaFirstName: u.FaFirstName,
			FaLastName:  u.FaLastName,
		}
		userList = append(userList, uil)
	}
	w.WriteJson(userList)
}

func (r *restServerAPI) GetUser(w grest.ResponseWriter, req *grest.Request) {
	currentUser := req.Env["REMOTE_USER_OBJECT"].(*datasource.User)

	email := req.PathParam("email")
	var user *datasource.User
	var err error
	if email == currentUser.Email {
		user = currentUser
	} else {
		if !currentUser.Admin {
			grest.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		user, err = r.ds.UserByEmail(email)
		if err != nil {
			grest.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}
	w.WriteJson(user)
}

func (r *restServerAPI) CreateUser(w grest.ResponseWriter, req *grest.Request) {
	currentUser := req.Env["REMOTE_USER_OBJECT"].(*datasource.User)
	if !currentUser.Admin {
		grest.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var u datasource.User
	err := req.DecodeJsonPayload(&u)
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO More Validation
	_, err = r.ds.UserByEmail(u.Email)
	if err == nil {
		grest.Error(w, "A user with this email already exists", http.StatusBadRequest)
		return
	}

	_, err = r.ds.GroupByEmail(u.Email)
	if err == nil {
		grest.Error(w, "A group with this email already exists", http.StatusBadRequest)
		return
	}

	u.SetPassword(u.Password, r.ds.ConfigByteArray("PASSWORD_SALT"))

	r.ds.StoreUser(&u)
	w.WriteJson(u)
}

func (r *restServerAPI) UpdateUser(w grest.ResponseWriter, req *grest.Request) {
	currentUser := req.Env["REMOTE_USER_OBJECT"].(*datasource.User)

	email := req.PathParam("email")
	var user *datasource.User
	var err error
	if email == currentUser.Email {
		user = currentUser
	} else {
		if !currentUser.Admin {
			grest.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		user, err = r.ds.UserByEmail(email)
		if err != nil {
			grest.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	action := req.FormValue("action")

	switch action {
	case "changePassword":
		salt := r.ds.ConfigByteArray("PASSWORD_SALT")
		var cp ChangePassword
		err = req.DecodeJsonPayload(&cp)
		if err != nil {
			grest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if user.HasPassword() && !user.AcceptsPassword(cp.OldPassword, salt) {
			grest.Error(w, "Old password is incorrect", http.StatusForbidden)
			return
		}
		user.SetPassword(cp.NewPassword, salt)
		// TODO notify

	case "update":
		var uTemp datasource.User
		err = req.DecodeJsonPayload(&uTemp)
		if err != nil {
			grest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user.UIDStr = uTemp.UIDStr
		user.InboxAddr = uTemp.InboxAddr // TODO notify the previous InboxAddr
		user.Active = uTemp.Active
		user.Admin = uTemp.Admin
		user.EnFirstName = uTemp.EnFirstName
		user.EnLastName = uTemp.EnLastName
		user.FaFirstName = uTemp.FaFirstName
		user.FaLastName = uTemp.FaLastName
		user.MobileNum = uTemp.MobileNum
		user.EmergencyNum = uTemp.EmergencyNum
		user.BirthDate = uTemp.BirthDate
		user.EnrolmentDate = uTemp.EnrolmentDate
		user.LeavingDate = uTemp.LeavingDate
	default:
		grest.Error(w, fmt.Sprintf("Unknown action: %s", action), http.StatusNotAcceptable)
		return
	}

	e := r.ds.StoreUser(user)
	if e != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		w.WriteJson(user)
	}
}

///////////////
// Groups /////

type GroupInList struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Public      bool   `json:"public"`
	Joinable    bool   `json:"joinable"`
	Manager     string `json:"manager"`
	Joined      bool   `json:"joined"`
}

func (r *restServerAPI) ListGroups(w grest.ResponseWriter, req *grest.Request) {
	groups, err := r.ds.Groups()
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := req.Env["REMOTE_USER_OBJECT"].(*datasource.User)

	var groupList []*GroupInList
	for _, g := range groups {
		if !user.Admin && !g.Joinable {
			continue
		}
		gil := &GroupInList{
			Email:       g.Email,
			Name:        g.Name,
			Description: g.Description,
			Active:      g.Active,
			Public:      g.Public,
			Manager:     g.Manager,
			Joined:      g.IsMemeber(user.Email),
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
	currentUser := req.Env["REMOTE_USER_OBJECT"].(*datasource.User)
	if !currentUser.Admin {
		grest.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// TODO check allowed domains

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

	user := req.Env["REMOTE_USER_OBJECT"].(*datasource.User)
	action := req.FormValue("action")

	switch action {
	case "join":
		if g.IsMemeber(user.Email) {
			grest.Error(w, "Already joined", http.StatusNotAcceptable)
			return
		}
		g.Members = append(g.Members, user.Email)
	case "leave":
		if g.Manager == user.Email {
			grest.Error(w, "You can't leave a group you which you manage", http.StatusNotAcceptable)
			return
		}
		leaved := false
		for i, m := range g.Members {
			if m == user.Email {
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
	case "update":
		if !user.Admin && g.Manager != user.Email {
			grest.Error(w, "You can't modify this group", http.StatusForbidden)
			return
		}
		var gTemp datasource.Group
		err := req.DecodeJsonPayload(&gTemp)
		if err != nil {
			grest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		g.Name = gTemp.Name
		g.Description = gTemp.Description
		g.Active = gTemp.Active
		g.Public = gTemp.Public
		g.Joinable = gTemp.Joinable
		g.Manager = gTemp.Manager
		g.CCs = gTemp.CCs

	default:
		grest.Error(w, fmt.Sprintf("Unknown action: %s", action), http.StatusNotAcceptable)
		return
	}

	e := r.ds.StoreGroup(g)
	if e != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		w.WriteJson(g)
	}
}
