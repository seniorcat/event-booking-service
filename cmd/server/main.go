package main

import (
	"fmt"
	"log"
	"net/http"

	myhttp "laschool.ru/event-booking-service/internal/http"
)

func main() {
	fmt.Println("Booking Service started...")

	router := myhttp.NewRouter()

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
