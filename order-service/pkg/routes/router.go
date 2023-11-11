package routes

import (
	"fmt"

	controller "github.com/AarushMahajan/food-delivery/order-service/pkg/controller"
	"github.com/gin-gonic/gin"
)

func welcomeRouter(c *gin.Context) {
	fmt.Println("This is homepage of OrderService")
	c.JSON(200, gin.H{"message": "This is Order service!!"})
}

func Routes() *gin.Engine {
	r := gin.Default()

	r.GET("/", welcomeRouter)
	r.POST("/place-order", controller.PlaceOrder)

	return r
}
