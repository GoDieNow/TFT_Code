package main

import (
	"context"
	"errors"
	"strconv"

	"github.com/Nerzal/gocloak/v7"
	"github.com/prometheus/client_golang/prometheus"
	l "gitlab.com/cyclops-utilities/logging"
)

type basicUserData struct {
	UserSub string
}

type userInfo struct {
	Username      string `json:"username"`
	Firstname     string `json:"firstname"`
	Lastname      string `json:"lastname"`
	EmailAddress  string `json:"email"`
	EmailVerified bool   `json:"emailverified"`
	ID            string `json:"id"`
}

// AuthAPIKey (Swagger func) assures that the token provided when connecting to
// the API is the one presented in the config file as the valid token for the
// deployment.
func AuthAPIKey(token string) (i interface{}, e error) {

	l.Warning.Printf("[SECURITY] APIKey Authentication: Trying entry with token: %v\n", token)

	if !cfg.APIKey.Enabled {

		e = errors.New("the API Key authentication is not active in this deployment")

		metricSecurity.With(prometheus.Labels{"mode": "APIKey", "state": "DISABLED"}).Inc()

		return

	}

	if cfg.APIKey.Token == token {

		i = basicUserData{
			UserSub: token,
		}

		metricSecurity.With(prometheus.Labels{"mode": "APIKey", "state": "ACCESS GRANTED"}).Inc()

	} else {

		e = errors.New("provided token is not valid")

		metricSecurity.With(prometheus.Labels{"mode": "APIKey", "state": "INVALID TOKEN"}).Inc()

	}

	return

}

// AuthKeycloak (Swagger func) is called whenever it is necessary to check if a
// token which is presented to access an API is valid.
// The functions assumes that in keycloak the roles are going to be in uppercase
// and in the form of SCOPE_ROLE.
// Example:
//		ADMIN_ROLE <-> scope admin
func AuthKeycloak(token string, scopes []string) (i interface{}, returnErr error) {

	l.Debug.Printf("[SECURITY] AuthKeycloak: Performing authentication check - token = %v...%v\n", token, scopes)

	if !cfg.Keycloak.Enabled {

		returnErr = errors.New("the Keycloak authentication is not active in this deployment")

		metricSecurity.With(prometheus.Labels{"mode": "Keycloak", "state": "DISABLED"}).Inc()

		return

	}

	keycloakService := getKeycloakService(cfg.Keycloak)
	client := gocloak.NewClient(keycloakService)

	ctx := context.Background()

	_, err := client.LoginClient(ctx, cfg.Keycloak.ClientID, cfg.Keycloak.ClientSecret, cfg.Keycloak.Realm)
	// clientToken, err := client.LoginClient(ctx, cfg.Keycloak.ClientID, cfg.Keycloak.ClientSecret, cfg.Keycloak.Realm)

	if err != nil {

		l.Warning.Printf("[SECURITY] Problems logging in to keycloak. Error: %v\n", err.Error())

		returnErr = errors.New("unable to log in to keycloak")

		metricSecurity.With(prometheus.Labels{"mode": "Keycloak", "state": "FAIL: Keycloak Server unavailable"}).Inc()

		return

	}

	u, err := client.GetUserInfo(ctx, token, cfg.Keycloak.Realm)

	if err != nil {

		l.Warning.Printf("[SECURITY] Problems getting user Info. Error: %v\n", err.Error())

		returnErr = errors.New("unable to get userInfo from keycloak")

		metricSecurity.With(prometheus.Labels{"mode": "Keycloak", "state": "FAIL: Keycloak Server unavailable userInfo"}).Inc()

		return

	}

	if u != nil {

		i = basicUserData{
			UserSub: *u.Sub,
		}

		metricSecurity.With(prometheus.Labels{"mode": "Keycloak", "state": "ACCESS GRANTED"}).Inc()

	}

	// 	userRoles := make(map[string]bool)
	//
	// 	roles, err := client.GetRoleMappingByUserID(ctx, clientToken.AccessToken, cfg.Keycloak.Realm, *u.Sub)
	//
	// 	if err != nil {
	//
	// 		l.Warning.Printf("[SECURITY] Problems getting roles by user ID. Error:\n" + err.Error())
	//
	// 		returnErr = errors.New("unable to retrieve user roles from keycloak")
	//
	// 		return
	//
	// 	}
	//
	// 	for _, m := range roles.RealmMappings {
	//
	// 		userRoles[*m.Name] = true
	//
	// 	}
	//
	// 	ug, err := client.GetUserGroups(ctx, clientToken.AccessToken, cfg.Keycloak.Realm, *u.Sub)
	//
	// 	if err != nil {
	//
	// 		l.Warning.Printf("[SECURITY] Problems getting groups by user ID. Error:\n" + err.Error())
	//
	// 		returnErr = errors.New("unable to get user groups from keycloak")
	//
	// 		return
	//
	// 	}
	//
	// 	for _, m := range ug {
	//
	// 		roles, err := client.GetRoleMappingByGroupID(ctx, clientToken.AccessToken, cfg.Keycloak.Realm, *m.ID)
	//
	// 		if err != nil {
	//
	// 			l.Warning.Printf("[SECURITY] Problems getting roles by group ID. Error:\n" + err.Error())
	//
	// 			returnErr = errors.New("unable get groups roles from keycloak")
	//
	// 			return
	//
	// 		}
	//
	// 		for _, n := range roles.RealmMappings {
	//
	// 			userRoles[*n.Name] = true
	//
	// 		}
	//
	// 	}
	//
	// 	control := false
	//
	// 	for _, sc := range scopes {
	//
	// 		if userRoles[strings.ToUpper(sc)+"_ROLE"] {
	//
	// 			control = true
	//
	// 		}
	//
	// 	}
	//
	// 	if !control {
	//
	// 		returnErr = errors2.New(401, "The required role is not present in the user permissions")
	//
	// 		return
	//
	// 	}

	return

}

// getKeycloaktService returns the keycloak service; note that there has to be exceptional
// handling of port 80 and port 443
func getKeycloakService(c keycloakConfig) (s string) {

	if c.UseHTTP {

		s = "http://" + c.Host

		if c.Port != 80 {

			s = s + ":" + strconv.Itoa(c.Port)
		}

	} else {

		s = "https://" + c.Host

		if c.Port != 443 {

			s = s + ":" + strconv.Itoa(c.Port)

		}

	}

	return

}
