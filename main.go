package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
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

	// we have 3 goroutines, we need a WaitGroup to orchestrate the work
	var wg sync.WaitGroup
	wg.Add(3)

	// fork the work in parallel
	go createTask(tasks, "sourceA", &wg)
	go createTask(tasks, "sourceB", &wg)
	go createTask(tasks, "sourceC", &wg)

	// cleanup goroutine
	go func() {
		wg.Wait()
		fmt.Println("All produced: let's close")
		close(tasks)
	}()

	// this create the JSON in parallel
	for t := range tasks {
		err := encoder.Encode(t)
		if err != nil {
			fmt.Println("Something wrong happened")
		}
	}
	//w.WriteHeader(http.StatusOK)
}

func createTask(tasks chan task, source string, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i <= 10; i++ {
		time.Sleep(100 * time.Millisecond)
		id := strconv.Itoa(i)
		task := task{
			ID:          source + "_" + id,
			Name:        "Name" + "_" + id,
			Description: "Does't matter",
		}
		fmt.Println("Produced Task: " + task.ID)
		tasks <- task
	}
}

type task struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}
