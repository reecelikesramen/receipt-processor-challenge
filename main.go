package main

import (
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const DefaultPort = "8080"

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurcahseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

var retailerRegex *regexp.Regexp
var priceRegex *regexp.Regexp
var itemShortDescRegex *regexp.Regexp

var inMemoryStore sync.Map

// TODO switch from indentedJSON to JSON after development

func main() {
	router := SetupAPI()

	// port from env or default
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	router.Run(":" + port)
}

func SetupAPI() *gin.Engine {
	// Compile RegExs once
	retailerRegex = regexp.MustCompile(`^[\w\s\-&]+$`)
	priceRegex = regexp.MustCompile(`^\d+\.\d{2}$`)
	itemShortDescRegex = regexp.MustCompile(`^[\w\s\-]+$`)

	router := gin.Default()
	router.POST("/receipts/process", processReceipt)
	router.GET("/receipts/:id/points", getReceiptPoints)

	return router
}

func processReceipt(c *gin.Context) {
	var newReceipt Receipt

	// Payload should bind to receipt type, otherwise bad request with custom message
	if err := c.ShouldBindJSON(&newReceipt); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Doesn't bind"})
		return
	}

	// Validate Retailer field against RegEx in schema or Bad Request
	if !retailerRegex.MatchString(newReceipt.Retailer) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Invalid retailer name"})
		return
	}

	// Total receipt points accumulator
	var points int = 0

	// 1 point per alphanumeric char in retailer name
	for _, char := range newReceipt.Retailer {
		if char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			points += 1
		}
	}

	// Validate Total against RegEx in schema or Bad Request
	if !priceRegex.MatchString(newReceipt.Total) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Invalid total format"})
		return
	}

	// Parse receipt total or Bad Request
	receiptTotal, err := strconv.ParseFloat(newReceipt.Total, 64)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Receipt total not a float"})
	}

	// 50 points if total is a round dollar amount, 25 points if total is a 25 cent amount
	if getChange(receiptTotal) == 0 {
		// 50 points + 25 points because round dollar around and is a multiple of 0.25
		points += 75
	} else if getChange(receiptTotal) == 25 {
		points += 25
	}

	// 5 points for every two items
	points += (len(newReceipt.Items) / 2) * 5

	// If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.
	for _, item := range newReceipt.Items {

		// Validate Short Description against RegEx in schema or Bad Request
		if !itemShortDescRegex.MatchString(item.ShortDescription) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Invalid short description format"})
			return
		}

		// Validate Item Price against RegEx in schema or Bad Request
		if !priceRegex.MatchString(item.Price) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Item price invalid format"})
			return
		}

		// Reduce nesting, continue if short description is not a multiple of 3
		if len(strings.TrimSpace(item.ShortDescription))%3 != 0 {
			continue
		}

		// Parse item price or Bad Request
		itemPrice, err := strconv.ParseFloat(item.Price, 64)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Item price not a float"})
		}

		// Round up item price * 0.2, add to points
		roundUp := math.Ceil(itemPrice * 0.2)
		points += int(roundUp)
	}

	timeString := newReceipt.PurcahseDate + "T" + newReceipt.PurchaseTime
	dateTime, err := time.Parse("2006-01-02T15:04", timeString)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"description": "The receipt is invalid. Date is wrong"})
		return
	}

	// 6 points if the day in the purchase date is odd.
	if dateTime.Day()%2 == 1 {
		points += 6
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	// [2:01pm, 3:59pm], description as after 2pm & before 4pm, interprating as exclusive range
	if dateTime.Hour() >= 14 && dateTime.Hour() < 16 && (dateTime.Hour() > 14 || dateTime.Minute() > 0) {
		points += 10
	}

	receiptGuid := uuid.New().String()

	inMemoryStore.Store(receiptGuid, points)

	c.IndentedJSON(http.StatusOK, gin.H{"id": receiptGuid})
}

func getReceiptPoints(c *gin.Context) {
	// don't need to check input against regex since the in memory store is populated by GUIDs and will always be valid
	points, ok := inMemoryStore.Load(c.Param("id"))

	// exit if we can't find this receipt ID
	if !ok {
		c.IndentedJSON(http.StatusNotFound, gin.H{"description": "No receipt found for that ID."})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"points": points})
}

// For a dollar amounted represented as a float, returns the change in cents as an integer
func getChange(dollars float64) int {
	return int((dollars * 100)) % 100
}
