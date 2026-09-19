package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/minhdanh/lococ/internal/config"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("cmd", "lococ-web")

func main() {
	config := config.NewConfig()

	port := strconv.Itoa(config.Port)
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

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
