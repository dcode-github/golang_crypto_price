package controllers

import (
	"cryptoProject/server/cache"
	"cryptoProject/server/model"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Request struct {
	Crypto string `json:"crypto"`
	Fiat   string `json:"fiat"`
}

type Response struct {
	Fiat  string  `json:"fiat"`
	Value float64 `json:"value"`
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

var (
	mu      sync.Mutex
	clients = make(map[string]*client)
	cacshe  = cache.NewCache()
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func Details(rw http.ResponseWriter, req *http.Request) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		http.Error(rw, "Unable to get IP address", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, found := clients[ip]; !found {
		clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
	}
	clients[ip].lastSeen = time.Now()

	if !clients[ip].limiter.Allow() {
		message := Message{
			Status: "Request Failed",
			Body:   "API at capacity, try again later",
		}
		rw.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(rw).Encode(&message)
		return
	}

	var reqData Request
	if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
		http.Error(rw, "Invalid request body", http.StatusBadRequest)
		return
	}

	crypto := reqData.Crypto
	fiat := reqData.Fiat

	key := model.CacheKey{
		Crypto: crypto,
		Fiat:   fiat,
	}

	if value, found := cacshe.Get(key); found {
		log.Println("Cache Hit")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(value)
		return
	}
	log.Println("Cache Miss")

	URL := "https://min-api.cryptocompare.com/data/price?fsym=" + crypto + "&tsyms=" + fiat
	resp, err := http.Get(URL)
	if err != nil {
		http.Error(rw, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(rw, "Failed to fetch data", resp.StatusCode)
		return
	}

	var respData map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		http.Error(rw, "Error decoding response", http.StatusInternalServerError)
		return
	}

	cacshe.Set(key, respData, 10*time.Second)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(rw).Encode(respData); err != nil {
		http.Error(rw, "Error encoding response", http.StatusInternalServerError)
	}
}
