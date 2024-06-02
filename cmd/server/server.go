package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/KPI-kujo205/2course-golang-lab4/httptools"
	"github.com/KPI-kujo205/2course-golang-lab4/signal"
)

var port = flag.Int("port", 8080, "server port")
var report = make(Report)

const (
	confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
	confHealthFailure    = "CONF_HEALTH_FAILURE"
	urlForDb             = "http://db:8083/db/"
)

type RequestStructure struct {
	Value string `json:"value"`
}

type ResponseStructure struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	flag.Parse()
	h := http.NewServeMux()
	client := http.DefaultClient

	h.HandleFunc("/health", handleHealth)
	h.HandleFunc("/api/v1/some-data", handleDataRequest(client))
	h.HandleFunc("/api/v1/some-data2", handleStaticDataTwo)
	h.HandleFunc("/api/v1/some-data3", handleStaticDataThree)

	h.Handle("/report", report)

	server := httptools.CreateServer(*port, h)
	go server.Start()

	buff := new(bytes.Buffer)
	body := RequestStructure{Value: time.Now().Format(time.RFC3339)}
	json.NewEncoder(buff).Encode(body)

	res, err := client.Post(fmt.Sprintf("%sgoyda", urlForDb), "application/json", buff)
	if err != nil {
		log.Fatalf("Invalid POST operation: %v", err)
	}

	defer res.Body.Close()

	signal.WaitForTerminationSignal()
}

func handleHealth(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("content-type", "text/plain")
	if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte("FAILURE"))
	} else {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("OK"))
	}
}

func handleDataRequest(client *http.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := client.Get(fmt.Sprintf("%s/%s", urlForDb, key))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		statusOk := resp.StatusCode >= 200 && resp.StatusCode < 300

		if !statusOk {
			rw.WriteHeader(resp.StatusCode)
			return
		}

		applyDelay()
		report.Process(r)

		var body ResponseStructure
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()

		handleJsonResponse(rw, body)
	}
}

func handleStaticDataTwo(rw http.ResponseWriter, r *http.Request) {
	applyDelay()
	report.Process(r)

	handleJsonResponse(rw, []string{"2", "3"})
}

func handleStaticDataThree(rw http.ResponseWriter, r *http.Request) {
	applyDelay()
	report.Process(r)

	handleJsonResponse(rw, []string{"3", "4"})
}

func handleJsonResponse(rw http.ResponseWriter, body interface{}) {
	rw.Header().Set("content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(rw).Encode(body)
}

func applyDelay() {
	respDelayString := os.Getenv(confResponseDelaySec)
	if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
		time.Sleep(time.Duration(delaySec) * time.Second)
	}
}
