package main

import (
	"fmt"
	"log"

	configDb "github.com/AarushMahajan/food-delivery/inventory-service/pkg/config"
	router "github.com/AarushMahajan/food-delivery/inventory-service/pkg/routes"
)

func main() {
	r := router.Routes()
	configDb.Connect()
	fmt.Println("Inventory Service starting on port 3000....")
	log.Fatal(r.Run(":3000"))
}
