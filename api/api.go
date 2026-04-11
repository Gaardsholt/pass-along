package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	external.Use(securityHeadersMiddleware)

	apiRouter := external.PathPrefix("/api").Subrouter()
	apiRouter.Use(noStoreMiddleware)
	apiRouter.HandleFunc("", NewHandler).Methods("POST")
	apiRouter.HandleFunc("/config", ConfigHandler).Methods("GET")
	apiRouter.HandleFunc("/{id}", GetHandler).Methods("GET")

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
		r.Body = http.MaxBytesReader(w, r.Body, int64(config.Config.MaxSecretBytes))
		err = json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: malformed json")
			return
		}
	} else {
		err = getFormData(w, r, &entry)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	if err := validateEntry(entry); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
		Files:          entry.Files,
		Expires:        expires,
		TimeAdded:      time.Now(),
		UnlimitedViews: entry.UnlimitedViews,
	}
	mySecret.UnlimitedViews = entry.UnlimitedViews
	lookupID, accessKey, token, err := types.GenerateToken()
	if err != nil {
		go metrics.SecretsCreatedWithError.Inc()
		writeError(w, http.StatusInternalServerError, "failed to create secret")
		return
	}

	encryptedSecret, err := mySecret.Encrypt(accessKey)
	if err != nil {
		go metrics.SecretsCreatedWithError.Inc()
		writeError(w, http.StatusInternalServerError, "failed to create secret")
		return
	}

	err = secretStore.Add(lookupID, encryptedSecret, entry.ExpiresIn)
	if err != nil {
		go metrics.SecretsCreatedWithError.Inc()
		writeError(w, http.StatusInternalServerError, "failed to add secret, please try again")
		log.Error().Err(err).Msg("Unable to add secret")
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", token)
}

// ConfigResponse is the struct being sent to the frontend for the `/api/config` endpoint
type ConfigResponse struct {
	ValidForOptions  map[int]string `json:"valid_for_options"`
	MaxFiles         int            `json:"max_files"`
	MaxFileSizeBytes int64          `json:"max_file_size_bytes"`
}

// ConfigHandler returns the server configuration
func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	options := map[int]string{}

	for _, v := range config.Config.ValidForOptions {
		options[v] = humanDuration(v)
	}

	response := ConfigResponse{
		ValidForOptions:  options,
		MaxFiles:         config.Config.MaxFiles,
		MaxFileSizeBytes: config.Config.MaxFileSizeBytes,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	token := vars["id"]
	lookupID, accessKey, err := types.ParseToken(token)
	if err != nil {
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "secret not found")
		return
	}

	secretData, gotData := secretStore.Get(lookupID)
	if !gotData {
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "secret not found")
		return
	}

	s, err := types.Decrypt(secretData, accessKey)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to decrypt secret")
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "secret not found")
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
		secretStore.Delete(lookupID)
	}

	log.Debug().Msg("Fetching a secret")

	response := map[string]interface{}{
		"content":         s.Content,
		"files":           s.Files,
		"expires":         s.Expires,
		"unlimited_views": s.UnlimitedViews,
	}
	jsonResponse, jsonError := json.Marshal(response)
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

func getFormData(w http.ResponseWriter, r *http.Request, entry *types.Entry) error {
	maxMultipartBytes := int64(config.Config.MaxFiles)*config.Config.MaxFileSizeBytes + int64(config.Config.MaxSecretBytes) + 1024*1024
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartBytes)
	err := r.ParseMultipartForm(maxMultipartBytes)
	if err != nil {
		return errors.New("failed to parse multipart form")
	}
	mForm := r.MultipartForm
	if mForm == nil {
		return errors.New("multipart form is missing")
	}

	dataValues, ok := mForm.Value["data"]
	if !ok || len(dataValues) == 0 || strings.TrimSpace(dataValues[0]) == "" {
		return errors.New("missing data payload")
	}

	myData := dataValues[0]
	err = json.Unmarshal([]byte(myData), entry)
	if err != nil {
		return errors.New("invalid data payload")
	}

	filesMap := map[string][]byte{}
	fileHeaders := mForm.File["files"]
	if len(fileHeaders) > config.Config.MaxFiles {
		return errors.New("too many files")
	}

	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			return errors.New("unable to read file")
		}
		defer file.Close()

		limitedReader := io.LimitReader(file, config.Config.MaxFileSizeBytes+1)
		fileBytes, err := io.ReadAll(limitedReader)
		if err != nil {
			return errors.New("unable to read file")
		}
		if int64(len(fileBytes)) > config.Config.MaxFileSizeBytes {
			return errors.New("file too large")
		}
		filesMap[fileHeader.Filename] = fileBytes
	}

	if len(filesMap) > 0 {
		entry.Files = filesMap
	}
	return nil
}

func validateEntry(entry types.Entry) error {
	if !config.Config.IsValidExpiration(entry.ExpiresIn) {
		return errors.New("invalid expires_in value")
	}

	if len(entry.Content) == 0 && len(entry.Files) == 0 {
		return errors.New("content or files must be provided")
	}

	if len(entry.Content) > config.Config.MaxSecretBytes {
		return errors.New("content too large")
	}

	if len(entry.Files) > config.Config.MaxFiles {
		return errors.New("too many files")
	}

	for _, content := range entry.Files {
		if int64(len(content)) > config.Config.MaxFileSizeBytes {
			return errors.New("file too large")
		}
	}

	return nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func noStoreMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "accelerometer=(),camera=(),geolocation=(),microphone=(),payment=(),usb=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; object-src 'none'; form-action 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'")
		if config.Config.EnableHSTS && (r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")) {
			w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains", config.Config.HSTSMaxAgeSeconds))
		}
		next.ServeHTTP(w, r)
	})
}
