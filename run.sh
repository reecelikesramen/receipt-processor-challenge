#!/bin/bash

RECEIPT_FILE=${1:-"examples/simple-receipt.json"}

# Check if the Docker image exists
if ! docker image inspect receipt-processor > /dev/null 2>&1; then
    echo "Docker image not found. Building the image..."
    docker build -t receipt-processor .
else
    echo "Docker image already exists."
fi

# Check if a container from the image is already running
if ! docker ps | grep -q "receipt-processor"; then
    echo "Starting the Docker container..."
    docker run -p 8080:8080 -d receipt-processor
else
    echo "Docker container is already running."
fi

json=$(curl -s -X POST http://localhost:8080/receipts/process -H "Content-Type: application/json" -d @$RECEIPT_FILE)

id=$(echo "$json" | sed -n 's/.*"id": *"\([^"]*\)".*/\1/p')

json=$(curl -s -X GET http://localhost:8080/receipts/$id/points)

points=$(echo "$json" | sed -n 's/.*"points": *\([0-9]*\).*/\1/p')

echo "Receipt points: $points"