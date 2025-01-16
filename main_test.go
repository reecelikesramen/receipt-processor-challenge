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

		println(w.Body.String())

		err = json.Unmarshal(w.Body.Bytes(), &response2)

		// No error in decoding
		assert.NoError(t, err)

		// Points equal whats expected
		assert.Equal(t, receiptAndPoints.points, response2.Points)
	}

}
