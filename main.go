package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type runTaskParams struct {
	userId    string
	chapterId string
	taskIndex int
}

type runTaskResult struct {
	Status  int    `json:"code"`
	Message string `json:"msg"`
}

func extractRunTaskParams(r *http.Request) (runTaskParams, error) {
	urlParams := r.URL.Query()
	var res runTaskParams

	res.userId = urlParams.Get("user_id")

	if len(res.userId) == 0 {
		return runTaskParams{}, errors.New("invalid user id")
	}

	res.chapterId = urlParams.Get("chapter_id")

	if len(res.chapterId) == 0 {
		return runTaskParams{}, errors.New("invalid chapter id")
	}

	taskIndex, err := strconv.Atoi(urlParams.Get("task"))
	if err != nil {
		log.Printf("Couldn't convert task_index to int. %v. id: %s", err, urlParams.Get("task"))
		return runTaskParams{}, errors.New("invalid task id")
	}

	res.taskIndex = taskIndex
	return res, nil
}

func HandleRunTask(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")

	params, err := extractRunTaskParams(r)
	if err != nil {
		log.Printf("Couldn't parse request params. Err: %s", err)
		w.Write([]byte("{}"))
		return
	}

	log.Println("Good params", params)

	var res runTaskResult
	res.Status = 0
	res.Message = "all tests passed"

	body, err := json.Marshal(res)

	if err != nil {
		log.Printf("Error in JSON marshal. Err: %s", err)
		w.Write([]byte("{}"))
		return
	}

	w.Write(body)
}

const version = 1.0
const addr = "127.0.0.1:8080"

func main() {
	log.Println("Started handyman", version)
	log.Println("GOMAXPROCS", runtime.GOMAXPROCS(-1))

	r := mux.NewRouter()
	r.HandleFunc("/run_task", HandleRunTask)

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
