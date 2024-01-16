package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func server(c *gin.Context) {
	var status string
	var user string
	defer func() {
		userStatus.WithLabelValues(user, status).Inc()
	}()
	var mr MyRequest
	if err := c.BindJSON(&mr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if rand.Float32() > 0.8 {
		status = "4xx"
	} else {
		status = "2xx"
	}

	user = mr.User
	log.Println(user, status)
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func producer() {
	userPool := []string{"bob", "alice", "jack"}
	for {
		user := userPool[rand.Intn(len(userPool))]
		postBody, _ := json.Marshal(MyRequest{
			User: user,
		})
		requestBody := bytes.NewBuffer(postBody)

		// Gunakan "http://localhost:8080" untuk menggunakan server Gin Anda
		http.Post("http://localhost:8080", "application/json", requestBody)
		time.Sleep(time.Second * 2)
	}
}

func main() {
	go producer()

	r := gin.Default()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.POST("/", server)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
