package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	errServerShuttingDown = "http: Server closed"
)

var secretStore datastore.SecretStore
var lock = sync.RWMutex{}

// StartServer starts the internal and external http server and initiates the secrets store
func StartServer() (internalServer *http.Server, externalServer *http.Server) {
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
	external.HandleFunc("/api/valid-for-options", ValidForHandler).Methods("GET")
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
		if err != nil && err.Error() != errServerShuttingDown {
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
		if err != nil && err.Error() != errServerShuttingDown {
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
	var err error
	var entry types.Entry

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		err = getFormData(r, &entry)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	log.Debug().Msg("Creating a new secret")

	expires := time.Now().Add(
		time.Hour*time.Duration(0) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(entry.ExpiresIn),
	)

	mySecret := types.Secret{
		Content:        entry.Content,
		Files:          entry.Files,
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

// ValidForHandler returns the options you can choose in "Valid for" field
func ValidForHandler(w http.ResponseWriter, r *http.Request) {
	options := map[int]string{}

	for _, v := range config.Config.ValidForOptions {
		options[v] = humanDuration(v)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

// humanDuration converts duration in seconds to human readable format
func humanDuration(duration int) string {
	if duration < 60 {
		if duration == 1 {
			return fmt.Sprintf("%d second", duration)
		}
		return fmt.Sprintf("%d seconds", duration)
	}

	if duration < 3600 {
		duration = duration / 60
		if duration == 1 {
			return fmt.Sprintf("%d minute", duration)
		}
		return fmt.Sprintf("%d minutes", duration)
	}

	if duration < 86400 {
		duration = duration / 60 / 60
		if duration == 1 {
			return fmt.Sprintf("%d hour", duration)
		}
		return fmt.Sprintf("%d hours", duration)
	}

	duration = duration / 60 / 60 / 24
	if duration == 1 {
		return fmt.Sprintf("%d day", duration)
	}

	return fmt.Sprintf("%d days", duration/60/60/24)
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

	isNotExpired := s.Expires.UTC().After(time.Now().UTC())
	if isNotExpired {
		go metrics.SecretsRead.Inc()
		gotData = true
	} else {
		gotData = false
		go metrics.ExpiredSecretsRead.Inc()
	}

	if !isNotExpired || !s.UnlimitedViews {
		secretStore.Delete(id)
	}

	log.Debug().Msg("Fetching a secret")

	jsonResponse, jsonError := json.Marshal(s)
	if jsonError != nil {
		fmt.Println("Unable to encode JSON")
	}

	if gotData {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		return
	}

	fmt.Fprintf(w, "")
}

func getFormData(r *http.Request, entry *types.Entry) error {
	r.ParseMultipartForm(32 << 20) // 32 MB
	mForm := r.MultipartForm

	myData := mForm.Value["data"][0]

	json.Unmarshal([]byte(myData), &entry)

	err := json.Unmarshal([]byte(myData), &entry)
	if err != nil {
		return err
	}

	filesMap := map[string][]byte{}

	for _, fileHeader := range mForm.File["files"] {
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		fileBytes, err := ioutil.ReadAll(file)
		filesMap[fileHeader.Filename] = fileBytes
	}

	if len(filesMap) > 0 {
		entry.Files = filesMap
	}
	return nil
}
