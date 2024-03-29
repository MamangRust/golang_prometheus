package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var userStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_user_status_count",
		Help: "Count of status returned by user.",
	},
	[]string{"user", "status"},
)

func init() {
	prometheus.MustRegister(userStatus)
}

type MyRequest struct {
	User string `json:"user"`
}

func server(w http.ResponseWriter, r *http.Request) {
	var status string
	var user string
	defer func() {
		userStatus.WithLabelValues(user, status).Inc()
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var mr MyRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if rand.Float32() > 0.8 {
		status = "4xx"
	} else {
		status = "2xx"
	}

	user = mr.User
	log.Println(user, status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func producer() {
	userPool := []string{"bob", "alice", "jack"}
	for {
		user := userPool[rand.Intn(len(userPool))]
		postBody, _ := json.Marshal(MyRequest{
			User: user,
		})
		requestBody := bytes.NewBuffer(postBody)

		http.Post("http://localhost:8080", "application/json", requestBody)
		time.Sleep(time.Second * 2)
	}
}

func main() {
	go producer()

	http.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	http.HandleFunc("/", server)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
