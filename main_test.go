package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type postResponse struct {
	Id string `json:"id"`
}

type getResponse struct {
	Points int `json:"points"`
}

type receiptPoints struct {
	receipt string
	points  int
}

var router *gin.Engine

func TestMain(m *testing.M) {
	router = SetupAPI()

	// run all tests and exit with code
	code := m.Run()
	os.Exit(code)
}

func TestPostValidReceipts(t *testing.T) {
	targetReceipt := `{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    },
    {
      "shortDescription": "Emils Cheese Pizza",
      "price": "12.25"
    },
    {
      "shortDescription": "Knorr Creamy Chicken",
      "price": "1.26"
    },
    {
      "shortDescription": "Doritos Nacho Cheese",
      "price": "3.35"
    },
    {
      "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
      "price": "12.00"
    }
  ],
  "total": "35.35"
}`
	simpleReceipt := `{
    "retailer": "Target",
    "purchaseDate": "2022-01-02",
    "purchaseTime": "13:13",
    "total": "1.25",
    "items": [
        {"shortDescription": "Pepsi - 12-oz", "price": "1.25"}
    ]
}`
	morningReceipt := `{
    "retailer": "Walgreens",
    "purchaseDate": "2022-01-02",
    "purchaseTime": "08:13",
    "total": "2.65",
    "items": [
        {"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
        {"shortDescription": "Dasani", "price": "1.40"}
    ]
}`
	mmCornerMarketReceipt := `{
  "retailer": "M&M Corner Market",
  "purchaseDate": "2022-03-20",
  "purchaseTime": "14:33",
  "items": [
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    },
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    },
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    },
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    }
  ],
  "total": "9.00"
}`
	payloads := []string{targetReceipt, simpleReceipt, morningReceipt, mmCornerMarketReceipt}

	// Perform each test with a clean slate
	for _, payload := range payloads {
		// Test request
		req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")

		// New recorder
		w := httptest.NewRecorder()

		// Serve with mocked HTTP
		router.ServeHTTP(w, req)

		// ASSERT Response code OK
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response postResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)

		// No error in decoding and JSON contains proper ID key
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Id)
	}
}

func TestCorrectReceiptPoints(t *testing.T) {
	receiptAndCorrectPoints := []receiptPoints{
		{receipt: `{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    },
    {
      "shortDescription": "Emils Cheese Pizza",
      "price": "12.25"
    },
    {
      "shortDescription": "Knorr Creamy Chicken",
      "price": "1.26"
    },
    {
      "shortDescription": "Doritos Nacho Cheese",
      "price": "3.35"
    },
    {
      "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
      "price": "12.00"
    }
  ],
  "total": "35.35"
}`, points: 28},
		{receipt: `{
	"retailer": "M&M Corner Market",
	"purchaseDate": "2022-03-20",
	"purchaseTime": "14:33",
	"items": [
		{
			"shortDescription": "Gatorade",
			"price": "2.25"
		},
		{
			"shortDescription": "Gatorade",
			"price": "2.25"
		},
		{
			"shortDescription": "Gatorade",
			"price": "2.25"
		},
		{
			"shortDescription": "Gatorade",
			"price": "2.25"
		}
	],
	"total": "9.00"
}`, points: 109},
	}

	for _, receiptAndPoints := range receiptAndCorrectPoints {
		// Test request
		postReq := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(receiptAndPoints.receipt))
		postReq.Header.Set("Content-Type", "application/json")

		// New recorder
		w := httptest.NewRecorder()

		// Serve with mocked HTTP
		router.ServeHTTP(w, postReq)

		// Parse response1
		var response1 postResponse
		err := json.Unmarshal(w.Body.Bytes(), &response1)

		// No error in decoding and JSON
		assert.NoError(t, err)

		// Reset recorder
		w = httptest.NewRecorder()

		// Get receipt points by ID
		getReq := httptest.NewRequest("GET", "/receipts/"+response1.Id+"/points", nil)
		router.ServeHTTP(w, getReq)

		// Parse response2
		var response2 getResponse

		err = json.Unmarshal(w.Body.Bytes(), &response2)

		// No error in decoding
		assert.NoError(t, err)

		// Points equal whats expected
		assert.Equal(t, receiptAndPoints.points, response2.Points)
	}
}

func TestInvalidRetailerFailure(t *testing.T) {
	invalidRetailer := `{
  "retailer": "!!!@@@###",
  "purchaseDate": "2025-01-15",
  "purchaseTime": "15:30",
  "items": [],
  "total": "0.00"
}`

	// Test request
	postReq := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(invalidRetailer))
	postReq.Header.Set("Content-Type", "application/json")

	// New recorder
	w := httptest.NewRecorder()

	// Serve with mocked HTTP
	router.ServeHTTP(w, postReq)

	// Assert Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvalidPriceFormat(t *testing.T) {
	invalidPriceTotal1 := `{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    }
  ],
  "total": "35"
}`
	invalidPriceTotal2 := `{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    }
  ],
  "total": "35.349"
}`
	invalidItemPrice := `{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.4"
    }
  ],
  "total": "35.00"
}`
	testPayloads := []string{invalidPriceTotal1, invalidPriceTotal2, invalidItemPrice}

	for _, payload := range testPayloads {
		// Test request
		req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")

		// New recorder
		w := httptest.NewRecorder()

		// Serve with mocked HTTP
		router.ServeHTTP(w, req)

		// Assert Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}
}

func TestOutOfBonusPointsTime(t *testing.T) {

	/* Each test case is designed to have 1 point */
	before2Pm := `{
  "retailer": "A",
  "purchaseDate": "2025-01-14",
  "purchaseTime": "13:59",
  "items": [{ "shortDescription": "B", "price": "1.01" }],
  "total": "1.01"
}`
	exactly2Pm := `{
  "retailer": "A",
  "purchaseDate": "2025-01-14",
  "purchaseTime": "14:00",
  "items": [{ "shortDescription": "B", "price": "1.01" }],
  "total": "1.01"
}`
	exactly4Pm := `{
  "retailer": "A",
  "purchaseDate": "2025-01-14",
  "purchaseTime": "16:00",
  "items": [{ "shortDescription": "B", "price": "1.01" }],
  "total": "1.01"
}`
	after4Pm := `{
  "retailer": "A",
  "purchaseDate": "2025-01-14",
  "purchaseTime": "16:01",
  "items": [{ "shortDescription": "B", "price": "1.01" }],
  "total": "1.01"
}`
	testPayloads := []string{before2Pm, exactly2Pm, exactly4Pm, after4Pm}

	for _, payload := range testPayloads {
		// Test request
		postReq := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(payload))
		postReq.Header.Set("Content-Type", "application/json")

		// New recorder
		w := httptest.NewRecorder()

		// Serve with mocked HTTP
		router.ServeHTTP(w, postReq)

		// Parse response1
		var response1 postResponse
		err := json.Unmarshal(w.Body.Bytes(), &response1)

		// No error in decoding and JSON
		assert.NoError(t, err)

		// Reset recorder
		w = httptest.NewRecorder()

		// Get receipt points by ID
		getReq := httptest.NewRequest("GET", "/receipts/"+response1.Id+"/points", nil)
		router.ServeHTTP(w, getReq)

		// Parse response2
		var response2 getResponse

		err = json.Unmarshal(w.Body.Bytes(), &response2)

		// No error in decoding
		assert.NoError(t, err)

		// Points equal whats expected
		assert.Equal(t, 1, response2.Points)
	}
}

func TestInBonusPointsTime(t *testing.T) {
	/* Each test case was designed to produce exactly 11 points */
	rightAfter2Pm := `{
  "retailer": "A",
  "purchaseDate": "2025-01-14",
  "purchaseTime": "14:01",
  "items": [{ "shortDescription": "B", "price": "1.01" }],
  "total": "1.01"
}`
	between := `{
  "retailer": "A",
  "purchaseDate": "2025-01-14",
  "purchaseTime": "15:00",
  "items": [{ "shortDescription": "B", "price": "1.01" }],
  "total": "1.01"
}`
	rightBefore4Pm := `{
  "retailer": "A",
  "purchaseDate": "2025-01-14",
  "purchaseTime": "15:59",
  "items": [{ "shortDescription": "B", "price": "1.01" }],
  "total": "1.01"
}`
	testPayloads := []string{rightAfter2Pm, between, rightBefore4Pm}

	for _, payload := range testPayloads {
		// Test request
		postReq := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(payload))
		postReq.Header.Set("Content-Type", "application/json")

		// New recorder
		w := httptest.NewRecorder()

		// Serve with mocked HTTP
		router.ServeHTTP(w, postReq)

		// Parse response1
		var response1 postResponse
		err := json.Unmarshal(w.Body.Bytes(), &response1)

		// No error in decoding and JSON
		assert.NoError(t, err)

		// Reset recorder
		w = httptest.NewRecorder()

		// Get receipt points by ID
		getReq := httptest.NewRequest("GET", "/receipts/"+response1.Id+"/points", nil)
		router.ServeHTTP(w, getReq)

		// Parse response2
		var response2 getResponse

		err = json.Unmarshal(w.Body.Bytes(), &response2)

		// No error in decoding
		assert.NoError(t, err)

		// Points equal whats expected
		assert.Equal(t, 11, response2.Points)
	}
}

func TestInvalidDateTimeFailure(t *testing.T) {
	invalidDate := `{
  "retailer": "Invalid Date Format Test",
  "purchaseDate": "2025-1-7",
  "purchaseTime": "15:00",
  "items": [{ "shortDescription": "Test Item", "price": "0.99" }],
  "total": "0.99"
}`
	invalidTime := `{
  "retailer": "Invalid Time Format Test",
  "purchaseDate": "2025-01-15",
  "purchaseTime": "3 PM",
  "items": [{ "shortDescription": "Test Item", "price": "0.99" }],
  "total": "0.99"
}`
	testPayloads := []string{invalidDate, invalidTime}

	for _, payload := range testPayloads {
		// Test request
		req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")

		// New recorder
		w := httptest.NewRecorder()

		// Serve with mocked HTTP
		router.ServeHTTP(w, req)

		// Assert Bad Request
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}
}
