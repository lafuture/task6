package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type DbExplorer struct {
	db *sql.DB
}

func WriteRows(w http.ResponseWriter, rows *sql.Rows) error {
	for rows.Next() {
		var id int
		var title, description, updated string
		rows.Scan(&id, &title, &description, &updated)
		fmt.Fprintf(w, "%d%s%s%s\n", id, title, description, updated)
	}

	return rows.Err()
}

func (h *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	query := r.URL.Query()

	switch r.Method {
	case "GET":
		switch len(path) {
		case 0:
			rows, err := h.db.Query("SELECT * FROM sample_db")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			WriteRows(w, rows)
		case 1:
			limitInt, _ := strconv.Atoi(query.Get("limit"))
			offsetInt, _ := strconv.Atoi(query.Get("offset"))
			table := path[0]

			rows, err := h.db.Query(
				"SELECT * FROM ? LIMIT ? OFFSET ?",
				table, limitInt, offsetInt,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			WriteRows(w, rows)
		case 2:
			limitInt, _ := strconv.Atoi(query.Get("limit"))
			offsetInt, _ := strconv.Atoi(query.Get("offset"))
			table := path[0]
			id := path[1]

			rows, err := h.db.Query(
				"SELECT * FROM ? LIMIT ? OFFSET ? WHERE id=?",
				table, limitInt, offsetInt, id,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			WriteRows(w, rows)
		}
	case "PUT":
		r.ParseForm()
		title := r.Form.Get("title")
		description := r.Form.Get("description")
		updated := r.Form.Get("updated")
		table := path[0]

		s := fmt.Sprintf("INSERT INTO %s(title, description, updated) VALUES(?, ?, ?)", table)
		_, err := h.db.Exec(s, title, description, updated)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "POST":
		r.ParseForm()
		title := r.Form.Get("title")
		description := r.Form.Get("description")
		updated := r.Form.Get("updated")
		table := path[0]
		id := path[1]

		s := fmt.Sprintf("UPDATE %s SET title = ?, description = ?, updated = ? WHERE id = ?", table)
		_, err := h.db.Exec(s, title, description, updated, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "DELETE":
		table := path[0]
		id := path[1]

		s := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
		_, err := h.db.Exec(s, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func NewDbExplorer(db *sql.DB) (*DbExplorer, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	return &DbExplorer{db: db}, nil
}
