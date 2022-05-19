package api

import (
	"encoding/json"
	"fmt"
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	ErrServerShuttingDown = "http: Server closed"
)

var secretStore datastore.SecretStore
var startupTime time.Time
var lock = sync.RWMutex{}

// StartServer starts the internal and external http server and initiates the secrets store
func StartServer() (internalServer *http.Server, externalServer *http.Server) {
	startupTime = time.Now()

	databaseType, err := config.Config.GetDatabaseType()
	if err != nil {
		log.Fatal().Err(err).Msgf("%s", err)
	}

	switch databaseType {
	case "in-memory":
		secretStore, err = memory.New(&lock)
	case "redis":
		secretStore, err = redis.New()
	}

	if err != nil {
		log.Fatal().Err(err).Msgf("%s", err)
	}

	registerPrometheusMetrics()

	internal := mux.NewRouter()
	external := mux.NewRouter()
	external.HandleFunc("/api", NewHandler).Methods("POST")
	external.HandleFunc("/api/{id}", GetHandler).Methods("GET")
	external.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	// external.HandleFunc("/", IndexHandler).Methods("GET")

	internal.HandleFunc("/healthz", healthz)
	internal.Handle("/metrics", promhttp.HandlerFor(pr, promhttp.HandlerOpts{})).Methods("GET")

	internalPort := config.Config.GetHealthPort()
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

	externalPort := config.Config.GetServerPort()
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
	log.Info().Msgf("Starting server with '%s' as datastore", databaseType)
	log.Info().Msgf("Site can now be accessed at http://localhost:%d", externalPort)
	log.Info().Msgf("Health and metrics and can be accessed on http://localhost:%d", internalPort)

	go secretStore.DeleteExpiredSecrets()

	return
}

// NewHandler creates a new secret in the secretstore
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
		go metrics.SecretsCreatedWithError.Inc()
		return
	}

	err = secretStore.Add(id, encryptedSecret, entry.ExpiresIn)
	if err != nil {
		http.Error(w, "failed to add secret, please try again", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Unable to add secret")
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", id)
}

// GetHandler retrieves a secret in the secret store
func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

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
		go metrics.SecretsRead.Inc()
	} else {
		gotData = false
		go metrics.ExpiredSecretsRead.Inc()
	}

	if !isNotExpired || !s.UnlimitedViews {
		secretStore.Delete(id)
	}

	log.Debug().Msg("Fetching a secret")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", decryptedSecret)
}
