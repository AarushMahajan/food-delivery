package main

import (
	"fmt"
	"log"

	configDb "github.com/AarushMahajan/food-delivery/order-service/pkg/config"
	router "github.com/AarushMahajan/food-delivery/order-service/pkg/routes"
)

func main() {
	r := router.Routes()
	configDb.Connect()
	fmt.Println("Order Service starting on port 8000....")
	log.Fatal(r.Run(":8000"))
}
