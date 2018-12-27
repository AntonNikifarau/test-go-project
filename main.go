package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var table = `
	CREATE TABLE IF NOT EXISTS user (
	id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY,
	firstname VARCHAR(30) NOT NULL,
	lastname VARCHAR(30) NOT NULL
)
`

var clearUsers = `DELETE FROM user;`

var userData = `
INSERT INTO user (firstname, lastname) VALUES ("nick", "kavtur"),
					      ("alex", "bondarev"),
					      ("pavel", "belkevich");
`

type User struct {
	ID        int    `db:"id"`
	FirstName string `db:"firstname"`
	LastName  string `db:"lastname"`
}

func main() {
	db, err := sqlx.Connect("mysql", "root:root@/demo")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(table)
	db.MustExec(clearUsers)
	db.MustExec(userData)

	r := mux.NewRouter()

	r.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadFile("./users.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t, err := template.New("sas").Parse(string(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var users []User
		if err := db.Select(&users, "SELECT * FROM user"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, users); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	r.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			http.Error(w, "invalid id", http.StatusInternalServerError)
			return
		}
		code, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "could not cast id to int", http.StatusInternalServerError)
			return
		}
		var user User
		if err := db.Get(&user, "SELECT * FROM user WHERE id = ?", code); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := ioutil.ReadFile("./edit.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t, err := template.New("sas").Parse(string(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	r.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id := r.PostFormValue("id")
		code, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		name := r.PostFormValue("firstname")
		lastName := r.PostFormValue("lastname")
		user := User{ID:code, FirstName: name, LastName: lastName}
		_, err = db.NamedExec("UPDATE user SET firstname=:firstname, lastname=:lastname WHERE id=:id", &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Location", "/user")
		w.WriteHeader(301)
	})

	http.ListenAndServe(":8081", r)

}
