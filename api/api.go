package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"sync"
	"time"

	"github.com/Gaardsholt/pass-along/config"
	"github.com/Gaardsholt/pass-along/datastore"
	"github.com/Gaardsholt/pass-along/memory"
	"github.com/Gaardsholt/pass-along/metrics"
	"github.com/Gaardsholt/pass-along/redis"
	"github.com/Gaardsholt/pass-along/types"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	ErrServerShuttingDown = "http: Server closed"
)

var pr *prometheus.Registry
var secretStore datastore.SecretStore
var startupTime time.Time
var templates map[string]*template.Template
var lock = sync.RWMutex{}

func StartServer() (internalServer *http.Server, externalServer *http.Server) {
	startupTime = time.Now()

	databaseType, err := config.Config.GetDatabaseType()
	if err != nil {
		log.Fatal().Err(err).Msgf("%s", err)
	}

	switch databaseType {
	case "in-memory":
		secretStore = memory.NewStore(&lock)
	case "redis":
		secretStore = redis.NewStore(&lock)
	}

	registerPrometheusMetrics()
	createTemplates()

	internal := mux.NewRouter()
	external := mux.NewRouter()
	// Start of static stuff
	fs := http.FileServer(http.Dir("./static"))
	external.PathPrefix("/assets").Handler(http.StripPrefix("/assets", fs))
	external.PathPrefix("/robots.txt").Handler(fs)
	external.PathPrefix("/favicon.ico").Handler(fs)
	// End of static stuff

	external.HandleFunc("/", IndexHandler).Methods("GET")
	external.HandleFunc("/", NewHandler).Methods("POST")
	external.HandleFunc("/{id}", GetHandler).Methods("GET")

	internal.HandleFunc("/healthz", healthz)
	internal.Handle("/metrics", promhttp.HandlerFor(pr, promhttp.HandlerOpts{})).Methods("GET")

	internalPort := 8888
	internalServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", internalPort),
		Handler: internal,
	}

	go func() {
		err := internalServer.ListenAndServe()
		if err != nil && err.Error() != ErrServerShuttingDown {
			log.Fatal().Err(err).Msgf("Unable to run the internal server at port %d", internalPort)
		}
	}()

	externalPort := 8080
	externalServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", externalPort),
		Handler: external,
	}
	go func() {
		err := externalServer.ListenAndServe()
		if err != nil && err.Error() != ErrServerShuttingDown {
			log.Fatal().Err(err).Msgf("Unable to run the external server at port %d", externalPort)
		}
	}()
	log.Info().Msgf("Starting server at port %d", externalPort)

	go secretStore.DeleteExpiredSecrets()

	return
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templates["index"].Execute(w, types.Page{Startup: startupTime})
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	var entry types.Entry
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debug().Msg("Creating a new secret")

	expires := time.Now().Add(
		time.Hour*time.Duration(0) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(entry.ExpiresIn),
	)

	mySecret := types.Secret{
		Content:        entry.Content,
		Expires:        expires,
		TimeAdded:      time.Now(),
		UnlimitedViews: entry.UnlimitedViews,
	}
	mySecret.UnlimitedViews = entry.UnlimitedViews
	id := mySecret.GenerateID()

	encryptedSecret, err := mySecret.Encrypt(id)
	if err != nil {
		metrics.SecretsCreatedWithError.Inc()
		return
	}

	_ = secretStore.Add(id, encryptedSecret, entry.ExpiresIn)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", id)
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
		newError := templates["read"].Execute(w, types.Page{Startup: startupTime})
		if newError != nil {
			fmt.Fprintf(w, "%s", newError)
		}
		return
	}

	id := vars["id"]
	secretData, gotData := secretStore.Get(id)
	if !gotData {
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "secret not found")
		return
	}

	s, err := types.Decrypt(secretData, id)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to decrypt secret")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	decryptedSecret := ""

	isNotExpired := s.Expires.UTC().After(time.Now().UTC())
	if isNotExpired {
		decryptedSecret = s.Content
		metrics.SecretsRead.Inc()
	} else {
		gotData = false
		metrics.ExpiredSecretsRead.Inc()
	}

	if !isNotExpired || !s.UnlimitedViews {
		secretStore.Delete(id)
	}

	log.Debug().Msg("Fetching a secret")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", decryptedSecret)
}

// healthz is a liveness probe.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func registerPrometheusMetrics() {
	pr = prometheus.NewRegistry()
	// pr.MustRegister(types.NewSecretsInCache(&secretStore))
	pr.MustRegister(metrics.SecretsRead)
	pr.MustRegister(metrics.ExpiredSecretsRead)
	pr.MustRegister(metrics.NonExistentSecretsRead)
	pr.MustRegister(metrics.SecretsCreated)
	pr.MustRegister(metrics.SecretsCreatedWithError)
	pr.MustRegister(metrics.SecretsDeleted)
}

func createTemplates() {
	templates = make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	templates["read"] = template.Must(template.ParseFiles("templates/base.html", "templates/read.html"))
}

// func secretCleaner() {
// 	for {
// 		time.Sleep(5 * time.Minute)
// 		secretStore.Lock.RLock()
// 		for k, v := range secretStore.Data {
// 			s, err := types.Decrypt(v, k)
// 			if err != nil {
// 				continue
// 			}

// 			isNotExpired := s.Expires.UTC().After(time.Now().UTC())
// 			if !isNotExpired {
// 				log.Debug().Msg("Found expired secret, deleting...")
// 				secretStore.Lock.RUnlock()
// 				secretStore.Delete(k)
// 				secretStore.Lock.RLock()
// 			}
// 		}
// 		secretStore.Lock.RUnlock()
// 	}
// }
