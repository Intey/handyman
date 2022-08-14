package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const version = "1.0"
const addrHandyman = "127.0.0.1:8080"
const addrWatchman = "http://127.0.0.1:8000/check"

const timeoutReplyToUser = 40 * time.Second
const timeoutReplyFromWatchman = 30 * time.Second

type runTaskParams struct {
	userId        string
	chapterId     string
	taskIndex     int
	sourceCode    string
	containerType string
}

type runTaskResult struct {
	Status  int    `json:"error_code"`
	Message string `json:"output"`
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

	if b, err := io.ReadAll(r.Body); err == nil {
		res.sourceCode = string(b)
	} else {
		panic("Couldn't read body")
	}

	if len(res.sourceCode) == 0 {
		return runTaskParams{}, errors.New("Empty source code")
	}

	// TODO: get container type
	res.containerType = "python_env"

	return res, nil
}

func generateTaskId(params runTaskParams) string {
	return "some_task_id"
}

func communicateWatchman(params runTaskParams) (runTaskResult, error) {
	postBody, _ := json.Marshal(map[string]string{
		"container_type": params.containerType,
		"source":    params.sourceCode,
		"task_id": generateTaskId(params),
	})
	reqBody := bytes.NewBuffer(postBody)

	client := http.Client{
		Timeout: timeoutReplyFromWatchman,
	}

	resp, err := client.Post(addrWatchman, "application/json", reqBody)

	if err != nil {
		log.Printf("Couldn't send request to watchman %v", err)
		return runTaskResult{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return runTaskResult{}, errors.New("HTTP error " + strconv.Itoa(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Couldn't read body %v", err)
		return runTaskResult{}, err
	}

	var res runTaskResult
	err = json.Unmarshal(body, &res)

	if err != nil {
		log.Printf("Couldn't parse json body %v", err)
		return runTaskResult{}, err
	}

	return res, nil
}

func HandleRunTask(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")

	params, err := extractRunTaskParams(r)
	if err != nil {
		body, _ := json.Marshal(map[string]string{
			"error": fmt.Sprintf("Invalid request: %s", err),
		})
		w.Write(body)
		return
	}

	log.Println("Parsed url params", params)

	runTaskRes, err := communicateWatchman(params)
	if err != nil {
		body, _ := json.Marshal(map[string]string{
			"error": fmt.Sprintf("Couldn't communicate with tasks runner: %s", err),
		})
		w.Write(body)
		return
	}

	log.Println("Successfully communicated watchman: " + string(runTaskRes.Status))
	json.NewEncoder(w).Encode(runTaskRes)
}

func main() {
	log.Printf("Started handyman %v listening on %v. GOMAXPROCS=%v",
		version, addrHandyman, runtime.GOMAXPROCS(-1))

	r := mux.NewRouter()
	r.HandleFunc("/run_task", HandleRunTask)

	srv := &http.Server{
		Handler:      r,
		Addr:         addrHandyman,
		WriteTimeout: timeoutReplyToUser,
		ReadTimeout:  timeoutReplyToUser,
	}
	log.Fatal(srv.ListenAndServe())
}
