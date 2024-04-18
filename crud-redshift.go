package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// courses struct

type Course struct {
	ID           int    `json:"course_id"`
	Name         string `json:"course_name"`
	Prerequisite string `json:"course_prerequisite"`
}

var db *sql.DB
var err error

func main() {

	// Connect to the database
	db, err = sql.Open("postgres", "host=redshift-cluster-1.cjzjw2l4zj1o.us-west-2.redshift.amazonaws.com dbname=redshift port=5439 user=awsuser password=Password1 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Verify DB connection
	err = db.Ping()
	if err != nil {

	}

	// Set up routes
	r := mux.NewRouter()

	// Route handles & endpoints
	r.HandleFunc("/courses", getCourses).Methods("GET")
	r.HandleFunc("/courses/{id}", getCourse).Methods("GET")
	r.HandleFunc("/courses", updateCourse).Methods("PUT")

	http.ListenAndServe(":8080", r)

}

func getCourses(w http.ResponseWriter, r *http.Request) {
	var courses []Course
	rows, err := db.Query("SELECT course_id, course_name, course_prerequisite FROM courses")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer rows.Close()

	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.Name, &c.Prerequisite); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		courses = append(courses, c)
	}
	json.NewEncoder(w).Encode(courses)
}

func getCourse(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var c Course
	err = db.QueryRow("SELECT course_id, course_name, course_prerequisite FROM courses WHERE course_id = $1", id).Scan(&c.ID, &c.Name, &c.Prerequisite)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(c)
}

func updateCourse(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var c Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = db.Exec("UPDATE courses SET course_name = $2, course_prerequisite = $3 WHERE course_id = $1", id, c.Name, c.Prerequisite)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(c)
}
