package main

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/minhdanh/thangmo-bot/internal"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("cmd", "thangmo-web")

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	config := config.NewConfig()
	templates := template.Must(template.ParseFiles("templates/index.html"))
	data := struct {
		TelegramChannel string
	}{
		strings.Replace(config.TelegramChannel, "@", "", -1),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := log.WithField("path", r.URL.Path).WithField("method", r.Method)
		l.Println(r.RemoteAddr + " " + r.UserAgent())
		if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			l.Println(err.Error())
		}
	})

	log.Println(http.ListenAndServe(":"+port, nil))
}
