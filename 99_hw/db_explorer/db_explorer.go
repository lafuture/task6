package main

import (
	"database/sql"
	"encoding/json"
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

func WriteTables(w http.ResponseWriter, rows *sql.Rows) {
	resp := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		resp = append(resp, tableName)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": map[string]interface{}{
			"tables": resp,
		},
	})
}

func WriteRows(w http.ResponseWriter, rows *sql.Rows) {
	resp := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var title, description string
		var updated *string
		if err := rows.Scan(&id, &title, &description, &updated); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp = append(resp, map[string]interface{}{"id": id, "title": title, "description": description, "updated": updated})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": map[string]interface{}{
			"records": resp,
		},
	})
}

func (h *DbExplorer) TableExists(table string) bool {
	rows, err := h.db.Query("SHOW TABLES")
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		if tableName == table {
			return true
		}
	}
	return false
}

func SendError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "unknown table",
	})
}

func (h *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	query := r.URL.Query()

	switch r.Method {
	case "GET":
		switch len(path) {
		case 0:
			rows, err := h.db.Query("SHOW TABLES")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer rows.Close()

			WriteTables(w, rows)
		case 1:
			limitInt, _ := strconv.Atoi(query.Get("limit"))
			offsetInt, _ := strconv.Atoi(query.Get("offset"))
			table := path[0]

			if h.TableExists(table) {
				s := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
				rows, err := h.db.Query(s, limitInt, offsetInt)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				defer rows.Close()

				WriteRows(w, rows)
			} else {
				SendError(w)
			}

		case 2:
			limitInt, _ := strconv.Atoi(query.Get("limit"))
			offsetInt, _ := strconv.Atoi(query.Get("offset"))
			table := path[0]
			id := path[1]

			if h.TableExists(table) {
				s := fmt.Sprintf("SELECT * FROM %s WHERE id=? LIMIT ? OFFSET ?", table)
				rows, err := h.db.Query(s, id, limitInt, offsetInt)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				defer rows.Close()

				WriteRows(w, rows)
			} else {
				SendError(w)
			}

		}
	case "PUT":
		r.ParseForm()
		table := path[0]

		if h.TableExists(table) {
			title := r.Form.Get("title")
			description := r.Form.Get("description")
			updated := r.Form.Get("updated")
			s := fmt.Sprintf("INSERT INTO %s(title, description, updated) VALUES(?, ?, ?)", table)
			res, err := h.db.Exec(s, title, description, updated)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			lastID, _ := res.LastInsertId()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"response": map[string]interface{}{
					"id": lastID,
				},
			})
		} else {
			SendError(w)
		}
	case "POST":
		r.ParseForm()
		table := path[0]
		id := path[1]

		if h.TableExists(table) {
			title := r.Form.Get("title")
			description := r.Form.Get("description")
			updated := r.Form.Get("updated")
			s := fmt.Sprintf("UPDATE %s SET title = ?, description = ?, updated = ? WHERE id = ?", table)
			res, err := h.db.Exec(s, title, description, updated, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			affected, _ := res.RowsAffected()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"response": map[string]interface{}{
					"updated": affected,
				},
			})
		} else {
			SendError(w)
		}

	case "DELETE":
		table := path[0]
		id := path[1]

		if h.TableExists(table) {
			s := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
			res, err := h.db.Exec(s, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			deleted, _ := res.RowsAffected()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"response": map[string]interface{}{
					"deleted": deleted,
				},
			})
		} else {
			SendError(w)
		}
	}
}

func NewDbExplorer(db *sql.DB) (*DbExplorer, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	return &DbExplorer{db: db}, nil
}
