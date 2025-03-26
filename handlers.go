package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task struct {
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendError(w, "Ошибка декодирования JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	now := time.Now()
	today := now.Format("20060102")
	var dateStr string

	if task.Date == "" || task.Date == "today" {
		dateStr = today
	} else {
		if _, err := time.Parse("20060102", task.Date); err != nil {
			sendError(w, "Неверный формат даты (ожидается YYYYMMDD)", http.StatusBadRequest)
			return
		}
		dateStr = task.Date
	}

	if task.Date == "" || task.Date == "today" {
		dateStr = today
	} else if task.Repeat != "" {
		if task.Repeat == "d 1" && dateStr == today {
		} else {
			nextDate, err := NextDate(now, dateStr, task.Repeat)
			if err != nil {
				sendError(w, "Неверный формат правила повторения: "+err.Error(), http.StatusBadRequest)
				return
			}
			dateStr = nextDate
		}
	} else if dateStr < today {
		dateStr = today
	}

	res, err := db.Exec(
		"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		dateStr,
		task.Title,
		task.Comment,
		task.Repeat,
	)
	if err != nil {
		sendError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		sendError(w, "Ошибка при получении ID задачи", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	var rows *sql.Rows
	var err error

	tasks := make([]map[string]string, 0)

	if search != "" {
		if date, err := time.Parse("02.01.2006", search); err == nil {
			dateStr := date.Format("20060102")
			rows, err = db.Query(
				"SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date, id",
				dateStr,
			)
		} else {
			searchPattern := "%" + search + "%"
			rows, err = db.Query(
				"SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date, id",
				searchPattern, searchPattern,
			)
		}
	} else {
		rows, err = db.Query(
			"SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date, id",
		)
	}

	if err != nil {
		sendError(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var date, title, comment, repeat string
		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			continue
		}
		tasks = append(tasks, map[string]string{
			"id":      strconv.Itoa(id),
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	var task struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	err := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err == sql.ErrNoRows {
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	} else if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
		return
	}

	fmt.Println("Задача из БД:", task)

	if task.Repeat != "" {
		now := time.Now()
		nextDate, err := NextDateForTask(now, task.Date, task.Repeat)
		if err == nil {
			task.Date = nextDate
		} else {
			fmt.Println("Ошибка", err)
		}
	}

	json.NewEncoder(w).Encode(task)
}

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "ID не найдено", http.StatusBadRequest)
		return
	}

	var currentDateStr, repeatRule string
	err := db.QueryRow("SELECT date, repeat FROM scheduler WHERE id = ?", id).
		Scan(&currentDateStr, &repeatRule)
	if err != nil {
		if err == sql.ErrNoRows {
			sendError(w, "Задание не найдено", http.StatusNotFound)
		} else {
			sendError(w, "Ошибка БД", http.StatusInternalServerError)
		}
		return
	}

	if repeatRule == "" {
		_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			sendError(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}
	} else {
		nextDate, err := NextDate(time.Now(), currentDateStr, repeatRule)
		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
		if err != nil {
			sendError(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}
	}
	nextDate, err := NextDate(time.Now(), currentDateStr, repeatRule)
	fmt.Println(currentDateStr, repeatRule, nextDate)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{}{})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	fmt.Println("Удаление задачи с ID:", id)

	res, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	fmt.Println("Количество удаленных строк:", rowsAffected)

	if rowsAffected == 0 {
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	json.NewEncoder(w).Encode(struct{}{})
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка декодирования JSON"})
		return
	}

	if request.ID == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан идентификатор"})
		return
	}

	if request.Title == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	now := time.Now()
	today := now.Format("20060102")
	var dateStr string

	if request.Date == "" || request.Date == "today" {
		dateStr = today
	} else {
		if _, err := time.Parse("20060102", request.Date); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат даты (ожидается YYYYMMDD)"})
			return
		}
		dateStr = request.Date
	}

	if request.Repeat != "" {
		if _, err := NextDate(now, dateStr, request.Repeat); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат правила повторения: " + err.Error()})
			return
		}
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", request.ID).Scan(&exists)
	if err != nil || !exists {
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	_, err = db.Exec(
		"UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		dateStr,
		request.Title,
		request.Comment,
		request.Repeat,
		request.ID,
	)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неверный формат даты"))
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
