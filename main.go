package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"text/template"

	"github.com/Gaardsholt/share-a-password/secret"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

var secretStore secret.SecretStore
var indexTemplate, notFoundTemplate *template.Template

func init() {
	secretStore = make(secret.SecretStore)

	tmpl, err := template.ParseFiles("layout.html")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse template file for 'index'")
		os.Exit(1)
	}
	indexTemplate = tmpl

	tmpl, err = template.ParseFiles("not-found.html")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse template file for 'not-found'")
		os.Exit(1)
	}
	notFoundTemplate = tmpl
}

func main() {

	r := mux.NewRouter()
	// Start of static stuff
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/js/").Handler(fs)
	r.PathPrefix("/css/").Handler(fs)
	r.PathPrefix("/robots.txt").Handler(fs)
	// End of static stuff

	r.HandleFunc("/", NewHandler).Methods("POST")
	r.HandleFunc("/{id}", GetHandler).Methods("GET")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static/")))).Methods("GET")

	port := 8080
	log.Info().Msgf("Starting server at port %d", port)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", port), r)).Msgf("Unable to start server at port %d", port)
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	var entry Entry
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
		notFoundTemplate.Execute(w, nil)
		return
	}

	if useHtml {
		indexTemplate.Execute(w, Entry{
			Content: secretData,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", secretData)
}

type Entry struct {
	Content   string `json:"content"`
	ExpiresIn int    `json:"expires_in"`
}
