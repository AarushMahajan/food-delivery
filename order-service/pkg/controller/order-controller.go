package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	configDb "github.com/AarushMahajan/food-delivery/order-service/pkg/config"
	"github.com/gin-gonic/gin"
)

type orderBody struct {
	Id       int64 `json:"id"`
	Quantity int64 `json:"quantity"`
}

type apiResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

var agentReserveUrl = "http://localhost:4000/reserve-agent"
var foodReserveUrl = "http://localhost:3000/reserve-food"
var agentBookUrl = "http://localhost:4000/book-agent"
var foodBookUrl = "http://localhost:3000/book-food"

func reserveResourcesAndBookOrder(ch chan apiResponse, wg *sync.WaitGroup, url string, rqBody []byte) {
	defer wg.Done()
	var responseBody apiResponse

	response, err := http.Post(url, "application/json", bytes.NewBuffer(rqBody))

	body, _ := io.ReadAll(response.Body)
	err = json.Unmarshal(body, &responseBody)

	defer response.Body.Close()
	if err != nil || response.StatusCode != 200 {
		responseBody.StatusCode = 500
		ch <- responseBody
		return
	}
	responseBody.StatusCode = 200
	ch <- responseBody
}

func PlaceOrder(c *gin.Context) {
	var body []orderBody
	err := c.BindJSON(&body)
	if err != nil {
		fmt.Println("Error while binding order body")
		c.AbortWithStatus(500)
		return
	}

	var wg sync.WaitGroup
	ch := make(chan apiResponse)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			placeOrderAPI(body, ch)
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for data := range ch {
		if data.StatusCode != 200 {
			c.AbortWithStatusJSON(412, gin.H{"message": "Order not placed"})
			return
		}
	}
	c.Status(200)
}

func placeOrderAPI(body []orderBody, ch chan apiResponse) apiResponse {
	reqBody, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error while Marshaling order body")
		chResp := apiResponse{StatusCode: 500}
		ch <- chResp
		return chResp
	}

	db := configDb.GetDB()
	// Reserve resources
	reserveResourceChannel := make(chan apiResponse)
	var wg sync.WaitGroup

	wg.Add(2)
	go reserveResourcesAndBookOrder(reserveResourceChannel, &wg, agentReserveUrl, nil)
	go reserveResourcesAndBookOrder(reserveResourceChannel, &wg, foodReserveUrl, reqBody)

	go func() {
		wg.Wait()
		close(reserveResourceChannel)
	}()

	for data := range reserveResourceChannel {
		if data.StatusCode != 200 {
			fmt.Println("Order not placed error: ", data.Message)
			chResp := apiResponse{Message: "Order not placed", StatusCode: 412}
			ch <- chResp
			return chResp
		}
	}

	bookOrderChannel := make(chan apiResponse)
	wg.Add(2)

	go reserveResourcesAndBookOrder(bookOrderChannel, &wg, agentBookUrl, nil)
	go reserveResourcesAndBookOrder(bookOrderChannel, &wg, foodBookUrl, reqBody)

	go func() {
		wg.Wait()
		close(bookOrderChannel)
	}()

	for data := range reserveResourceChannel {
		if data.StatusCode != 200 {
			fmt.Println("Order not placed error: ", data.Message)
			chResp := apiResponse{Message: "Order not placed", StatusCode: 412}
			ch <- chResp
			return chResp
		}
	}

	_, err = db.Exec("Insert into orders (username, created_at, updated_at) values (?,?,?)", "Aarush", time.Now(), time.Now())
	if err != nil {
		fmt.Println("error inserting order: ", err)
		chResp := apiResponse{StatusCode: 500}
		ch <- chResp
		return chResp
	}

	fmt.Println("Order placed successfully")
	chResp := apiResponse{StatusCode: 200, Message: "Order Placed successfully"}
	ch <- chResp
	return chResp
}
