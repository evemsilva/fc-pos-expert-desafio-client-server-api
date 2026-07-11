package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CurrencyExchange struct {
	Bid string `json:"bid,omitempty"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error on search exchange on server: %v", err)
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var currencyExchange CurrencyExchange
	if err = json.Unmarshal(body, &currencyExchange); err != nil {
		panic(err)
	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(fmt.Sprintf("Dólar: %s", currencyExchange.Bid))
	if err != nil {
		panic(err)
	}

	if err = f.Close(); err != nil {
		panic(err)
	}
}
