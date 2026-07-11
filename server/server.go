package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CurrencyExchange struct {
	USDBRL Result `json:"USDBRL,omitempty"`
}

type Result struct {
	Code       string         `json:"code,omitempty"`
	CodeIn     string         `json:"codein,omitempty"`
	Name       string         `json:"name,omitempty"`
	High       string         `json:"high,omitempty"`
	Low        string         `json:"low,omitempty"`
	VarBid     string         `json:"varBid,omitempty"`
	PctChange  string         `json:"pctChange,omitempty"`
	Bid        string         `json:"bid,omitempty"`
	Ask        string         `json:"ask,omitempty"`
	CreateDate CustomDateTime `json:"create_date,omitempty"`
}

type CustomDateTime time.Time

func (c *CustomDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	layout := "2006-01-02 15:04:05"
	parsedTime, err := time.Parse(layout, s)
	if err != nil {
		return err
	}
	*c = CustomDateTime(parsedTime)
	return nil
}

func (c CustomDateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(c)
	return json.Marshal(t.Format("2006-01-02T15:04:05"))
}

type ExchangeOutput struct {
	Bid string `json:"bid,omitempty"`
}

type ExchangeService struct {
	db *sql.DB
}

func NewExchangeService(db *sql.DB) *ExchangeService {
	return &ExchangeService{db: db}
}

func (s ExchangeService) SearchExchangeUSD2BRLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	exchange, err := s.searchExchangeUSD2BRL()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.saveExchangeRate(exchange); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(ExchangeOutput{Bid: exchange.USDBRL.Bid}); err != nil {
		log.Printf("Error on parsing result: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s ExchangeService) searchExchangeUSD2BRL() (*CurrencyExchange, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error on search exchange on external api: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var currencyExchange CurrencyExchange
	if err = json.Unmarshal(body, &currencyExchange); err != nil {
		return nil, err
	}

	return &currencyExchange, nil
}

func (s ExchangeService) saveExchangeRate(exchange *CurrencyExchange) error {
	sqlInsert := "insert into exchange (code,codein,name,high,low,var_bid,ptc_change,bid,ask,created_at) VALUES (?,?,?,?,?,?,?,?,?,?)"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := s.db.ExecContext(ctx, sqlInsert, exchange.USDBRL.Code, exchange.USDBRL.CodeIn, exchange.USDBRL.Name, exchange.USDBRL.High, exchange.USDBRL.Low, exchange.USDBRL.VarBid, exchange.USDBRL.PctChange, exchange.USDBRL.Bid, exchange.USDBRL.Ask, time.Time(exchange.USDBRL.CreateDate))
	if err != nil {
		log.Printf("Error on save exchange rate on database: %v", err)
		return err
	}

	return nil
}

func createConnection() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", "exchange.db")
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS exchange (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL,
		codein TEXT NOT NULL,
		name TEXT NOT NULL,
		high REAL NOT NULL,
		low REAL NOT NULL,
		var_bid REAL NOT NULL,
		ptc_change REAL NOT NULL,
		bid REAL NOT NULL,
		ask REAL NOT NULL,
		created_at DATETIME NOT NULL
	);`

	if _, err = conn.Exec(createTableSQL); err != nil {
		return nil, err
	}

	return conn, nil
}

func main() {
	conn, err := createConnection()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	exchangeService := NewExchangeService(conn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", exchangeService.SearchExchangeUSD2BRLHandler)

	if err = http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
