package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sort"
	"os/exec"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/oz/osdb"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type APIResponse struct {
	Status			string	`json:"status"`
	StatusMessage	string	`json:"status_message"`
	Data			APIData	`json:"data"`
}

type APIData struct {
	MovieCount	int			`json:"movie_count"`
	Limit		int			`json:"limit"`
	Medias		[]APIMedia	`json:"movies"`
	Media		APIMedia	`json:"movie"`
}

type APIMedia struct {
	Title			string			`json:"title"`
	Year			int				`json:"year"`
	Rating			float64			`json:"rating"`
	CoverURL		string			`json:"large_cover_image"`
	Summary			string			`json:"summary"`
	ID				int				`json:"id"`
	IMDBID			string			`json:"imdb_code"`
	Runtime			int				`json:"runtime"`
	BackgroundURL	string			`json:"background_image_original"`
	Torrents		[]APITorrent	`json:"torrents"`
	Comments		[]MediaComments	`json:"comments"`
	Show			bool
	Omdb			APIOMDB			`json:"omdbinfos"`
	AlreadySeen		bool			`json:"already_seen,omitempty"`
}

type APITorrent struct {
	Title	string	`json:"title"`
	Quality	string	`json:"quality,omitempty"`
	Seeds	int		`json:"seeds,omitempty"`
	Peers	int		`json:"peers,omitempty"`
	URL		string	`json:"url"`
	Size	string	`json:"size,omitempty"`
	Hash	string	`json:"hash,omitempty"`
}

type APIOMDB struct {
	Title		string	`json:"title"`
	Released	string	`json:"released"`
	Runtime		string	`json:"runtime"`
	Genre		string	`json:"genre"`
	Director	string	`json:"director"`
	Actors		string	`json:"actors"`
	Plot		string	`json:"plot"`
	ImdbRating	string	`json:"imdbRating"`
}

type Medias []APIMedia

func (m Medias) Len() int {
	return len(m)
}

func (m Medias) Less(i, j int) bool {
	return m[i].Title < m[j].Title
}

func (m Medias) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type ImdbID struct {
	ID string `json:"imdbid"`
}

/* 
 * createMediaHandler is inserting a media (id, imdb_id) in the database
 * everytime someone open a new media page that was never opened before.
 */
func createMediaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			var imdbid ImdbID
			err := json.NewDecoder(r.Body).Decode(&imdbid)
			if err != nil {
				log.Println("[Hypertube] - createMediaHandler(): Error decoding JSON: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if imdbid.ID != "" {
				var ID int64
				err := db.QueryRow("SELECT id FROM medias WHERE imdb_id = ?", imdbid.ID).Scan(&ID)
				if err == sql.ErrNoRows {
					stmt, err := db.Prepare("INSERT INTO medias(imdb_id) VALUES(?)")
					if err != nil {
						log.Println("[Hypertube] - createMediaHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					_, err = stmt.Exec(imdbid.ID)
					if err != nil {
						log.Println("[Hypertube] - createMediaHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
					return
				}
				if err != nil {
					log.Println("[Hypertube] - createMediaHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				return
			} else {
				log.Println("[Hypertube] - createMediaHandler(): Bad Request no imdbid received" + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

		} else {
			log.Println("[Hypertube] - createMediaHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

/* 
 * getMediaSubtitlesHandler find and download subtitles in the user 
 * prefered language
 */
func getMediaSubtitlesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {

			token, err := tokenParse(getAuthorizationToken(r.Header))
			if err != nil {
				log.Println("[Hypertube] - getMediaSubtitlesHandler(): Token error: " + err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

				imdbID := r.URL.Query().Get("imdb_id")
				isMovie := r.URL.Query().Get("movie")

				if imdbID != "" && isMovie == "true" {

					osdbClient, err := osdb.NewClient()
					if err != nil {
						log.Println("[Hypertube] - getMediaSubtitlesHandler(): Error creating OSDB client: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if err = osdbClient.LogIn("ealbecke", "coucou42", "eng"); err != nil {
						log.Println("[Hypertube] - getMediaSubtitlesHandler(): OSDB LogIn Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					preferedLanguage, err := getUserPreferedLanguage(strconv.FormatFloat(claims["id"].(float64), 'E', -1, 64))
					if err != nil {
						log.Println("[Hypertube] - getMediaSubtitlesHandler(): Error getting user prefered language: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					var languages []string
					if preferedLanguage == "French" {
						languages = append(languages, "fre")
					} else {
						languages = append(languages, "eng")
					}

					res, err := osdbClient.IMDBSearchByID([]string{strings.TrimPrefix(imdbID, "tt")}, languages)
					if err != nil {
						log.Println("[Hypertube] - getMediaSubtitlesHandler(): Error with osdbClient.IMDBSearchByID call: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					var subtitlesPath []string

					if len(res) > 0 {
						sub := res[0]
						if err := osdbClient.DownloadTo(&sub, "../app/public/subtitles/" + sub.SubFileName); err != nil {
							log.Println("[Hypertube] - getMediaSubtitlesHandler(): Error with osdbClient.DownloadTo call: " + err.Error())
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
						var args = []string{"-i", "../app/public/subtitles/" + sub.SubFileName, "-c:s", "webvtt", "../app/public/subtitles/" + sub.SubFileName + ".vtt"}
						if err := exec.Command("./ffmpeg", args...).Run(); err != nil {
							//log.Println("[Hypertube] - getMediaSubtitlesHandler(): Error with executing ffmpeg call: " + err.Error())
							//w.WriteHeader(http.StatusInternalServerError)
							//return
						}
						subtitlesPath = append(subtitlesPath, "/subtitles/" + sub.SubFileName + ".vtt")
					}

					if err = osdbClient.LogOut(); err != nil {
						log.Println("[Hypertube] - getMediaSubtitlesHandler(): OSDB LogOut Error: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					writeJSON(w, subtitlesPath)

				} else {
					log.Println("[Hypertube] - getMediaSubtitlesHandler(): Missing imdbID query in URL")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			} else {
				log.Println("[Hypertube] - getMediaSubtitlesHandler(): Token error")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			log.Println("[Hypertube] - getMediaSubtitlesHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

/* 
 * listMediaHandler list and sort all the medias found
 */
func listMediaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {

			page := r.URL.Query().Get("page")
			if page == "" {
				page = "1"
			}

			if r.URL.Query().Get("q") != "" {
				
				var list Medias

				movies, err := listQueryMovies(r.URL.Query().Get("q"), page)
				if err != nil {
					log.Println("[Hypertube] - listMediaHandler(): Error with listQueryMovies call: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				list = append(list, movies...)
				if page == "1" {
					shows, err := listQueryShow(r.URL.Query().Get("q"))
					if err != nil {
						log.Println("[Hypertube] - listMediaHandler(): Error with listQueryShow call: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					list = append(list, shows...)
				}
				list = sortByFilters(r.URL.Query(), list)
				sort.Sort(list)
				writeJSON(w, list)

			} else {

				token, err := tokenParse(getAuthorizationToken(r.Header))
				if err != nil {
					log.Println("[Hypertube] - listMediaHandler(): Token error: " + err.Error())
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

					var list Medias

					movies, err := listMovies(page)
					if err != nil {
						log.Println("[Hypertube] - listMediaHandler(): Error with listMovies call: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					shows, err := listShows(page)
					if err != nil {
						log.Println("[Hypertube] - listMediaHandler(): Error with listShows call: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					list = append(list, movies...)
					list = append(list, shows...)
					sort.Sort(list)

					rows, err := db.Query("SELECT imdb_id FROM medias_seen WHERE user_id = ?", claims["id"])
					if err != nil {
						log.Println("[Hypertube] - listMediaHandler(): SQL Error: " + err.Error())
						w.WriteHeader(http.StatusNotFound)
						return
					}

					var imdbsID []string
					for rows.Next() {
						var imdbID string
						err := rows.Scan(&imdbID)
						if err != nil {
							log.Println("[Hypertube] - listMediaHandler(): SQL Error: " + err.Error())
							w.WriteHeader(http.StatusNotFound)
							return
						}
						imdbsID = append(imdbsID, imdbID)
					}
					for listKey, listvalue := range list {
						for _, imdbsIDValue := range imdbsID {
							if listvalue.IMDBID == imdbsIDValue {
								list[listKey].AlreadySeen = true
							}
						}
					}

					writeJSON(w, list)
				}
			}
		} else {
			log.Println("[Hypertube] - listMediaHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

/* 
 * getMediaDetailsHandler gets all details on one media
 */
func getMediaDetailsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {

			showID := r.URL.Query().Get("show_id")
			movieID := r.URL.Query().Get("movie_id")
			imdbID := r.URL.Query().Get("imdbid")

			if (imdbID != "") {

				var ID int64
				err := db.QueryRow("SELECT id FROM medias WHERE imdb_id = ?", imdbID).Scan(&ID)
				if err != nil {
					log.Println("[Hypertube] - getMediaDetailsHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if showID != "" {
					media, err := getShowDetails(imdbID)
					if err != nil {
						log.Println("[Hypertube] - getMediaDetailsHandler(): Error with getShowDetails call: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					media.Comments, err = getComments(ID)
					if err != nil {
						log.Println("[Hypertube] - getMediaDetailsHandler(): Error with getComments call for show: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					writeJSON(w, media)
				} else if movieID != "" {
					media, err := getMovieDetails(movieID, imdbID)
					if err != nil {
						log.Println("[Hypertube] - getMediaDetailsHandler(): Error with getMovieDetails call: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					media.Comments, err = getComments(ID)
					if err != nil {
						log.Println("[Hypertube] - getMediaDetailsHandler(): Error with getComments call for movie: " + err.Error())
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					writeJSON(w, media)
				} else {
					log.Println("[Hypertube] - getMediaDetailsHandler(): Bad Request missing showID and movieID")
					w.WriteHeader(http.StatusBadRequest)
					return
				}

			} else {
				log.Println("[Hypertube] - getMediaDetailsHandler(): Bad Request missing imdbID")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			log.Println("[Hypertube] - getMediaDetailsHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type MediaComments struct {
	AuthorID	string
	Message		string
	Username	string
}

/* 
 * getComments gets all comments of one media
 */
func getComments(id int64) ([]MediaComments, error) {

	rows, err := db.Query("SELECT user_id, username, comment FROM comments WHERE media_id = ? ORDER BY creation_date DESC", id)
	if err != nil {
		return nil, err
	}

	var comments []MediaComments
	for rows.Next() {
		var comment MediaComments
		err := rows.Scan(&comment.AuthorID, &comment.Username, &comment.Message)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

type Comment struct {
	CommentMsg		string	`json:"comment"`
	ImdbId			string	`json:"mediaImdbId"`
}

func (c Comment) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.CommentMsg, validation.Required, is.ASCII, validation.Length(1, 200)),
		validation.Field(&c.ImdbId, validation.Required, is.Alphanumeric, validation.Length(1, 25)),
	)
}

/* 
 * createCommentHandler insert a new comment in the database
 */
func createCommentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			var comment Comment
			if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
				log.Println("[Hypertube] - createCommentHandler(): Error decoding JSON: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return 
			}

			err := comment.Validate()
			if err != nil {
				log.Println("[Hypertube] - createCommentHandler(): Error Bad Request Validation: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var mediaID int64
			err = db.QueryRow("SELECT id FROM medias WHERE imdb_id = ?", comment.ImdbId).Scan(&mediaID)
			if err != nil {
				log.Println("[Hypertube] - createCommentHandler(): SQL Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			token, err := tokenParse(getAuthorizationToken(r.Header))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			
				var userID int64
				var username string
				err = db.QueryRow("SELECT id, username FROM users WHERE id = ?", claims["id"]).Scan(&userID, &username)
				if err == sql.ErrNoRows {
					log.Println("[Hypertube] - createCommentHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				stmt, err := db.Prepare("INSERT INTO comments(media_id, user_id, username, comment) VALUES(?, ?, ?, ?)")
				if err != nil {
					log.Println("[Hypertube] - createCommentHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				_, err = stmt.Exec(mediaID, userID, username, comment.CommentMsg)
				if err != nil {
					log.Println("[Hypertube] - createCommentHandler(): SQL Error: " + err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				writeJSON(w, MediaComments{ strconv.FormatInt(userID, 10), comment.CommentMsg, username })
			}
		} else {
			log.Println("[Hypertube] - createCommentHandler(): Wrong http method call: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}