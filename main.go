package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/tasks", taskHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	tasks := make(chan task)

	//json.NewEncoder(w).Encode(newTask)
	go createTask(tasks)
	for t := range tasks {
		encoder.Encode(t)
	}
	//w.WriteHeader(http.StatusOK)
}

func createTask(tasks chan task) {
	for i := 1; i <= 10; i++ {
		time.Sleep(100 * time.Millisecond)
		id := strconv.Itoa(i)
		task := task{
			ID:          id,
			Name:        "Name" + id,
			Description: "Does't matter",
		}
		tasks <- task
	}
	close(tasks)
	//json.NewEncoder(w).Encode(newTask)
}

type task struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}
