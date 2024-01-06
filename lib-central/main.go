package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type User struct {
	JMBG              string
	Name              string
	Address           string
	NumberOfBooksLent int
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
	log.Printf("creating users table")
	rows, err := db.Query(`
	CREATE TABLE IF NOT EXISTS users (
		jmbg				varchar(13) PRIMARY KEY NOT NULL,
		name				varchar(255) NOT NULL,
		address	 			varchar(255),
		num_of_books_lent	smallint NOT NULL
	);
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Printf("users table created")
}

func main() {
	log.Printf("starting http server on port %s\n", os.Getenv("PORT"))
	http.HandleFunc("/user", usersHandler)
	http.HandleFunc("/user/lending", usersLendHandler)
	http.HandleFunc("/health", healthcheck)

	http.ListenAndServe(os.Getenv("PORT"), nil)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + " " + r.URL.Path)
	if r.Method == http.MethodGet {

		jmbg := r.FormValue("jmbg")
		// if len(jmbg) != 13 {
		// 	log.Println("jmbg not valid")
		// }

		row := db.QueryRow("SELECT * FROM users WHERE jmbg = $1", jmbg)

		usr := new(User)
		err := row.Scan(&usr.JMBG, &usr.Name, &usr.Address, &usr.NumberOfBooksLent)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		fmt.Fprintf(w, "%s,%s,%s,%v\n", usr.JMBG, usr.Name, usr.Address, usr.NumberOfBooksLent)
	} else if r.Method == http.MethodPost {
		log.Println("processing POST request")

		jmbg := r.FormValue("jmbg")
		name := r.FormValue("name")
		address := r.FormValue("address")

		if !userExists(jmbg) {
			http.Error(w, http.StatusText(400), 400)
		}
		result, err := db.Exec("INSERT INTO users(jmbg, name, address, num_of_books_lent) VALUES($1, $2, $3, 0)", jmbg, name, address)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "User %s created successfully (%d row affected)\n", name, rowsAffected)
	} else {
		http.NotFound(w, r)
	}
}

func usersLendHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s\n", r.Method, r.URL.Path)
	jmbg := r.FormValue("jmbg")
	if !userExists(jmbg) {
		log.Println("user with given JMBG does not exist")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodPost {

		result, err := db.Exec("UPDATE users SET num_of_books_lent = num_of_books_lent + 1 WHERE jmbg = $1 AND num_of_books_lent < 3", jmbg)
		if err != nil {
			log.Printf("could not update the number of books lent: %s\n", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if rowsAffected != 1 {
			log.Println("did not increment 1 record")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		fmt.Fprintf(w, "User with JMBG %s updated successfully\n", jmbg)
		return
	} else if r.Method == http.MethodDelete {

		result, err := db.Exec("UPDATE users SET num_of_books_lent = num_of_books_lent - 1 WHERE jmbg = $1 AND num_of_books_lent > 0", jmbg)
		if err != nil {
			log.Printf("could not update the number of books lent: %s\n", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if rowsAffected != 1 {
			log.Println("did not increment 1 record")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		fmt.Fprintf(w, "User with JMBG %s updated successfully\n", jmbg)
		return
	} else {
		http.NotFound(w, r)
		return
	}
}

func incrementNumberOfBooksLent(increment int, jmbg string) error {
	incrementString := strconv.Itoa(increment)
	log.Println("(inc/dec)rementing number of books lent")

	if !userExists(jmbg) {
		log.Println("user with given JMBG does not exist")
		return errors.New("user with given jmbg does not exist")
	}

	result, err := db.Exec("UPDATE users SET num_of_books_lent = num_of_books_lent + ("+incrementString+") WHERE jmbg = $1 AND num_of_books_lent < 3 AND num_of_books_lent > 0", jmbg)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if rowsAffected != 1 {
		log.Println("did not increment 1 record")
		return errors.New("did not increment 1 record")
	}
	return nil
}

func userExists(jmbg string) bool {
	_, err := db.Query("select * from users where jmbg = $1", jmbg)
	if err != nil {
		return false
	}
	return true
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "OK")
}
