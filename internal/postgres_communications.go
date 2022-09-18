package internal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gammazero/workerpool"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// An exported global variable to hold the database connection pool
// Completely thread-safe and ok. Fear not, my friend
var DB *sql.DB

var WP *workerpool.WorkerPool

const connStr = "postgresql://senjun:some_password@127.0.0.1:5432/senjun?sslmode=disable"

func ConnectDb() *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Couldn't call Open() for db")
	}

	err = db.Ping()

	if err != nil {
		log.WithFields(log.Fields{
			"connStr": connStr,
			"error":   err,
		}).Fatal("Couldn't communicate db")
	}

	return db
}

type updateChapterStatusParams struct {
	userId    string
	chapterId string
	status    string
}

func extractUpdateChapterStatusParams(r *http.Request) (updateChapterStatusParams, error) {
	urlParams := r.URL.Query()
	var res updateChapterStatusParams

	res.userId = urlParams.Get("user_id")

	if len(res.userId) == 0 {
		return updateChapterStatusParams{}, errors.New("invalid user id")
	}

	res.chapterId = urlParams.Get("chapter_id")

	if len(res.chapterId) == 0 {
		return updateChapterStatusParams{}, errors.New("invalid chapter id")
	}

	res.status = urlParams.Get("status")
	if len(res.status) == 0 {
		return updateChapterStatusParams{}, errors.New("invalid status")
	}

	return res, nil
}

func UpdateChapterStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	params, err := extractUpdateChapterStatusParams(r)
	if err != nil {
		body, _ := json.Marshal(map[string]string{
			"error": fmt.Sprintf("Invalid request: %s", err),
		})
		w.Write(body)
		return
	}

	log.WithFields(log.Fields{
		"userId":    params.userId,
		"chapterId": params.chapterId,
		"status":    params.status,
	}).Debug("Parsed url params")

}

// Returns postgres TYPE edu_material_status
func getEduMaterialStatus(code int) string {
	if code == 0 {
		return "completed"
	}

	return "in_progress"
}

func UpdateTaskStatus(userId string, taskId string, statusCode int, solutionText string) {
	taskStatus := getEduMaterialStatus(statusCode)
	const attemptsCount = 1

	const query = `
		INSERT INTO 
		task_progress(user_id, task_id, status, solution_text, attempts_count)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT ON CONSTRAINT unique_user_task_id
		DO UPDATE SET 
		status = EXCLUDED.status, 
		solution_text = EXCLUDED.solution_text,
		attempts_count = task_progress.attempts_count + EXCLUDED.attempts_count
	`
	_, err := DB.Exec(query, userId, taskId, taskStatus, solutionText, attemptsCount)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id":  userId,
			"task_id":  taskId,
			"db_error": err.Error(),
		}).Error("Couldn't update task status for user")
		return
	}

	log.WithFields(log.Fields{
		"user_id": userId,
		"task_id": taskId,
	}).Info("Updated task status for user")
}
