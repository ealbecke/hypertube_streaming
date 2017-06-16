package main

import (
	"net/http"
	"strings"
	"encoding/json"
	"database/sql"
	"time"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/crypto/bcrypt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/sessions"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

var mySigningKey = []byte("secret")

var store = sessions.NewCookieStore([]byte("something-very-secret"))

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}


/* 
 * tokenParse return the token object of the user
 */
func tokenParse(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})
}

/* 
 * getAuthorizationToken return the token bearer string
 */
func getAuthorizationToken(header http.Header) string {
	return strings.Trim(strings.TrimPrefix(header.Get("Authorization"), "Bearer"), " ")
}

type UserLogInForm struct {
	Username	string
	Password	string
}

func (u UserLogInForm) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Username, validation.Required, is.Alphanumeric, validation.Length(1, 25)),
		validation.Field(&u.Password, validation.Required, validation.Length(1, 25)),
	)
}

/* 
 * loginHandler
 */
func loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			var ulf UserLogInForm
			ulf.Username = r.FormValue("username")
			ulf.Password = r.FormValue("password")

			err := ulf.Validate()
			if err != nil {
				log.Println("[Hypertube] - loginHandler(): Error Bad Request Validation: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var ID int64
			err = db.QueryRow("SELECT id FROM users WHERE username = ? AND provider_id  = ?", ulf.Username, "1").Scan(&ID)
			if err == sql.ErrNoRows {
				log.Println("[Hypertube] - registerHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			}

			var hashedPassword string
			err = db.QueryRow("SELECT password FROM users WHERE username = ? AND provider_id = ?", ulf.Username, "1").Scan(&hashedPassword)
			if err == sql.ErrNoRows {
				log.Println("[Hypertube] - registerHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(ulf.Password)); err != nil {
				log.Println("[Hypertube] - registerHandler(): Error in bcrypt.CompareHashAndPassword() call: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			data := struct { Token string `json:"token"` }{ getToken(ID) }
			writeJSON(w, data)
		} else {
			log.Println("[Hypertube] - registerHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type UserSignUpForm struct {
	Email		string
	Username	string
	Firstname	string
	Lastname	string
	Password	string
}

func (u UserSignUpForm) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Email, validation.Required, is.Email, validation.Length(8, 200)),
		validation.Field(&u.Username, validation.Required, is.Alphanumeric, validation.Length(1, 25)),
		validation.Field(&u.Firstname, validation.Required, is.ASCII, validation.Length(1, 25)),
		validation.Field(&u.Lastname, validation.Required, is.ASCII, validation.Length(1, 25)),
		validation.Field(&u.Password, validation.Required, validation.Length(8, 50)),
	)
}

/* 
 * registerHandler
 */
func registerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseMultipartForm(0)
			
			var usf UserSignUpForm
			usf.Email = r.FormValue("email")
			usf.Username = r.FormValue("username")
			usf.Firstname = r.FormValue("firstname")
			usf.Lastname = r.FormValue("lastname")
			usf.Password = r.FormValue("password")

			err := usf.Validate()
			if err != nil {
				log.Println("[Hypertube] - registerHandler(): Error Bad Request Validation: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var ID int64
			err = db.QueryRow("SELECT id FROM users WHERE provider_id = ? AND username = ? OR email = ?", "1", usf.Username, usf.Email).Scan(&ID)
			if err == sql.ErrNoRows {

				hash, err := bcrypt.GenerateFromPassword([]byte(usf.Password), bcrypt.DefaultCost)
				if err != nil {
					log.Println("[Hypertube] - registerHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				stmt, err := db.Prepare("INSERT INTO users(email, username, firstname, lastname, password, language_id, provider_id) VALUES(?, ?, ?, ?, ?, ?, ?)")
				if err != nil {
					log.Println("[Hypertube] - registerHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				res, err := stmt.Exec(usf.Email, usf.Username, usf.Firstname, usf.Lastname, string(hash), "1", "1")
				if err != nil {
					log.Println("[Hypertube] - registerHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				ID, err = res.LastInsertId()
				if err != nil {
					log.Println("[Hypertube] - registerHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				saveFormFile("profilePicture", "../app/public/images/profiles/" + usf.Username + strconv.FormatInt(ID, 10), w, r)
				data := struct { Token string `json:"token"` }{ getToken(ID) }
				writeJSON(w, data)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			log.Println("[Hypertube] - registerHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type passwordChangeForm struct {
	Token		string	`json:"token"`
	Password	string	`json:"password"`
}

func (p passwordChangeForm) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Token, validation.Required),
		validation.Field(&p.Password, validation.Required, validation.Length(8, 50)),
	)
}

/* 
 * passwordChangeHandler
 */
func passwordChangeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var pc passwordChangeForm

			pc.Token = r.FormValue("token")
			pc.Password = r.FormValue("password")

			err := pc.Validate()
			if err != nil {
				log.Println("[Hypertube] - passwordChangeHandler(): Error Bad Request Validation: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var ID int64
			err = db.QueryRow("SELECT id FROM users WHERE token_mail = ?", pc.Token).Scan(&ID)
			if err == sql.ErrNoRows {
				log.Println("[Hypertube] - passwordChangeHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			}
			
			hash, err := bcrypt.GenerateFromPassword([]byte(pc.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("[Hypertube] - passwordChangeHandler(): Error in bcrypt.GenerateFromPassword() call: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			stmt, err := db.Prepare("UPDATE users SET password = ?, token_mail = \"NULL\" WHERE id = ?;")
			if err != nil {
				log.Println("[Hypertube] - passwordChangeHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = stmt.Exec(hash, ID)
			if err != nil {
				log.Println("[Hypertube] - passwordChangeHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			data := struct { Token string `json:"token"` }{ getToken(ID) }
			writeJSON(w, data)
		} else {
			log.Println("[Hypertube] - passwordChangeHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type recoverForm struct {
	Email	string	`json:"email"`
}

func (r recoverForm) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email, validation.Length(8, 200)),
	)
}

/* 
 * recoverHandler
 */
func recoverHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			var rf recoverForm
			rf.Email = r.FormValue("email")

			err := rf.Validate()
			if err != nil {
				log.Println("[Hypertube] - recoverHandler(): Error Bad Request Validation: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var ID int64
			err = db.QueryRow("SELECT id FROM users WHERE email = ? AND provider_id = 1", rf.Email).Scan(&ID)
			if err == sql.ErrNoRows {
				log.Println("[Hypertube] - recoverHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusNotFound)
				return
			}

			mailToken := strings.TrimSuffix(randToken(), "=")

			stmt, err := db.Prepare("UPDATE users SET token_mail = ? WHERE email = ? AND provider_id = 1;")
			if err != nil {
				log.Println("[Hypertube] - recoverHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = stmt.Exec(mailToken, rf.Email)
			if err != nil {
				log.Println("[Hypertube] - recoverHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			sendEmail(rf.Email, "To: " + rf.Email +"\r\n" + "Subject: Hypertube: Forget password!\r\n" + "\r\n" + "Click on the link to create a new password:\r\n" + "http://localhost:3000/passwordchange?token=" + mailToken)
			w.WriteHeader(http.StatusOK)

		} else {
			log.Println("[Hypertube] - recoverHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type UserEditForm struct {
	Email		string
	Username	string
	Firstname	string
	Lastname	string
	Language	string
	OldPassword	string
	NewPassword	string
}

func (u UserEditForm) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Email, validation.Required, is.Email, validation.Length(8, 200)),
		validation.Field(&u.Firstname, validation.Required, is.ASCII, validation.Length(1, 25)),
		validation.Field(&u.Lastname, validation.Required, is.ASCII, validation.Length(1, 25)),
		validation.Field(&u.OldPassword, validation.Length(8, 50)),
		validation.Field(&u.NewPassword, validation.Length(8, 50)),
	)
}

/* 
 * userEditHandler
 */
func userEditHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			token, err := tokenParse(getAuthorizationToken(r.Header))
			if err != nil {
				log.Println("[Hypertube] - userEditHandler(): Token error: " + err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

				r.ParseMultipartForm(0)

				var uef UserEditForm
				uef.Email = r.FormValue("email")
				uef.Firstname = r.FormValue("firstname")
				uef.Lastname = r.FormValue("lastname")
				uef.Language = r.FormValue("language")
				uef.OldPassword = r.FormValue("oldPassword")
				uef.NewPassword = r.FormValue("newPassword")

				err := uef.Validate()
				if err != nil {
					log.Println("[Hypertube] - userEditHandler(): Error Bad Request Validation: " + err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				var hashedPassword, username string
				var ID int64
				err = db.QueryRow("SELECT id, password, username FROM users WHERE id = ?", claims["id"]).Scan(&ID, &hashedPassword, &username)
				if err == sql.ErrNoRows {
					log.Println("[Hypertube] - userEditHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if uef.OldPassword != "" && uef.NewPassword != "" {

					if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(uef.OldPassword)); err != nil {
						log.Println("[Hypertube] - userEditHandler(): User Error with bcrypt.CompareHashAndPassword() call" + err.Error())
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(uef.NewPassword), bcrypt.DefaultCost)
					if err != nil {
						log.Println("[Hypertube] - userEditHandler(): Error with bcrypt.GenerateFromPassword() call" + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					stmt, err := db.Prepare("UPDATE users SET email = ?, firstname = ?, lastname = ?, password = ?, language_id = ? WHERE id = ?")
					if err != nil {
						log.Println("[Hypertube] - userEditHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					_, err = stmt.Exec(uef.Email, uef.Firstname, uef.Lastname, newHashedPassword, uef.Language, ID)
					if err != nil {
						log.Println("[Hypertube] - userEditHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					if _, _, err := r.FormFile("profilePicture"); err != http.ErrMissingFile {
						saveFormFile("profilePicture", "../app/public/images/profiles/" + username + strconv.FormatInt(ID, 10), w, r)
					}

					w.WriteHeader(http.StatusOK)

				} else if uef.OldPassword == "" && uef.NewPassword == "" {

					stmt, err := db.Prepare("UPDATE users SET email = ?, firstname = ?, lastname = ?, language_id = ? WHERE id = ?")
					if err != nil {
						log.Println("[Hypertube] - userEditHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					_, err = stmt.Exec(uef.Email, uef.Firstname, uef.Lastname, uef.Language, ID)
					if err != nil {
						log.Println("[Hypertube] - userEditHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					if _, _, err := r.FormFile("profilePicture"); err != http.ErrMissingFile {
						saveFormFile("profilePicture", "../app/public/images/profiles/" + username + strconv.FormatInt(ID, 10), w, r)
					}

					w.WriteHeader(http.StatusOK)
					
				} else {
					log.Println("[Hypertube] - userEditHandler(): Bad Request")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			} else {
				log.Println("[Hypertube] - userEditHandler(): Wrong http method call: " + r.Method)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		}
	}
}

/* 
 * OAuth - JWT
 */
var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})

/* 
 * getToken return a fresh token
 */
func getToken(id int64) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, _ := token.SignedString(mySigningKey)
	return tokenString
}

var googleConf = &oauth2.Config{
	ClientID:		"32794147297-n81jf8qb2ppjjlfa7qagbekuuveokjlt.apps.googleusercontent.com",
	ClientSecret:	"6Sop3n2n3F6rbvqw-ingx8n7",
	RedirectURL:	"http://localhost:9090/api/auth/google/callback",
	Scopes:	[]string{
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
	},
	Endpoint: google.Endpoint,
}

type googleUser struct {
	Sub			string	`json:"sub"`
	Name		string	`json:"name"`
	Username	string
	Firstname	string	`json:"given_name"`
	Lastname	string	`json:"family_name"`
	PictureURL	string	`json:"picture"`
	Email		string	`json:"email"`
	Gender		string	`json:"gender"`
	Locale		string	`json:"locale"`
}

func getGoogleLoginURL(state string) string {
	return googleConf.AuthCodeURL(state)
}

/* 
 * googleAuthHandler
 */
func googleAuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := randToken()
		session, err := store.Get(r, "googleState")
		if err != nil {
			log.Println("[Hypertube] - googleAuthHandler(): Error recovering google state in session" + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		session.Values["state"] = state
		session.Save(r, w)
		http.Redirect(w, r, getGoogleLoginURL(state), http.StatusFound)
	}
}

/* 
 * googleCallbackHandler
 */
func googleCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := store.Get(r, "googleState")
		if err != nil {
			log.Println("[Hypertube] - googleCallbackHandler(): Error recovering google state in session" + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		retrievedState := session.Values["state"]
		queryValues := r.URL.Query()
		if retrievedState != queryValues["state"][0] {
			log.Println("[Hypertube] - googleCallbackHandler(): Error recovering state from google url")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := googleConf.Exchange(oauth2.NoContext, queryValues["code"][0])
		if err != nil {
			log.Println("[Hypertube] - googleCallbackHandler(): Error with googleConf.Exchange() call: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		client := googleConf.Client(oauth2.NoContext, token)
		email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			log.Println("[Hypertube] - googleCallbackHandler(): Error with googleConf.Client() call: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer email.Body.Close()

		var gu googleUser
		err = json.NewDecoder(email.Body).Decode(&gu)
		if err != nil {
			log.Println("[Hypertube] - googleCallbackHandler(): Error decoding google user JSON: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var ID int64
		err = db.QueryRow("SELECT id FROM users WHERE email = ? AND provider_id = ?", gu.Email, "2").Scan(&ID)
		if err == sql.ErrNoRows {
			gu.Username = strings.Replace(gu.Name, " ", "", -1)
			stmt, err := db.Prepare("INSERT INTO users(email, username, firstname, lastname, password, language_id, provider_id) VALUES(?, ?, ?, ?, ?, ?, ?)")
			if err != nil {
				log.Println("[Hypertube] - googleCallbackHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			res, err := stmt.Exec(gu.Email, gu.Username, gu.Firstname, gu.Lastname, "NULL", "2", "2")
			if err != nil {
				log.Println("[Hypertube] - googleCallbackHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ID, err = res.LastInsertId()
			if err != nil {
				log.Println("[Hypertube] - googleCallbackHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = saveImageFromURL(gu.PictureURL, "../app/public/images/profiles/" + gu.Username + strconv.FormatInt(ID, 10))
			if err != nil {
				log.Println("[Hypertube] - googleCallbackHandler(): Error with saveImageFromURL() call: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "http://localhost:3000/login?token=" + getToken(ID), http.StatusFound)
	}
}

/* 
 * 42Auth
 */
var fortyTwoConf = &oauth2.Config{
	ClientID:		"aa4152b7edb811da73c9e32101f3400155a6873356b27e6def28b0e7378d73b3",
	ClientSecret:	"dc6ef4a6bba07e8f2c8ada80e40f92f5ef4ec8af71e8458f3c04f0038796aeec",
	RedirectURL:	"http://localhost:9090/api/auth/fortytwo/callback",
	Scopes:	[]string{
		"public",
	},
	Endpoint: oauth2.Endpoint{
		AuthURL:	"https://api.intra.42.fr/oauth/authorize",
		TokenURL:	"https://api.intra.42.fr/oauth/token",
	},
}

type fortyTwoUser struct {
	Name		string	`json:"displayname"`
	Username	string	`json:"login"`
	Firstname	string	`json:"first_name"`
	Lastname	string	`json:"last_name"`
	Profile		string	`json:"profile"`
	PictureURL	string	`json:"image_url"`
	Email		string	`json:"email"`
}

func getFortyTwoLoginURL(state string) string {
	return fortyTwoConf.AuthCodeURL(state)
}

/* 
 * fortyTwoAuthHandler
 */
func fortyTwoAuthHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		state := randToken()
		session, err := store.Get(r, "fortyTwoState")
		if err != nil {
			log.Println("[Hypertube] - fortyTwoAuthHandler(): Error recovering fortyTwo state in session" + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		session.Values["state"] = state
		session.Save(r, w)
		http.Redirect(w, r, getFortyTwoLoginURL(state), http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

/* 
 * fortyTwoCallbackHandler
 */
func fortyTwoCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		session, err := store.Get(r, "fortyTwoState")
		if err != nil {
			log.Println("[Hypertube] - fortyTwoCallbackHandler(): Error recovering fortyTwo state in session" + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		retrievedState := session.Values["state"]
		queryValues := r.URL.Query()
		if retrievedState != queryValues["state"][0] {
			log.Println("[Hypertube] - fortyTwoCallbackHandler(): Error recovering state from google url")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := fortyTwoConf.Exchange(oauth2.NoContext, queryValues["code"][0])
		if err != nil {
			log.Println("[Hypertube] - fortyTwoCallbackHandler(): Error with fortyTwoConf.Exchange() call: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		client := fortyTwoConf.Client(oauth2.NoContext, token)
		email, err := client.Get("https://api.intra.42.fr/v2/me")
		if err != nil {
			log.Println("[Hypertube] - fortyTwoCallbackHandler(): Error with fortyTwoConf.Client() call: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer email.Body.Close()

		var fu fortyTwoUser
		err = json.NewDecoder(email.Body).Decode(&fu)
		if err != nil {
			log.Println("[Hypertube] - fortyTwoCallbackHandler(): Error decoding fortyTwo user JSON: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var ID int64
		err = db.QueryRow("SELECT id FROM users WHERE email = ? AND provider_id = ?", fu.Email, "3").Scan(&ID)
		if err == sql.ErrNoRows {
			stmt, err := db.Prepare("INSERT INTO users(email, username, firstname, lastname, password, language_id, provider_id) VALUES(?, ?, ?, ?, ?, ?, ?)")
			if err != nil {
				log.Println("[Hypertube] - fortyTwoCallbackHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			res, err := stmt.Exec(fu.Email, fu.Username, fu.Firstname, fu.Lastname, "NULL", "2", "3")
			if err != nil {
				log.Println("[Hypertube] - fortyTwoCallbackHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ID, err = res.LastInsertId()
			if err != nil {
				log.Println("[Hypertube] - fortyTwoCallbackHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = saveImageFromURL(fu.PictureURL, "../app/public/images/profiles/" + fu.Username + strconv.FormatInt(ID, 10))
			if err != nil {
				log.Println("[Hypertube] - fortyTwoCallbackHandler(): Error with saveImageFromURL() call: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "http://localhost:3000/login?token=" + getToken(ID), http.StatusFound)
	}
}