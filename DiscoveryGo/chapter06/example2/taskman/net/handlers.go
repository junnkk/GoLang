package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"text/template"

	"github.com/jaeyeom/gogo/task"

	"github.com/gorilla/mux"
)

const (
	apiPathPrefix  = "/api/v1/task/"
	htmlPathPrefix = "/task/"
	idPattern      = "/{id:[0-9]+}"
)

// FIXME: m is NOT thread-safe.
var m = task.NewInMemoryAccessor()

func getTasks(r *http.Request) ([]task.Task, error) {
	var result []task.Task
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	encodedTasks, ok := r.PostForm["task"]
	if !ok {
		return nil, errors.New("task parameter expected")
	}
	for _, encodedTask := range encodedTasks {
		var t task.Task
		if err := json.Unmarshal([]byte(encodedTask), &t); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func apiGetHandler(w http.ResponseWriter, r *http.Request) {
	id := task.ID(mux.Vars(r)["id"])
	t, err := m.Get(id)
	err = json.NewEncoder(w).Encode(Response{
		ID:    id,
		Task:  t,
		Error: ResponseError{Err: err},
	})
	if err != nil {
		log.Println(err)
	}
	// w.WriteHeader(400) - can handle error code
}

func apiPutHandler(w http.ResponseWriter, r *http.Request) {
	id := task.ID(mux.Vars(r)["id"])
	tasks, err := getTasks(r)
	if err != nil {
		log.Println(err)
		return
	}
	for _, t := range tasks {
		err = m.Put(id, t)
		err = json.NewEncoder(w).Encode(Response{
			ID:    id,
			Task:  t,
			Error: ResponseError{Err: err},
		})
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func apiPostHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := getTasks(r)
	if err != nil {
		log.Println(err)
		return
	}
	for _, t := range tasks {
		id, err := m.Post(t)
		err = json.NewEncoder(w).Encode(Response{
			ID:    id,
			Task:  t,
			Error: ResponseError{Err: err},
		})
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func apiDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := task.ID(mux.Vars(r)["id"])
	err := m.Delete(id)
	err = json.NewEncoder(w).Encode(Response{
		ID:    id,
		Error: ResponseError{Err: err},
	})
	if err != nil {
		log.Println(err)
		return
	}
}

var tmpl = template.Must(template.ParseGlob("html/*.html"))

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	id := task.ID(mux.Vars(r)["id"])
	t, err := m.Get(id)
	err = tmpl.ExecuteTemplate(w, "task.html", &Response{
		ID:    id,
		Task:  t,
		Error: ResponseError{Err: err},
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func main() {
	r := mux.NewRouter()
	r.PathPrefix(htmlPathPrefix).
		Path(idPattern).
		Methods("GET").
		HandlerFunc(htmlHandler)

	s := r.PathPrefix(apiPathPrefix).Subrouter()
	s.HandleFunc(idPattern, apiGetHandler).Methods("GET")
	s.HandleFunc(idPattern, apiPutHandler).Methods("PUT")
	s.HandleFunc("/", apiPostHandler).Methods("POST")
	s.HandleFunc(idPattern, apiDeleteHandler).Methods("DELETE")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8887", nil))
}
