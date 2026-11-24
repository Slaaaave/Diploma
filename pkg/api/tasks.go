package api

import (
	"net/http"

	"github.com/Slaaaave/Diploma/pkg/db"
)

const tasksLimit = 50

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	search := r.URL.Query().Get("search")

	tasks, err := db.Tasks(tasksLimit, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения задач: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string][]*db.Task{"tasks": tasks})
}
