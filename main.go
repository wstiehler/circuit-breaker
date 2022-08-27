package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

var startTime time.Time = time.Now()

func server() {
	e := gin.Default()

	e.GET("/ping", func(ctx *gin.Context) {
		if time.Since(startTime) < 5*time.Second {
			ctx.String(http.StatusInternalServerError, "Server is too busy. Please try again later.")
			return
		}
		ctx.String(http.StatusOK, "pong")
	})

	fmt.Printf("Server is running on port: 8080 \n")
	e.Run(":8080")
}

func DoReq() error {
	resp, err := http.Get("http://localhost:8080/ping")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("Status code error: %d %s", resp.StatusCode, resp.Status))
	}

	return nil
}

func main() {
	go server()

	cb := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "circuit-breaker",
			MaxRequests: 3,
			Timeout:     3 * time.Second,
			Interval:    1 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				fmt.Printf("CircuitBreaker '%s' changed from %s to %s\n", name, from, to)
			},
		},
	)
	fmt.Println("CircuitBreaker is ready")
	for i := 0; i < 100; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			err := DoReq()
			return nil, err
		})
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
