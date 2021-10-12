package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/Gaardsholt/pass-along/config"
	"github.com/Gaardsholt/pass-along/metrics"
	. "github.com/Gaardsholt/pass-along/types"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var secretStore SecretStore
var templates map[string]*template.Template
var startupTime time.Time
var pr *prometheus.Registry

var lock = sync.RWMutex{}

func init() {
	config.LoadConfig()

	startupTime = time.Now()
	secretStore = SecretStore{
		Data: make(map[string][]byte),
		Lock: &lock,
	}

	templates = make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	templates["read"] = template.Must(template.ParseFiles("templates/base.html", "templates/read.html"))

	pr = prometheus.NewRegistry()
	pr.MustRegister(NewSecretsInCache(&secretStore))
	pr.MustRegister(metrics.SecretsRead)
	pr.MustRegister(metrics.ExpiredSecretsRead)
	pr.MustRegister(metrics.NonExistentSecretsRead)
	pr.MustRegister(metrics.SecretsCreated)
	pr.MustRegister(metrics.SecretsCreatedWithError)
	pr.MustRegister(metrics.SecretsDeleted)
}

func secretCleaner() {
	for {
		time.Sleep(5 * time.Second)
		secretStore.Lock.RLock()
		for k, v := range secretStore.Data {
			s, err := Decrypt(v, k)
			if err != nil {
				continue
			}

			isNotExpired := s.Expires.UTC().After(time.Now().UTC())
			if !isNotExpired {
				log.Debug().Msg("Found expired secret, deleting...")
				secretStore.Lock.RUnlock()
				secretStore.Delete(k)
				secretStore.Lock.RLock()
			}
		}
		secretStore.Lock.RUnlock()
	}
}

func main() {

	// Start loop that checks for expired secrets and deletes them
	go secretCleaner()

	r := mux.NewRouter()
	// Start of static stuff
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/js/").Handler(fs)
	r.PathPrefix("/css/").Handler(fs)
	r.PathPrefix("/favicon.ico").Handler(fs)
	r.PathPrefix("/robots.txt").Handler(fs)
	// End of static stuff

	r.HandleFunc("/", IndexHandler).Methods("GET")
	r.HandleFunc("/", NewHandler).Methods("POST")
	r.PathPrefix("/metrics").Handler(promhttp.HandlerFor(pr, promhttp.HandlerOpts{})).Methods("GET")
	// r.HandleFunc("/metrics", promhttp.Handler()).Methods("GET")
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
