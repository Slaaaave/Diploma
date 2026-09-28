package api

import (
	"fmt"
	"time"

	"github.com/Slaaaave/Diploma/pkg/date"
	"github.com/Slaaaave/Diploma/pkg/db"
)

func checkTask(t *db.Task) error {
	if t.Title == "" {
		return fmt.Errorf("не указан заголовок задачи")
	}

	now := time.Now()
	today := now.Truncate(24 * time.Hour)

	if t.Date == "" {
		t.Date = now.Format(date.DateFormat)
	}

	parsed, err := time.Parse(date.DateFormat, t.Date)
	if err != nil {
		return fmt.Errorf("дата указана в неверном формате")
	}

	if t.Repeat != "" {
		if _, err := date.NextDate(now, t.Date, t.Repeat); err != nil {
			return fmt.Errorf("неверное правило повторения: %w", err)
		}
	}

	if parsed.Before(today) {
		if t.Repeat == "" {
			t.Date = now.Format(date.DateFormat)
		} else {
			next, err := date.NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return err
			}
			t.Date = next
		}
	}

	return nil
}
