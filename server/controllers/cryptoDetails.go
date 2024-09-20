package controllers

import (
	"encoding/json"
	"log"
	"net/http"
)

type Request struct {
	Crypto string `json:"crypto"`
	Fiat   string `json:"fiat"`
}

type Response struct {
	Fiat  string  `json:"fiat"`
	Value float64 `json:"value"`
}

func Details(rw http.ResponseWriter, req *http.Request) {
	var c Request
	err := json.NewDecoder(req.Body).Decode(&c)
	if err != nil {
		panic(err)
	}
	log.Println(c)
	crypto := c.Crypto
	fiat := c.Fiat
	URL := "https://min-api.cryptocompare.com/data/price?fsym=" + crypto + "&tsyms=" + fiat
	resp, err := http.Get(URL)
	if err != nil {
		panic(err)
	}
	var respData map[string]float64
	defer resp.Body.Close()
	body := json.NewDecoder(resp.Body).Decode(&respData)
	log.Println(body)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	error := json.NewEncoder(rw).Encode(respData)
	if error != nil {
		panic(error)
	}
}
