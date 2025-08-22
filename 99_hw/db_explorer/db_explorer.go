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
			rows, err := h.db.Query("SHOW TABLES")
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

			flag := false
			allrows, err := h.db.Query("SHOW TABLES")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer allrows.Close()
			for allrows.Next() {
				var tableName string
				err := allrows.Scan(&tableName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				if tableName == table {
					flag = true
					break
				}
			}

			if flag {
				s := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
				rows, err := h.db.Query(s, limitInt, offsetInt)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer rows.Close()

				WriteRows(w, rows)
			} else {
				http.Error(w, "unknown table", http.StatusBadRequest)
				return
			}

		case 2:
			limitInt, _ := strconv.Atoi(query.Get("limit"))
			offsetInt, _ := strconv.Atoi(query.Get("offset"))
			table := path[0]
			id := path[1]

			flag := false
			allrows, err := h.db.Query("SHOW TABLES")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer allrows.Close()
			for allrows.Next() {
				var tableName string
				err := allrows.Scan(&tableName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				if tableName == table {
					flag = true
					break
				}
			}

			if flag {
				s := fmt.Sprintf("SELECT * FROM %s WHERE id=? LIMIT ? OFFSET ?", table)
				rows, err := h.db.Query(s, id, limitInt, offsetInt)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer rows.Close()

				WriteRows(w, rows)
			} else {
				http.Error(w, "unknown table", http.StatusBadRequest)
			}

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

		flag := false
		allrows, err := h.db.Query("SHOW TABLES")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer allrows.Close()
		for allrows.Next() {
			var tableName string
			err := allrows.Scan(&tableName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if tableName == table {
				flag = true
				break
			}
		}

		if flag {
			s := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
			_, err := h.db.Exec(s, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else {
				http.Error(w, "unknown table", http.StatusBadRequest)
			}
		}
	}

}

func NewDbExplorer(db *sql.DB) (*DbExplorer, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	return &DbExplorer{db: db}, nil
}

//func (h *DbExplorer) CaseGet(w http.ResponseWriter, path []string, query url.Values) {
//	switch len(path) {
//	case 0:
//		rows, err := h.db.Query("SHOW TABLES")
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		defer rows.Close()
//
//		WriteRows(w, rows)
//	case 1:
//		limitInt, _ := strconv.Atoi(query.Get("limit"))
//		offsetInt, _ := strconv.Atoi(query.Get("offset"))
//		table := path[0]
//
//		s := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
//		rows, err := h.db.Query(s, limitInt, offsetInt)
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		defer rows.Close()
//
//		WriteRows(w, rows)
//	case 2:
//		limitInt, _ := strconv.Atoi(query.Get("limit"))
//		offsetInt, _ := strconv.Atoi(query.Get("offset"))
//		table := path[0]
//		id := path[1]
//
//		s := fmt.Sprintf("SELECT * FROM %s WHERE id=? LIMIT ? OFFSET ?", table)
//		rows, err := h.db.Query(s, id, limitInt, offsetInt)
//
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		defer rows.Close()
//
//		WriteRows(w, rows)
//	}
//}
//
//func (h *DbExplorer) CasePut(w http.ResponseWriter, r *http.Request, path []string) {
//	r.ParseForm()
//	title := r.Form.Get("title")
//	description := r.Form.Get("description")
//	updated := r.Form.Get("updated")
//	table := path[0]
//
//	s := fmt.Sprintf("INSERT INTO %s(title, description, updated) VALUES(?, ?, ?)", table)
//	_, err := h.db.Exec(s, title, description, updated)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//}
//
//func (h *DbExplorer) CasePost(w http.ResponseWriter, r *http.Request, path []string) {
//	r.ParseForm()
//	title := r.Form.Get("title")
//	description := r.Form.Get("description")
//	updated := r.Form.Get("updated")
//	table := path[0]
//	id := path[1]
//
//	s := fmt.Sprintf("UPDATE %s SET title = ?, description = ?, updated = ? WHERE id = ?", table)
//	_, err := h.db.Exec(s, title, description, updated, id)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//}
//
//func (h *DbExplorer) CaseDelete(w http.ResponseWriter, r *http.Request, path []string) {
//	table := path[0]
//	id := path[1]
//
//	s := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
//	_, err := h.db.Exec(s, id)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//}
