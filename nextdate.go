package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("неверный формат правила повторения")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("неверный формат для правила 'd'")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", fmt.Errorf("неверное число дней: %v", err)
		}

		if days <= 0 || days > 400 {
			return "", errors.New("интервал дней должен быть от 1 до 400")
		}

		nextDate := parsedDate
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(now) {
				break
			}
		}

		return nextDate.Format("20060102"), nil

	case "y":
		if len(parts) != 1 {
			return "", errors.New("неверный формат для правила 'y'")
		}

		if parsedDate.Month() == 2 && parsedDate.Day() == 29 {
			nextDate := parsedDate.AddDate(1, 0, 0)
			if !isLeap(nextDate.Year()) {
				nextDate = time.Date(nextDate.Year(), 3, 1, 0, 0, 0, 0, time.UTC)
			}
			if nextDate.After(now) {
				return nextDate.Format("20060102"), nil
			}
			for {
				nextDate = nextDate.AddDate(1, 0, 0)
				if isLeap(nextDate.Year()) {
					nextDate = time.Date(nextDate.Year(), 2, 29, 0, 0, 0, 0, time.UTC)
				} else {
					nextDate = time.Date(nextDate.Year(), 3, 1, 0, 0, 0, 0, time.UTC)
				}
				if nextDate.After(now) {
					break
				}
			}
			return nextDate.Format("20060102"), nil
		}

		nextDate := parsedDate
		for {
			nextDate = nextDate.AddDate(1, 0, 0)
			if nextDate.After(now) {
				break
			}
		}

		return nextDate.Format("20060102"), nil

	default:
		return "", errors.New("неподдерживаемый формат правила повторения")
	}
}
func NextDateForTask(now time.Time, lastDate string, repeat string) (string, error) {

	if repeat == "" {
		return "", nil
	}

	parsedDate, err := time.Parse("20060102", lastDate)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("неверный формат правила повторения")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("неверный формат для правила 'd'")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("интервал дней должен быть от 1 до 400")
		}

		nextDate := parsedDate
		for nextDate.Before(now) || nextDate.Equal(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}
		nextDate = nextDate.AddDate(0, 0, -days)
		return nextDate.Format("20060102"), nil

	case "y":
		if len(parts) != 1 {
			return "", errors.New("неверный формат для правила 'y'")
		}

		nextDate := parsedDate.AddDate(1, 0, 0)
		for nextDate.Before(now) || nextDate.Equal(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}

		return nextDate.Format("20060102"), nil

	default:
		return "", errors.New("неподдерживаемый формат правила")
	}
}

func isLeap(year int) bool {
	if year%4 != 0 {
		return false
	} else if year%100 != 0 {
		return true
	} else {
		return year%400 == 0
	}
}
