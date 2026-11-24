package api

import (
	"net/http"
	"time"

	"github.com/Slaaaave/Diploma/pkg/date"
)

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	nowParam := r.URL.Query().Get("now")
	dateParam := r.URL.Query().Get("date")
	repeatParam := r.URL.Query().Get("repeat")

	now, err := time.Parse(date.DateFormat, nowParam)
	if err != nil {
		http.Error(w, "неверный параметр now", http.StatusBadRequest)
		return
	}

	next, err := date.NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(next))
}
