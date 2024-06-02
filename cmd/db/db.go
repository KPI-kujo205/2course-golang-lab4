package main

import (
	"encoding/json"
	"flag"
	"github.com/KPI-kujo205/2course-golang-lab4/httptools"
	"github.com/KPI-kujo205/2course-golang-lab4/signal"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/KPI-kujo205/2course-golang-lab4/datastore"
)

var port = flag.Int("port", 8083, "server port")

type ResponseStructure struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RequestStructure struct {
	Value string `json:"value"`
}

func main() {
	flag.Parse()
	h := http.NewServeMux()

	dir, err := ioutil.TempDir("", "temp-dir")
	if err != nil {
		log.Fatal(err)
	}

	Db, err := datastore.NewDb(dir, 45)
	if err != nil {
		log.Fatal(err)
	}
	defer Db.Close()

	h.HandleFunc("/db/", func(rw http.ResponseWriter, req *http.Request) {
		dispatchRequestsForDb(Db, rw, req)
	})

	server := httptools.CreateServer(*port, h)
	go server.Start()

	signal.WaitForTerminationSignal()
}

func dispatchRequestsForDb(Db *datastore.Db, rw http.ResponseWriter, req *http.Request) {
	url := req.URL.String()
	key := url[4:]

	switch req.Method {
	case http.MethodGet:
		dispatchGetRequest(Db, rw, key)
	case http.MethodPost:
		dispatchPostRequest(Db, rw, req, key)
	default:
		rw.WriteHeader(http.StatusBadRequest)
	}
}

func dispatchGetRequest(Db *datastore.Db, rw http.ResponseWriter, key string) {
	value, err := Db.Get(key)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(rw).Encode(ResponseStructure{Key: key, Value: value}); err != nil {
		log.Println("An error occurred while encoding response: ", err)
	}
}

func dispatchPostRequest(Db *datastore.Db, rw http.ResponseWriter, req *http.Request, key string) {
	var body RequestStructure

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := Db.Put(key, body.Value); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}
