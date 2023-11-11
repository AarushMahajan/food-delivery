package routes

import (
	"fmt"

	controller "github.com/AarushMahajan/food-delivery/inventory-service/pkg/controller"
	"github.com/gin-gonic/gin"
)

func welcomeRouter(c *gin.Context) {
	fmt.Println("This is homepage of invenotry")
	c.JSON(200, gin.H{"message": "This is invenotry service!!"})
}

func Routes() *gin.Engine {
	r := gin.Default()

	r.GET("/", welcomeRouter)
	r.GET("/list-food", controller.GetAllInventory)
	r.POST("/reserve-food", controller.ReserveItem)
	r.POST("/book-food", controller.BookItem)

	return r
}
