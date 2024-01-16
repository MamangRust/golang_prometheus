## Golang Prometheus

In this tutorial, we will cover the integration of a Golang application with Prometheus and Grafana for monitoring and visualization. The example Golang application will expose metrics that Prometheus will scrape, and Grafana will be used to create dashboards.

## Golang Application Metrics

```go
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


```

## Docker-compose configuration

```yaml
version: '3'

services:
  prometheus:
    image: prom/prometheus:v2.49.1
    container_name: prometheus
    ports:
      - '9090:9090'
    volumes:
      - ./prometheus:/etc/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
    networks:
      - monitoring

  grafana:
    image: grafana/grafana-oss
    container_name: grafana
    ports:
      - '3000:3000'
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
    networks:
      - monitoring
    depends_on:
      - prometheus

  golang-app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: golang-app
    ports:
      - '8080:8080'
    networks:
      - monitoring
    depends_on:
      - prometheus

networks:
  monitoring:

volumes:
  grafana-storage:
```

## Prometheus Configuration

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'golang-app'
    static_configs:
      - targets: ['golang-app:8080']
```

## Grafana Dashboard Setup

Access Grafana at http://localhost:3000 with the username admin and password admin.
Add Prometheus as a data source using http://prometheus:9090.
Import a pre-configured Golang metrics dashboard from the Grafana dashboard marketplace or create a custom dashboard.

![Alt text](/images/grafana.png)

## Demo

![Demo](/images/image.png)
