package api

import (
	"fmt"
	"net/http"
	"time"

	grest "github.com/ant0ine/go-json-rest/rest"
	"github.com/cafebazaar/bahram/datasource"
	"github.com/cafebazaar/blacksmith/logging"
	jwt "github.com/dgrijalva/jwt-go"
)

type restServerAPI struct {
	rest *grest.Api
	ds   *datasource.DataSource
}

func newRestServerAPI(datasource *datasource.DataSource) *restServerAPI {
	rest := grest.NewApi()
	rest.Use(grest.DefaultDevStack...)

	rest.Use(&grest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *grest.Request) bool {
			// TODO Origin check
			return true
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin", "Authorization"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	var bearerAuthMiddleware = &AuthBearerMiddleware{
		Realm: "RestAuthentication",
		Authenticator: func(token string) string {
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return datasource.ConfigByteArray("TOKEN_SIGN_KEY"), nil
			})

			if err == nil && parsedToken.Valid {
				return parsedToken.Claims["email"].(string)
			} else {
				return ""
			}
		},
		Authorizer: func(request *grest.Request, userID string) bool {
			user, err := datasource.UserByEmail(userID)
			if err != nil {
				logging.Log(debugTag, "Couldn't fetch user for userID=%s", userID)
				return false
			}
			request.Env["REMOTE_USER_OBJECT"] = user

			return true
		},
	}
	rest.Use(&grest.IfMiddleware{
		Condition: func(request *grest.Request) bool {
			return request.URL.Path != "/login"
		},
		IfTrue: bearerAuthMiddleware,
	})

	return &restServerAPI{
		rest: rest,
		ds:   datasource,
	}
}

func (r *restServerAPI) MakeHandler() (http.Handler, error) {
	router, err := grest.MakeRouter(
		// Auth
		grest.Post("/login", r.Login),
		// Users
		grest.Get("/me", r.Me),
		grest.Get("/users", r.ListUsers),
		grest.Get("/users/#email", r.GetUser),
		grest.Post("/users/#email", r.CreateUser),
		grest.Put("/users/#email", r.UpdateUser),
		// Groups
		grest.Get("/groups", r.ListGroups),
		grest.Get("/groups/#email", r.GetGroup),
		grest.Post("/groups/#email", r.CreateGroup),
		grest.Put("/groups/#email", r.UpdateGroup),
	)
	if err != nil {
		return nil, err
	}

	r.rest.SetApp(router)
	return r.rest.MakeHandler(), nil
}

type userPass struct {
	Email    string
	Password string
}

func (r *restServerAPI) Login(w grest.ResponseWriter, req *grest.Request) {
	up := userPass{}
	err := req.DecodeJsonPayload(&up)
	if err != nil {
		grest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if up.Email == "" || up.Password == "" {
		grest.Error(w, "user/password missing", http.StatusBadRequest)
	}

	user, err := r.ds.UserByEmail(up.Email)
	if err != nil {
		grest.Error(w, "user/password failed", http.StatusBadRequest)
		return
	}
	if !user.AcceptsPassword(up.Password, r.ds.ConfigByteArray("PASSWORD_SALT")) {
		grest.Error(w, "user/password failed", http.StatusBadRequest)
		return
	}

	if !user.Active {
		grest.Error(w, "user isn't activated", http.StatusForbidden)
		return
	}

	token := jwt.New(jwt.GetSigningMethod("HS512"))
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	token.Claims["email"] = user.Email
	token.Claims["isAdmin"] = user.Admin
	tokenString, err := token.SignedString(r.ds.ConfigByteArray("TOKEN_SIGN_KEY"))
	if err != nil {
		logging.Log(debugTag, "Signing failed: %s", err)
		grest.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}
	w.WriteJson(map[string]string{
		"token": tokenString,
	})
}
