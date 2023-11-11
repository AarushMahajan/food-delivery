package main

import (
	"fmt"
	"log"

	dbConfig "github.com/AarushMahajan/food-delivery/delivery-service/pkg/config"
	routes "github.com/AarushMahajan/food-delivery/delivery-service/pkg/routes"
)

func main() {
	r := routes.Router()
	dbConfig.Connect()
	fmt.Println("Delivery Service started on port 4000...")
	log.Fatal(r.Run(":4000"))
}
