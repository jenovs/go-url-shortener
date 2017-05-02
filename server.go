package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jenovs/api-url-shortener/config"
	"github.com/speps/go-hashids"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var db *sql.DB
var err error

type urlError struct {
	Error string `json:"error"`
}

type Response struct {
	Url   string `json:"original_url"`
	Short string `json:"short_url"`
}

func getHash(str string) string {
	hd := hashids.NewData()
	hd.Salt = str
	h := hashids.NewWithData(hd)
	id, _ := h.Encode([]int{1, 2, 3})
	return id
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}
	return ":" + port
}

func createUrlTable() {
	create_table := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL NOT NULL PRIMARY KEY,
			url TEXT NOT NULL,
			short TEXT NOT NULL
		);
	`
	stmt, err := db.Prepare(create_table)
	checkErr(err)

	_, err = stmt.Exec()
	checkErr(err)
}

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[1:]

	if len(url) == 0 {
		http.ServeFile(w, r, "./index.html")
	} else if url == "favicon.ico" {
		http.ServeFile(w, r, "./favicon.ico")
	} else if p := strings.Split(url, "/"); p[0] == "new" {
		var addr string
		if len(url) > 4 {
			addr = url[4:]
		}
		short, err := createShort(addr)
		if err != nil {
			var urlErr urlError
			urlErr.Error = err.Error()
			res, _ := json.Marshal(urlErr)
			w.Header().Set("Content-Type", "application/json")
			w.Write(res)
			return
		}
		var response Response
		response.Url = addr
		response.Short = "http://127.0.0.1:3000/" + short
		res, _ := json.Marshal(response)

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		return
	} else {
		res, err := getFromShort(url)
		if err == nil {

			http.Redirect(w, r, res, http.StatusSeeOther)
		}
	}
}

func getFromShort(s string) (string, error) {
	rows, err := db.Query("SELECT * FROM urls WHERE short=?", s)
	checkErr(err)

	var url string

	for rows.Next() {
		var id string
		var short string
		err = rows.Scan(&id, &url, &short)
	}

	if len(url) > 0 {
		return url, nil
	}
	return "", nil
}

func createShort(s string) (string, error) {
	matched, err := regexp.MatchString("^(http)[s]?(://).+\\.[a-zA-Z]{2,}$", s)
	checkErr(err)

	if !matched {
		return "", errors.New("Invalid url")
	}
	short := getHash(s)

	// insert into db
	stmt, err := db.Prepare("INSERT INTO urls SET id=?, url=?, short=?")
	checkErr(err)

	_, err = stmt.Exec(0, s, short)
	checkErr(err)

	return short, nil
}

func main() {
	// connect to db
	db_user := os.Getenv("db_user")
	db_pass := os.Getenv("db_pass")
	db_addr := os.Getenv("db_addr")
	db_name := os.Getenv("db_name")
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s?charset=utf8", db_user, db_pass, db_addr, db_name))
	checkErr(err)

	defer db.Close()

	err = db.Ping()
	checkErr(err)

	// create table `urls` if it doesn't exist
	if _, err := db.Query("SELECT 1 FROM urls LIMIT 1"); err != nil {
		createUrlTable()
	}

	handler := new(Handler)
	log.Fatal(http.ListenAndServe(getPort(), handler))
}
