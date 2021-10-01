package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Gaardsholt/share-a-password/secret"
	"github.com/gorilla/mux"
)

var secretStore secret.SecretStore

func main() {
	fmt.Println("Hejsa")

	secretStore = make(secret.SecretStore)

	r := mux.NewRouter()
	r.HandleFunc("/", NewHandler).Methods("POST")
	r.HandleFunc("/{id}", GetHandler).Methods("GET")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static/")))).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	myId, err := secretStore.Add(post.Content, post.ExpiresIn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", myId)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	secretData, gotData := secretStore.Get(vars["id"])
	if !gotData {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", secretData)
}

type Post struct {
	Content   string `json:"content"`
	ExpiresIn int    `json:"expires_in"`
}
