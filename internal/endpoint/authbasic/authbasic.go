package authbasic

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"s1767.xyz/idp/internal/endpoint/authcommon"
	"s1767.xyz/idp/internal/endpoint/utils"
	"s1767.xyz/idp/internal/storage/userdb"
)

func New(au authcommon.Authenticator, udb userdb.UserDb, contentDir string) (http.Handler, error) {
	// create the template
	b, err := os.ReadFile(
		filepath.Join(contentDir, "login.html"),
	)
	if err != nil {
		return nil, err
	}
	loginTemplate, err := template.New("login").Parse(string(b))
	if err != nil {
		return nil, err
	}

	ab := &authBasic{
		contentDir: contentDir,
		login:      loginTemplate,
		userdb:     udb,
		auth:       au,
	}

	return ab, nil
}

type authBasic struct {
	contentDir string
	login      *template.Template
	userdb     userdb.UserDb
	auth       authcommon.Authenticator
}

func (ab *authBasic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ab.authStart(w, r)
	} else if r.Method == "POST" {
		ab.authVerify(w, r)
	}
}

func (ab *authBasic) authStart(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Path + "?" + r.URL.RawQuery
	data := struct {
		Action string
	}{
		Action: action,
	}

	err := ab.login.Execute(w, data)
	if err != nil {
		log.Errorf("authbasic: failed to execute login template: %s", err)
	}
}

func (ab *authBasic) authVerify(w http.ResponseWriter, r *http.Request) {
	// check for required paramaters
	required := []string{
		"name",
		"password",
	}
	if utils.CheckParameters(r, required) == false {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("authbasic: missing params")
		return
	}

	// extract form parameters
	username := r.FormValue("name")
	password := r.FormValue("password")

	// verify the username password
	if ab.userdb.Verify(username, password) == false {
		w.WriteHeader(http.StatusUnauthorized)
		log.Errorf("authbasic: login failed")
		return
	}

	ab.auth.Authenticate(w, r, username, "")
}
