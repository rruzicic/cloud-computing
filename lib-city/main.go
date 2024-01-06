package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	_ "github.com/lib/pq"
)

type BookLend struct {
	LenderId string
	BookName string
	ISBN     string
	Author   string
	LendDate string
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Printf("creating book_lends table")
	rows, err := db.Query(`
	CREATE TABLE IF NOT EXISTS book_lends (
		jmbg				varchar(13) NOT NULL,
		book_name			varchar(255) NOT NULL,
		isbn	 			varchar(255),
		author	 			varchar(255),
		lend_date			DATE DEFAULT CURRENT_DATE
	);
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Printf("book_lends table created")
}

func main() {
	log.Printf("starting http server on port %s\n", os.Getenv("PORT"))
	http.HandleFunc("/books/lending", booksLendHandler)
	http.HandleFunc("/health", healthcheck)
	http.ListenAndServe(os.Getenv("PORT"), nil)
}

func booksLendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		jmbg := r.FormValue("jmbg")
		bookName := r.FormValue("bookname")
		isbn := r.FormValue("isbn")
		author := r.FormValue("author")

		params := url.Values{}
		params.Add("jmbg", jmbg)
		url := os.Getenv("CENTRAL_LIB_BASE_URL") + "/user/lending"
		resp, err := http.PostForm(url, params)
		log.Println("sending request POST " + url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != 200 {
			log.Printf("could not lend a book, received a response with status code: %d\n", resp.StatusCode)
			http.Error(w, "could not lend a book", http.StatusInternalServerError)
			return
		}

		result, err := db.Exec("INSERT INTO book_lends(jmbg, book_name, isbn, author) VALUES($1, $2, $3, $4)", jmbg, bookName, isbn, author)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Book %s successfully lent (%d row affected)\n", bookName, rowsAffected)
	} else if r.Method == http.MethodDelete {
		log.Println("giving the book back to the library")
		jmbg := r.FormValue("jmbg")
		isbn := r.FormValue("isbn")

		result, err := db.Exec("DELETE FROM book_lends WHERE jmbg = $1 AND isbn = $2", jmbg, isbn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rowsAffected != 1 {
			log.Println("rows affected not 1")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		url := os.Getenv("CENTRAL_LIB_BASE_URL") + "/user/lending?jmbg=" + jmbg
		log.Printf("sending request DELETE %s\n", url)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("could not give the book back: got %d response code\n", resp.StatusCode)
			http.Error(w, "could not give the book back", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Book with ISBN: %s successfully returned (%d row affected)\n", isbn, rowsAffected)

	} else {
		http.NotFound(w, r)
		return
	}
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "OK")
}
