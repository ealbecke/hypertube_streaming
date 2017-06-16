package main

import (
	"database/sql"
	"syscall"
	"log"
	"fmt"
	"os"
	"os/signal"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
	"github.com/anacrolix/torrent"
	"github.com/jasonlvhit/gocron"
)

var db *sql.DB
var c *torrent.Client

func main() {
	log.Println("[Hypertube] - Starting API")

	var err error
	db, err = sql.Open("mysql", "root:@(127.0.0.1:3307)/Hypertube")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c, err = torrent.NewClient(&torrent.Config{
		DataDir:	"./downloads",
		NoUpload:	true,
		Seed:		false,
		DisableTCP:	false,
		ListenAddr:	fmt.Sprintf(":%d", 50007),
		Debug: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func(interruptChannel chan os.Signal) {
		for range interruptChannel {
			log.Println("[Hypertube] - Exiting API")
			c.Close()

			err := os.RemoveAll("./downloads")
			if err != nil {
				log.Println("[Hypertube] - main(): Error in os.RemoveAll() call: " + err.Error())
			}
			os.MkdirAll("./downloads", 0777)
			if err != nil {
				log.Println("[Hypertube] - main(): Error in os.MkdirAll() call: " + err.Error())
			}
			stmt, err := db.Prepare("DELETE FROM torrents")
			if err != nil {
				log.Println("[Hypertube] - main(): SQL Error: " + err.Error())
			}
			_, err = stmt.Exec()
			if err != nil {
				log.Println("[Hypertube] - main(): SQL Error: " + err.Error())
			}

			os.Exit(0)
		}
	}(interruptChannel)
	
	mux := http.NewServeMux()
	mux.Handle("/api/auth/register", registerHandler())
	mux.Handle("/api/auth/login", loginHandler())
	mux.Handle("/api/auth/recover", recoverHandler())
	mux.Handle("/api/auth/edit", userEditHandler())
	mux.Handle("/api/auth/passwordchange", passwordChangeHandler())
	mux.Handle("/api/auth/fortytwo", fortyTwoAuthHandler())
	mux.Handle("/api/auth/fortytwo/callback", fortyTwoCallbackHandler())
	mux.Handle("/api/auth/google", googleAuthHandler())
	mux.Handle("/api/auth/google/callback", googleCallbackHandler())
	mux.Handle("/api/user", jwtMiddleware.Handler(getUserHandler()))
	mux.Handle("/api/profile", jwtMiddleware.Handler(getProfileDetailsHandler()))
	mux.Handle("/api/create_message", jwtMiddleware.Handler(createCommentHandler()))
	mux.Handle("/api/create_media", jwtMiddleware.Handler(createMediaHandler()))
	mux.Handle("/api/get_list_medias", jwtMiddleware.Handler(listMediaHandler()))
	mux.Handle("/api/get_media_subtitles", jwtMiddleware.Handler(getMediaSubtitlesHandler()))
	mux.Handle("/api/get_media_details", jwtMiddleware.Handler(getMediaDetailsHandler()))
	mux.Handle("/api/get_media_stream", getMediaStreamHandler())
	mux.Handle("/api/start_download_media", jwtMiddleware.Handler(getDownloadMediaHandler()))

	s := gocron.NewScheduler()
	s.Every(1).Days().Do(removeOldMedias)
	s.Start()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})
	handler := c.Handler(mux)
	log.Fatal(http.ListenAndServe(":9090", handler))
}