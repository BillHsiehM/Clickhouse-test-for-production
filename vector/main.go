package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bradleypeabody/gouuidv6"
)

type requestBody struct {
	Message string `json:"message"`
	Col1    string `json:"col1,omitempty"`
	Col2    string `json:"col2,omitempty"`
	Col3    string `json:"col3,omitempty"`
	Col4    string `json:"col4,omitempty"`
}

func main() {
	if len(os.Args) != 2 {
		panic(errors.New("invalid number of arguments"))
	}
	// Number of concurrent workers
	concurrency, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(errors.New("invalid concurrency value"))
	}

	// Rate limit: 100 requests per second
	rateLimit := time.Second / time.Duration(200)

	// Create a semaphore to control concurrency
	semaphore := make(chan struct{}, concurrency)

	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Create a ticker for rate limiting
	ticker := time.NewTicker(rateLimit)
	defer ticker.Stop()

	for {
		// Wait for the next tick
		<-ticker.C

		// Acquire a semaphore token
		semaphore <- struct{}{}

		// Increment the wait group counter
		wg.Add(1)

		// Run the task concurrently
		go func() {
			defer wg.Done()

			if err := postUUID(); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				//fmt.Println(" Successfully sent: ", time.Now().Format(time.RFC3339))
			}

			// Release the semaphore token
			<-semaphore
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func postUUID() error {
	// Generate a random UUID string
	randomUUID := gouuidv6.New().String()
	body := requestBody{
		Message: randomUUID,
		Col1:    RandStringBytes(6),
		Col2:    RandStringBytes(10),
		Col3:    RandStringBytes(3),
		Col4:    RandStringBytes(3),
	}
	//fmt.Print(randomUUID)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON body: %v", err)
	}

	// Make the POST request
	resp, err := http.Post("http://127.0.0.1:8888", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("unable to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	return nil
}
