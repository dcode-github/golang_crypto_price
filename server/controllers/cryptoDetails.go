package controllers

import (
	"cryptoProject/server/cache"
	"cryptoProject/server/model"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Request struct {
	Crypto string `json:"crypto"`
	Fiat   string `json:"fiat"`
}

type Response struct {
	Fiat  string  `json:"fiat"`
	Value float64 `json:"value"`
}

var cacshe = cache.NewCache()

func Details(rw http.ResponseWriter, req *http.Request) {
	var c Request
	err := json.NewDecoder(req.Body).Decode(&c)
	if err != nil {
		panic(err)
	}
	crypto := c.Crypto
	fiat := c.Fiat

	key := model.CacheKey{
		Crypto: crypto,
		Fiat:   fiat,
	}

	if value, found := cacshe.Get(key); found {
		log.Println("Cache Hit")
		json.NewEncoder(rw).Encode(value)
		return
	}
	log.Println("Cache Miss")

	URL := "https://min-api.cryptocompare.com/data/price?fsym=" + crypto + "&tsyms=" + fiat
	resp, err := http.Get(URL)
	if err != nil {
		panic(err)
	}
	var respData map[string]float64
	defer resp.Body.Close()
	body := json.NewDecoder(resp.Body).Decode(&respData)
	log.Println(body)
	cacshe.Set(key, respData, 10*time.Second)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	error := json.NewEncoder(rw).Encode(respData)
	if error != nil {
		panic(error)
	}
}
