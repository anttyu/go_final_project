package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/api/tasks", getTasksHandler).Methods("GET")
	r.HandleFunc("/api/task", getTaskHandler).Methods("GET")
	r.HandleFunc("/api/task", addTaskHandler).Methods("POST")
	r.HandleFunc("/api/task/done", doneTaskHandler).Methods("POST")
	r.HandleFunc("/api/task", deleteTaskHandler).Methods("DELETE")
	r.HandleFunc("/api/nextdate", nextDateHandler).Methods("GET")
	r.HandleFunc("/api/task", updateTaskHandler).Methods("PUT")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
