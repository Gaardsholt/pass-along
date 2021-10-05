package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"text/template"

	"github.com/Gaardsholt/pass-along/secret"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

var secretStore secret.SecretStore
var templates map[string]*template.Template

func init() {
	secretStore = make(secret.SecretStore)

	templates = make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	templates["not_found"] = template.Must(template.ParseFiles("templates/base.html", "templates/not-found.html"))
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
	if err := templates["index"].Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	var entry Entry
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info().Msg("Creating a new secret")

	myId, err := secretStore.Add(entry.Content, entry.ExpiresIn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", myId)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	useHtml := false
	ctHdr := r.Header.Get("Content-Type")
	contentType, _, err := mime.ParseMediaType(ctHdr)
	if err != nil || contentType != "application/json" {
		useHtml = true
	}

	secretData, gotData := secretStore.Get(vars["id"])
	if !gotData {
		if err := templates["not_found"].Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	log.Info().Msg("Fetching a secret")

	if useHtml {
		if err := templates["read"].Execute(w, Entry{
			Content: secretData,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", secretData)
}

type Entry struct {
	Content   string `json:"content"`
	ExpiresIn int    `json:"expires_in"`
}
