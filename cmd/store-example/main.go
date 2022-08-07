package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	suffix := rand.Intn(1000)
	payload := fmt.Sprintf(`{"name": "testing demo %04d"}`, suffix)
	body := strings.NewReader(payload)
	resp, _ := http.Post("http://localhost:8080/v1/tasks/", "application/json", body)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))
}
