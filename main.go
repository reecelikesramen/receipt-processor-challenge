package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const DefaultPort = "8080"

type item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type receipt struct {
	Retailer     string `json:"retailer"`
	PurcahseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []item `json:"items"`
	Total        string `json:"total"`
}

func main() {
	router := gin.Default()
	router.POST("/receipts/process", postReceipt)

	// port from env or default
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	router.Run(":" + port)
}

func postReceipt(c *gin.Context) {
	c.IndentedJSON(http.StatusCreated, "hello")
}
