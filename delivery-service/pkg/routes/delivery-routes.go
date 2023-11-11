package routes

import (
	"net/http"

	controller "github.com/AarushMahajan/food-delivery/delivery-service/pkg/controller"
	"github.com/gin-gonic/gin"
)

func welcomeRouter(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"name": "delivery-service"})
}

func Router() *gin.Engine {

	r := gin.Default()

	r.GET("/", welcomeRouter)
	r.GET("/all-agents", controller.GetAllAgents)
	r.POST("/reserve-agent", controller.ReserveAgent)
	r.POST("/book-agent", controller.BookAgent)

	return r
}
