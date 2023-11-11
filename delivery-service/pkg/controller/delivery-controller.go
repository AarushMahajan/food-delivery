package controller

import (
	"fmt"
	"net/http"

	dbConfig "github.com/AarushMahajan/food-delivery/delivery-service/pkg/config"
	"github.com/gin-gonic/gin"
)

type rider struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Is_reserved bool   `json:"isReserved"`
	Is_booked   bool   `json:"isBooked"`
	Order_id    int    `json:"orderId"`
}

type Order struct {
	OrderID int `json:"orderId"`
}

func GetAllAgents(c *gin.Context) {
	fmt.Println("GetAllAgents")
	db := dbConfig.GetDB()

	rows, err := db.Query("Select id, name, is_reserved from riders")
	// TODO: check why we have close this
	defer rows.Close()

	if err != nil {
		panic(err.Error())
	}

	riderJson := []rider{}
	// TODO: check any other good way to fetch data
	for rows.Next() {
		var agent rider

		err := rows.Scan(&agent.Id, &agent.Name, &agent.Is_reserved)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		riderJson = append(riderJson, agent)
	}

	c.JSON(http.StatusOK, riderJson)
}

func ReserveAgent(c *gin.Context) {

	db := dbConfig.GetDB()
	trx, _ := db.Begin()

	//find the non reserved agents
	row := trx.QueryRow("Select id, name, is_reserved from riders where is_reserved = 0 and order_id is null LIMIT 1 FOR UPDATE")

	var agent rider
	err := row.Scan(&agent.Id, &agent.Name, &agent.Is_reserved)
	if err != nil {
		trx.Rollback()
		fmt.Println("error ::: No delivery agent found!")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "No delivery agent found"})
		return
	}

	// reserved the agents
	_, err = trx.Exec("update riders set is_reserved = true where id = ?", agent.Id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	trx.Commit()
	fmt.Println("agent reserved successfully", agent.Id, agent.Name)

	c.JSON(200, gin.H{"message": "success"})
}

func BookAgent(c *gin.Context) {

	var order Order
	err := c.BindJSON(&order)
	if err != nil {
		fmt.Println("Error reading request body")
		c.AbortWithStatusJSON(500, "Error reading request body")
	}

	orderId := order.OrderID
	fmt.Println(orderId)

	db := dbConfig.GetDB()
	trx, _ := db.Begin()
	var agent rider

	// Find the reserved agent
	row := trx.QueryRow("Select id, name, is_reserved from riders where is_reserved = true and order_id is null limit 1 FOR UPDATE")

	err = row.Scan(&agent.Id, &agent.Name, &agent.Is_reserved)
	if err != nil {
		fmt.Print("Error scanning")
		c.AbortWithStatus(500)
		return
	}

	// update the reserved agent status to is_booked = true
	_, err = trx.Exec("Update riders set is_reserved = false, is_booked = true, order_id = ? where id = ?", orderId, agent.Id)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	trx.Commit()
	fmt.Println("Agent booked successfully")

	c.JSON(200, gin.H{"message": "success"})
}
