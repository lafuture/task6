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

func WriteRow(w http.ResponseWriter, rows *sql.Rows) {
	if !rows.Next() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "record not found",
		})
		return
	}

	cols, _ := rows.Columns()

	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range cols {
		valPtrs[i] = &vals[i]
	}
	if err := rows.Scan(valPtrs...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{}
	for i, col := range cols {
		switch v := vals[i].(type) {
		case nil:
			resp[col] = nil
		case []byte:
			s := string(v)
			if s == "" && col == "updated" {
				resp[col] = nil
			} else {
				resp[col] = s
			}
		default:
			resp[col] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": map[string]interface{}{
			"record": resp,
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

func DecodePut(w http.ResponseWriter, r *http.Request) (interface{}, interface{}, interface{}) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var title, description, updated interface{}
	if v, ok := body["title"]; ok {
		title = v
	}
	if v, ok := body["description"]; ok {
		description = v
	}
	if v, ok := body["updated"]; ok {
		updated = v
	}

	return title, description, updated
}

func (h *DbExplorer) DecodePost(w http.ResponseWriter, r *http.Request, table, id string) (interface{}, interface{}, interface{}) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, nil
	}

	row := h.db.QueryRow(fmt.Sprintf("SELECT title, description, updated FROM %s WHERE id=?", table), id)
	var oldtitle, olddescription, oldupdated sql.NullString
	if err := row.Scan(&oldtitle, &olddescription, &oldupdated); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, nil
	}
	title, description := oldtitle.String, olddescription.String
	var updated interface{}

	if _, ok := body["id"]; ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "field id have invalid type",
		})
		return nil, nil, nil
	}
	if v, ok := body["title"]; ok {
		if s, ok := v.(string); ok {
			title = s
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "field title have invalid type",
			})
			return nil, nil, nil
		}
	}
	if v, ok := body["description"]; ok {
		if s, ok := v.(string); ok {
			description = s
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "field description have invalid type",
			})
			return nil, nil, nil
		}
	}
	if v, ok := body["updated"]; ok {
		if v == nil {
			updated = nil
		} else if s, ok := v.(string); ok {
			if s == "" {
				updated = nil
			} else {
				updated = s
			}
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "field updated have invalid type",
			})
			return nil, nil, nil
		}
	}

	return title, description, updated
}

func (h *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	query := r.URL.Query()

	if r.URL.Path == "/" {
		path = []string{}
	}

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
			limitInt, err := strconv.Atoi(query.Get("limit"))
			if err != nil || limitInt <= 0 {
				limitInt = 100
			}

			offsetInt, err := strconv.Atoi(query.Get("offset"))
			if err != nil || offsetInt < 0 {
				offsetInt = 0
			}

			table := path[0]

			if !h.TableExists(table) {
				SendError(w)
				return
			}

			s := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
			rows, err := h.db.Query(s, limitInt, offsetInt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer rows.Close()

			WriteRows(w, rows)

		case 2:
			id := path[1]
			intid, err := strconv.Atoi(id)
			if err != nil {
				SendError(w)
				return
			}

			table := path[0]

			if !h.TableExists(table) {
				SendError(w)
				return
			}

			s := fmt.Sprintf("SELECT * FROM %s WHERE id=?", table)
			rows, err := h.db.Query(s, intid)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer rows.Close()

			WriteRow(w, rows)
		}

	case "PUT":
		r.ParseForm()
		table := path[0]

		if !h.TableExists(table) {
			SendError(w)
			return
		}

		title, description, updated := DecodePut(w, r)

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

	case "POST":
		r.ParseForm()
		table := path[0]
		id := path[1]

		if !h.TableExists(table) {
			SendError(w)
			return
		}

		title, description, updated := h.DecodePost(w, r, table, id)
		if title == nil && description == nil && updated == nil {
			// была ошибка, уже отправлен JSON
			return
		}

		s := fmt.Sprintf("UPDATE %s SET title = ?, description = ?, updated = ? WHERE id = ?", table)
		res, err := h.db.Exec(s, title, description, updated, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		affected, _ := res.RowsAffected()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"response": map[string]interface{}{
				"updated": affected,
			},
		})

	case "DELETE":
		table := path[0]
		id := path[1]

		if !h.TableExists(table) {
			SendError(w)
			return
		}

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
	}
}

func NewDbExplorer(db *sql.DB) (*DbExplorer, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	return &DbExplorer{db: db}, nil
}
