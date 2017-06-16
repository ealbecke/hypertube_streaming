package main

import (
	"net/http"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	ID					string
	Email				string
	Username			string
	Firstname			string
	Lastname			string
	Language_id			string
	Provider_id			string
	Creation_date		string
	Edit_date			string
	CommentsNumber		int
	MoviesSeenNumber	int
}

/* 
 * getUserHandler return user information
 */
func getUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {

			token, err := tokenParse(getAuthorizationToken(r.Header))
			if err != nil {
				log.Println("[Hypertube] - getUserHandler(): Token error: " + err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

				var user User
				var creationTimestamp, editTimestamp []uint8
				err = db.QueryRow("SELECT id, email, username, firstname, lastname, language_id, provider_id, creation_date, edit_date FROM users WHERE id = ?", claims["id"]).Scan(&user.ID, &user.Email, &user.Username, &user.Firstname, &user.Lastname, &user.Language_id, &user.Provider_id, &creationTimestamp, &editTimestamp)
				if err != nil {
					log.Println("[Hypertube] - getUserHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				err = db.QueryRow("SELECT COUNT(comments.id) FROM comments WHERE user_id = ?", claims["id"]).Scan(&user.CommentsNumber)
				if err != nil {
					log.Println("[Hypertube] - getUserHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				err = db.QueryRow("SELECT COUNT(medias_seen.user_id) FROM medias_seen WHERE user_id = ?", claims["id"]).Scan(&user.MoviesSeenNumber)
				if err != nil {
					log.Println("[Hypertube] - getUserHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				user.Creation_date = string(creationTimestamp)
				user.Edit_date = string(editTimestamp)

				writeJSON(w, user)

			} else {
				log.Println("[Hypertube] - getUserHandler(): Token error: " + err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			log.Println("[Hypertube] - getUserHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type profileInfos struct {
	UserID				string
	Username			string 
	Firstname			string
	Lastname			string
	CreationDate		string
	CommentsNumber		int
	MoviesSeenNumber	int
}

/* 
 * getProfileDetailsHandler return all the user profile informations
 */
func getProfileDetailsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {

			token, err := tokenParse(getAuthorizationToken(r.Header))
			if err != nil {
				log.Println("[Hypertube] - getProfileDetailsHandler(): Token error: " + err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

				if r.URL.Query().Get("id") != "" {

					var userInfos profileInfos
					err := db.QueryRow("SELECT users.id, users.username, users.firstname, users.lastname, users.creation_date FROM users WHERE users.id = ?", r.URL.Query().Get("id")).Scan(&userInfos.UserID, &userInfos.Username, &userInfos.Firstname, &userInfos.Lastname ,&userInfos.CreationDate)
					if err != nil {
						log.Println("[Hypertube] - getProfileDetailsHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					err = db.QueryRow("SELECT COUNT(comments.id) FROM comments WHERE user_id = ?", r.URL.Query().Get("id")).Scan(&userInfos.CommentsNumber)
					if err != nil {
						log.Println("[Hypertube] - getProfileDetailsHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					err = db.QueryRow("SELECT COUNT(medias_seen.user_id) FROM medias_seen WHERE user_id = ?", r.URL.Query().Get("id")).Scan(&userInfos.MoviesSeenNumber)
					if err != nil {
						log.Println("[Hypertube] - getProfileDetailsHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					writeJSON(w, userInfos)
					return
				} else {
					log.Println("[Hypertube] - getProfileDetailsHandler(): Bad request")
					w.WriteHeader(http.StatusBadRequest)
					return		
				}

			} else {
				log.Println("[Hypertube] - getProfileDetailsHandler(): Token error: " + err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			log.Println("[Hypertube] - getProfileDetailsHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

/*
*	getUserPreferedLanguage query database to return user's language in string
*/
func getUserPreferedLanguage(id string) (string, error) {
	var preferedLanguage string
	err := db.QueryRow("SELECT language FROM languages WHERE id = (SELECT language_id FROM users WHERE id = ?);", id).Scan(&preferedLanguage)
	if err != nil {
		return "", err
	}
	return preferedLanguage, nil
}