package main

import (
	"fmt"
	"log"
	"net/http"

	myhttp "laschool.ru/event-booking-service/internal/http"
	"laschool.ru/event-booking-service/internal/http/middleware"
)

func main() {
	fmt.Println("Booking Service started...")

	router := myhttp.NewRouter()
	loggingRouter := middleware.LoggingMiddleware(router)

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	err := http.ListenAndServe(addr, loggingRouter)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
