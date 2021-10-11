package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"text/template"
	"time"

	"github.com/Gaardsholt/pass-along/config"
	"github.com/Gaardsholt/pass-along/secret"
	. "github.com/Gaardsholt/pass-along/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

var secretStore secret.SecretStore
var templates map[string]*template.Template
var startupTime time.Time

func init() {
	config.LoadConfig()

	startupTime = time.Now()
	secretStore = make(secret.SecretStore)

	templates = make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	templates["read"] = template.Must(template.ParseFiles("templates/base.html", "templates/read.html"))
}

func main() {

	r := mux.NewRouter()
	// Start of static stuff
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/js/").Handler(fs)
	r.PathPrefix("/css/").Handler(fs)
	r.PathPrefix("/robots.txt").Handler(fs)
	// End of static stuff

	r.HandleFunc("/", IndexHandler).Methods("GET")
	r.HandleFunc("/", NewHandler).Methods("POST")
	r.HandleFunc("/{id}", GetHandler).Methods("GET")

	port := 8080
	log.Info().Msgf("Starting server at port %d", port)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", port), r)).Msgf("Unable to start server at port %d", port)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templates["index"].Execute(w, Page{Startup: startupTime})
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	var entry Entry
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debug().Msg("Creating a new secret")

	myId, err := secretStore.Add(entry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", myId)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	useHtml := false
	ctHeader := r.Header.Get("Content-Type")
	contentType, _, err := mime.ParseMediaType(ctHeader)
	if err != nil || contentType != "application/json" {
		useHtml = true
	}

	if useHtml {
		_, hasData := secretStore[vars["id"]]
		if !hasData {
			w.WriteHeader(http.StatusGone)
		}
		newError := templates["read"].Execute(w, Page{Startup: startupTime})
		if newError != nil {
			fmt.Fprintf(w, "%s", newError)
		}
		return
	}

	secretData, gotData := secretStore.Get(vars["id"])
	if !gotData {
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "secret not found")
		return
	}

	log.Debug().Msg("Fetching a secret")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", secretData)
}
