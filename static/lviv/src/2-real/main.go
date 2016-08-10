package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	port      = flag.Int("port", 4200, "Application port")
	dbconnect = flag.String("db", "", "Database connection string")
)

var store *sql.DB

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func main() {
	flag.Parse()

	err := initDB(*dbconnect)
	if err != nil {
		log.Fatalf("Can't connect to database: %s", err)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	http.HandleFunc("/healthcheck", healthHandler)
	http.HandleFunc("/", productsHandler)
	log.Printf("Listening on %s...", addr)
	err = http.ListenAndServe(addr, nil)
	log.Fatal(err)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := store.Ping()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(products)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func initDB(dbconnect string) error {
	var err error
	store, err = sql.Open("postgres", dbconnect)
	if err != nil {
		return err
	}
	err = store.Ping()
	if err != nil {
		return err
	}
	return nil
}

func getProducts() ([]Product, error) {
	var products []Product
	query := `SELECT id, name, price FROM products`
	rows, err := store.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}
