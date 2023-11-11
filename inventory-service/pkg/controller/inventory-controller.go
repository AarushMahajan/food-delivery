package controller

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"

	configDb "github.com/AarushMahajan/food-delivery/inventory-service/pkg/config"
)

type item struct {
	Id            int    `json:"id"`
	Item_Name     string `json:"itemName"`
	Item_quantity int    `json:"itemQuantity"`
	Is_reserved   bool   `json:"isReserved"`
	Is_booked     bool   `json:"isBooked"`
}

type reserveRequestBody struct {
	Id       int `json:"id"`
	Quantity int `json:"quantity"`
}

type DBResult struct {
	Success bool
	Error   error
}

func GetAllInventory(c *gin.Context) {

	db := configDb.GetDB()
	var storeJson []item
	rows, err := db.Query("SELECT id, item_name, Item_quantity, is_reserved, is_booked FROM store")
	if err != nil {
		fmt.Println("Error querying store")
		c.AbortWithStatus(500)
		return
	}

	for rows.Next() {
		var s item
		err := rows.Scan(&s.Id, &s.Item_Name, &s.Item_quantity, &s.Is_reserved, &s.Is_booked)
		if err != nil {
			fmt.Println("Error while scane")
			c.AbortWithStatus(500)
			return
		}
		storeJson = append(storeJson, s)
	}

	c.JSON(200, storeJson)
}

func fetchItemsAsync(trx *sql.Tx, i reserveRequestBody, wg *sync.WaitGroup, ch chan item, isReserved bool) {
	defer wg.Done()
	var items1 item
	var query string
	db := configDb.GetDB()
	if isReserved {
		// is_reserved = true and
		query = "Select id, item_name, item_quantity, is_reserved from store where  item_quantity >= ? and id = ? FOR UPDATE"
	} else {
		query = "Select id, item_name, item_quantity, is_reserved from store where is_reserved = false and item_quantity >= ? and id = ? FOR UPDATE"
	}
	row := db.QueryRow(query, i.Quantity, i.Id)

	err := row.Scan(&items1.Id, &items1.Item_Name, &items1.Item_quantity, &items1.Is_reserved)
	if err != nil {
		fmt.Println("error while selecting", err)
		trx.Rollback()
		ch <- items1
	}
	ch <- items1
}

func ReserveItem(c *gin.Context) {

	var requestBody []reserveRequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		fmt.Println("Error while converting request body to json")
		c.AbortWithStatus(500)
		return
	}

	itemIds := make([]interface{}, len(requestBody))
	params := make([]string, len(requestBody))

	for i, v := range requestBody {
		itemIds[i] = v.Id
		params[i] = "?"
	}
	// paramsStr := strings.Join(params, ", ")

	var wg sync.WaitGroup

	ch := make(chan item)

	db := configDb.GetDB()
	trx, _ := db.Begin()

	// Send Request in Parallel
	for _, v := range requestBody {
		wg.Add(1)
		go fetchItemsAsync(trx, v, &wg, ch, false)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for value := range ch {
		if value.Id == 0 {
			fmt.Println("Item not available")
			c.AbortWithStatusJSON(400, gin.H{"message": "Item not available"})
			return
		}
		fmt.Println("Item vale is:: ", value.Item_Name)
	}

	// var mtx sync.Mutex
	// mtx.Lock()
	// defer mtx.Unlock()

	// _, err = trx.Exec("Update store set is_reserved = true where id in ("+paramsStr+")", itemIds...)
	// if err != nil {
	// 	fmt.Println("Error updating store: ", err)
	// 	trx.Rollback()
	// 	// mtx.Unlock()
	// 	c.AbortWithStatus(500)
	// 	return
	// }

	// trx.Commit()
	fmt.Println("Reserved item successfully")

	c.JSON(200, gin.H{"message": "success"})
}

func bookItemDBAsync(trx *sql.Tx, item reserveRequestBody, wg *sync.WaitGroup, mtx *sync.Mutex, ch chan DBResult, totalQuantity int) {
	defer wg.Done()
	quantity := totalQuantity - item.Quantity

	// mtx.Lock()
	// defer mtx.Unlock()
	_, err := trx.Exec("Update store set is_reserved = false, item_quantity = ? where id = ?", quantity, item.Id)
	if err != nil {
		ch <- DBResult{Success: false, Error: err}
		// mtx.Unlock()
		return
	}
	ch <- DBResult{Success: true, Error: nil}
}

func BookItem(c *gin.Context) {
	fmt.Printf("Book Item fnc")
	var requestBody []reserveRequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		fmt.Println("Error while converting request body to json")
		c.AbortWithStatus(500)
		return
	}

	db := configDb.GetDB()
	trx, _ := db.Begin()

	var wg sync.WaitGroup
	ch := make(chan item)

	for _, v := range requestBody {
		wg.Add(1)
		go fetchItemsAsync(trx, v, &wg, ch, true)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	idQuantityMap := make(map[int64]int64)
	for value := range ch {
		if value.Id == 0 {
			fmt.Println("Item not available")
			c.AbortWithStatusJSON(400, gin.H{"message": "Item not available"})
			return
		}
		idQuantityMap[int64(value.Id)] = int64(value.Item_quantity)
	}

	itemUpdateCh := make(chan DBResult)
	var mtx sync.Mutex

	for _, it := range requestBody {
		wg.Add(1)
		go bookItemDBAsync(trx, it, &wg, &mtx, itemUpdateCh, int(idQuantityMap[int64(it.Id)]))
	}

	go func() {
		wg.Wait()
		close(itemUpdateCh)
	}()

	for res := range itemUpdateCh {
		if !res.Success {
			fmt.Println("error while updating item", res.Error)
			c.AbortWithStatusJSON(500, gin.H{"message": "failuer"})
			return
		}
		fmt.Println(res.Success)
	}

	trx.Commit()
	fmt.Println("Item placed successfully!!")

	c.JSON(200, gin.H{"message": "success"})
}
