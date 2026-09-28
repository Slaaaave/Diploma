// Package db отвечает за хранение задач в SQLite.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

// Task — задача планировщика.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

const schema = `
CREATE TABLE scheduler (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	date     CHAR(8)      NOT NULL DEFAULT "",
	title    VARCHAR(256) NOT NULL DEFAULT "",
	comment  TEXT         NOT NULL DEFAULT "",
	repeat   VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX idx_scheduler_date ON scheduler (date);
`

var conn *sql.DB

// Init открывает базу данных по пути dbFile и создаёт таблицу scheduler,
// если файла базы ещё не существует.
func Init(dbFile string) error {
	_, statErr := os.Stat(dbFile)
	needsCreate := errors.Is(statErr, os.ErrNotExist)

	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("не удалось открыть базу данных: %w", err)
	}

	if err := database.Ping(); err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	if needsCreate {
		if _, err := database.Exec(schema); err != nil {
			return fmt.Errorf("не удалось создать схему базы данных: %w", err)
		}
	}

	conn = database
	return nil
}

// Close закрывает соединение с базой данных.
func Close() error {
	if conn == nil {
		return nil
	}
	return conn.Close()
}

// AddTask добавляет новую задачу и возвращает её id.
func AddTask(t *Task) (int64, error) {
	res, err := conn.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		t.Date, t.Title, t.Comment, t.Repeat,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Tasks возвращает список задач, отсортированных по дате.
// Если search непустой, фильтрует по дате (dd.mm.yyyy) либо по подстроке
// в заголовке или комментарии.
func Tasks(limit int, search string) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	switch {
	case search == "":
		rows, err = conn.Query(
			`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`,
			limit,
		)
	default:
		if d, dateErr := time.Parse("02.01.2006", search); dateErr == nil {
			rows, err = conn.Query(
				`SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`,
				d.Format("20060102"), limit,
			)
		} else {
			like := "%" + search + "%"
			rows, err = conn.Query(
				`SELECT id, date, title, comment, repeat FROM scheduler
				 WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`,
				like, like, limit,
			)
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		var id int64
		t := &Task{}
		if err := rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		t.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetTask возвращает задачу по id.
func GetTask(id string) (*Task, error) {
	t := &Task{}
	var idVal int64
	row := conn.QueryRow(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id,
	)
	if err := row.Scan(&idVal, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("задача с id %q не найдена", id)
		}
		return nil, err
	}
	t.ID = strconv.FormatInt(idVal, 10)
	return t, nil
}

// UpdateTask обновляет существующую задачу.
func UpdateTask(t *Task) error {
	res, err := conn.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		t.Date, t.Title, t.Comment, t.Repeat, t.ID,
	)
	if err != nil {
		return err
	}
	return checkAffected(res, t.ID)
}

// UpdateDate обновляет только дату задачи (используется при отметке "выполнено").
func UpdateDate(id, date string) error {
	res, err := conn.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, date, id)
	if err != nil {
		return err
	}
	return checkAffected(res, id)
}

// DeleteTask удаляет задачу по id.
func DeleteTask(id string) error {
	res, err := conn.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res, id)
}

func checkAffected(res sql.Result, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("задача с id %q не найдена", id)
	}
	return nil
}
