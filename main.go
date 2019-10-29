package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/profile"
)

const port = 8080
const randTask = 50

func main() {
	defer profile.Start().Stop()

	f, err := os.OpenFile("gomodu.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
	log.SetOutput(f)
	log.Println("init started")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/agg", aggregateHandler)
	router.HandleFunc("/task/{source}", taskHandler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vars := mux.Vars(r)
	source := vars["source"]

	taskDuration := time.Duration(rand.Intn(10)+10) * time.Millisecond
	nbTasks := rand.Intn(randTask) + randTask
	tasks := make([]task, nbTasks)
	for i := 0; i < nbTasks; i++ {
		time.Sleep(taskDuration)
		id := strconv.Itoa(i)
		task := task{
			ID:          source + "_" + id,
			Name:        "Name" + "_" + id,
			Description: "Does't matter",
		}
		log.Printf("%s Produced Task: %s in %s \n", source, task.ID, taskDuration)
		tasks[i] = task
	}
	createSlice := time.Since(start)
	err := json.NewEncoder(w).Encode(tasks)
	if err != nil {
		log.Println("error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	//w.WriteHeader(http.StatusOK)

	total := time.Since(start)
	log.Printf("%s finished to produced %d tasks in %s/%s \n", source, nbTasks, createSlice, total)
}

func aggregateHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	encoder := json.NewEncoder(w)
	tasks := make(chan task)

	// we have 3 goroutines, we need a WaitGroup to orchestrate the work
	var wg sync.WaitGroup
	wg.Add(3)

	// fork the work in parallel
	go readTask("sourceA", tasks, &wg)
	go readTask("sourceB", tasks, &wg)
	go readTask("sourceC", tasks, &wg)

	// cleanup goroutine
	go func() {
		start := time.Now()
		wg.Wait()
		elapsed := time.Since(start)
		log.Printf("All produced: let's close, done in %s \n", elapsed)
		close(tasks)
	}()

	// this create the JSON in parallel
	for t := range tasks {
		err := encoder.Encode(t)
		if err != nil {
			log.Println("error:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	elapsed := time.Since(start)
	log.Println("Finished to aggregated in", elapsed)
}

func readTask(source string, tasks chan task, wg *sync.WaitGroup) {
	defer wg.Done()
	res, err := http.Get("http://localhost:" + strconv.Itoa(port) + "/task/" + source)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println("error:", err)
		return
	}

	var newTasks []task
	jsonErr := json.Unmarshal(body, &newTasks)
	if jsonErr != nil {
		fmt.Println("error:", err)
		return
	}

	for _, task := range newTasks {
		tasks <- task
	}
}

type task struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}
