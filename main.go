package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"

	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// Todo todo struct
type Todo struct {
	gorm.Model `json:"-"`
	Task       string `json:"task"`
	ID         string `json:"id"`
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	task := r.FormValue("task")
	todo := &Todo{Task: task, ID: uuid.New().String()}
	result := db.Create(todo)
	if result.Error != nil {
		log.Error(result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
	log.Info("saved todo item")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	todo := &Todo{ID: id}
	result := db.Delete(todo)
	if result.Error != nil {
		log.Error(result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info("deleted todo item")
}

func getItems(w http.ResponseWriter, r *http.Request) {
	var all []Todo
	result := db.Find(&all)
	if result.Error != nil {
		log.Error(result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Task > all[j].Task
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
	log.Infof("got %d items", len(all))
}

func main() {
	host := os.Getenv("POSTGRESQL_HOST")
	username := os.Getenv("POSTGRESQL_USERNAME")
	password := os.Getenv("POSTGRESQL_PASSWORD")
	database := os.Getenv("POSTGRESQL_DATABASE")
	log.Infof("Connecting to DB '%s'...", host)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, 5432, username, password, database)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Todo{})

	log.Infof("Connected to DB '%s'.", host)

	router := mux.NewRouter()
	router.HandleFunc("/healthz", healthz).Methods("GET")
	router.HandleFunc("/todo", getItems).Methods("GET")
	router.HandleFunc("/todo", createItem).Methods("POST")
	router.HandleFunc("/todo/{id}", deleteItem).Methods("DELETE")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	log.Info("Starting API server...")
	http.ListenAndServe(":8080", handler)
}
