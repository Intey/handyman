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
	"strings"
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
	err     error
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

	if strings.HasPrefix(res.chapterId, "python") {
		res.containerType = "python_env"
	} else if strings.HasPrefix(res.chapterId, "rust") {
		res.containerType = "rust_env"
	} else {
		return runTaskParams{}, errors.New("Couldn't specify container for chapter " + res.chapterId)
	}

	return res, nil
}

func generateTaskId(params runTaskParams) string {
	return fmt.Sprintf("%s_%s_%d_%d", params.chapterId, params.userId,
		params.taskIndex, time.Now().UnixNano())
}

func communicateWatchman(params runTaskParams, c chan runTaskResult) {
	defer close(c)
	res := new(runTaskResult)

	taskId := generateTaskId(params)
	postBody, _ := json.Marshal(map[string]string{
		"container_type": params.containerType,
		"source":         params.sourceCode,
		"task_id":        taskId,
	})
	reqBody := bytes.NewBuffer(postBody)

	client := http.Client{
		Timeout: timeoutReplyFromWatchman,
	}

	resp, err := client.Post(addrWatchman, "application/json", reqBody)

	if err != nil {
		log.Printf("Couldn't send request to watchman %v", err)
		res.err = err
		c <- *res
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		res.err = errors.New("HTTP error " + strconv.Itoa(resp.StatusCode))
		c <- *res
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Couldn't read body %v", err)
		res.err = err
		c <- *res
		return
	}

	err = json.Unmarshal(body, &res)

	if err != nil {
		log.Printf("Couldn't parse json body %v", err)
		res.err = err
		c <- *res
		return
	}

	res.err = nil
	c <- *res
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

	log.Printf("Parsed url params. userId=%v chapterId=%v taskIndex=%v",
		params.userId, params.chapterId, params.taskIndex)
	c := make(chan runTaskResult)

	go communicateWatchman(params, c)
	res := <-c

	if res.err != nil {
		body, _ := json.Marshal(map[string]string{
			"error": fmt.Sprintf("Couldn't communicate with tasks runner: %s", res.err),
		})
		w.Write(body)
		return
	}

	log.Printf("Successfully communicated watchman: %v", res.Status)
	json.NewEncoder(w).Encode(res)
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
