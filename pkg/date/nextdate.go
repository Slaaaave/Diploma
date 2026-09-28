package date

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

const maxIterations = 100000

func NextDate(now time.Time, date string, repeat string) (string, error) {
	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не может быть пустым")
	}

	start, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты %q: %w", date, err)
	}

	fields := strings.Fields(repeat)

	switch fields[0] {
	case "d":
		return nextDaily(now, start, fields)
	case "y":
		return nextYearly(now, start)
	case "w":
		return nextByWeekday(now, start, fields)
	case "m":
		return nextByMonthDay(now, start, fields)
	default:
		return "", fmt.Errorf("неподдерживаемый формат правила повторения: %q", repeat)
	}
}

func nextDaily(now, start time.Time, fields []string) (string, error) {
	if len(fields) != 2 {
		return "", fmt.Errorf("правило \"d\" должно быть вида \"d <число дней>\"")
	}
	days, err := strconv.Atoi(fields[1])
	if err != nil || days < 1 || days > 400 {
		return "", fmt.Errorf("интервал дней должен быть числом от 1 до 400")
	}

	next := start
	for i := 0; i < maxIterations; i++ {
		next = next.AddDate(0, 0, days)
		if next.After(now) {
			return next.Format(DateFormat), nil
		}
	}
	return "", fmt.Errorf("не удалось вычислить следующую дату")
}

func nextYearly(now, start time.Time) (string, error) {
	next := start
	for i := 0; i < maxIterations; i++ {
		next = next.AddDate(1, 0, 0)
		if next.After(now) {
			return next.Format(DateFormat), nil
		}
	}
	return "", fmt.Errorf("не удалось вычислить следующую дату")
}

func nextByWeekday(now, start time.Time, fields []string) (string, error) {
	if len(fields) != 2 {
		return "", fmt.Errorf("правило \"w\" должно быть вида \"w <дни недели через запятую>\"")
	}
	weekdays, err := parseIntList(fields[1], 1, 7)
	if err != nil {
		return "", err
	}

	next := start
	for i := 0; i < maxIterations; i++ {
		next = next.AddDate(0, 0, 1)
		if !next.After(now) {
			continue
		}
		if containsInt(weekdays, isoWeekday(next)) {
			return next.Format(DateFormat), nil
		}
	}
	return "", fmt.Errorf("не удалось вычислить следующую дату")
}

func nextByMonthDay(now, start time.Time, fields []string) (string, error) {
	if len(fields) < 2 || len(fields) > 3 {
		return "", fmt.Errorf("правило \"m\" должно быть вида \"m <дни>\" или \"m <дни> <месяцы>\"")
	}

	days, err := parseIntList(fields[1], -2, 31)
	if err != nil {
		return "", err
	}
	for _, d := range days {
		if d == 0 || (d < 0 && d != -1 && d != -2) {
			return "", fmt.Errorf("недопустимый день месяца: %d", d)
		}
	}

	var months []int
	if len(fields) == 3 {
		months, err = parseIntList(fields[2], 1, 12)
		if err != nil {
			return "", err
		}
	}

	next := start
	for i := 0; i < maxIterations; i++ {
		next = next.AddDate(0, 0, 1)
		if !next.After(now) {
			continue
		}
		if len(months) > 0 && !containsInt(months, int(next.Month())) {
			continue
		}
		if dayMatches(next, days) {
			return next.Format(DateFormat), nil
		}
	}
	return "", fmt.Errorf("не удалось вычислить следующую дату")
}

func dayMatches(t time.Time, days []int) bool {
	last := lastDayOfMonth(t)
	for _, d := range days {
		switch {
		case d > 0 && t.Day() == d:
			return true
		case d == -1 && t.Day() == last:
			return true
		case d == -2 && t.Day() == last-1:
			return true
		}
	}
	return false
}

func lastDayOfMonth(t time.Time) int {
	firstOfNextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func parseIntList(s string, min, max int) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("некорректное число %q в правиле повторения", p)
		}
		if n < min || n > max {
			return nil, fmt.Errorf("число %d вне допустимого диапазона [%d, %d]", n, min, max)
		}
		result = append(result, n)
	}
	return result, nil
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
