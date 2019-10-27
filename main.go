package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	go createTask("sourceA", tasks, &wg)
	go createTask("sourceB", tasks, &wg)
	go createTask("sourceC", tasks, &wg)

	// cleanup goroutine
	go func() {
		start := time.Now()
		wg.Wait()
		elapsed := time.Since(start)
		fmt.Printf("All produced: let's close, done in %s", elapsed)
		close(tasks)
	}()

	// this create the JSON in parallel
	for t := range tasks {
		err := encoder.Encode(t)
		if err != nil {
			fmt.Printf("Something wrong happened")
		}
	}

	w.Header().Set("Content-Type", "application/json")
}

func createTask(source string, tasks chan task, wg *sync.WaitGroup) {
	defer wg.Done()
	taskDuration := time.Duration(rand.Intn(10)+10) * time.Millisecond
	nbTasks := rand.Intn(50) + 50
	for i := 1; i <= nbTasks; i++ {
		time.Sleep(taskDuration)
		id := strconv.Itoa(i)
		task := task{
			ID:          source + "_" + id,
			Name:        "Name" + "_" + id,
			Description: "Does't matter",
		}
		fmt.Printf("%s Produced Task: %s in %s \n", source, task.ID, taskDuration)
		tasks <- task
	}
}

type task struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}
