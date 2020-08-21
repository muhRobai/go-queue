package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
)

func (c *initAPI) CreateQueueHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var p QueueRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		log.Println(err)
		http.Error(w, "faild-convert-json", http.StatusInternalServerError)
		return
	}

	resp, err := c.CreateQueue(ctx, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "faild-convert-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *initAPI) DeleteQueueHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var p MessageRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		log.Println(err)
		http.Error(w, "faild-convert-json", http.StatusInternalServerError)
		return
	}

	resp, err := c.DeleteQueue(ctx, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "faild-convert-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *initAPI) CallQueueHanler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var p MessageRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		log.Println(err)
		http.Error(w, "faild-convert-json", http.StatusInternalServerError)
		return
	}

	resp, err := c.CallMessage(ctx, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "faild-convert-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *initAPI) initDB() {

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(dbPort)
	if err != nil {
		log.Println(err.Error())
		return
	}

	dbConfig := &pgx.ConnConfig{
		Port:     uint16(port),
		Host:     dbHost,
		User:     dbUser,
		Password: dbPass,
		Database: dbName,
	}

	connection := pgx.ConnPoolConfig{
		ConnConfig:     *dbConfig,
		MaxConnections: 5,
	}

	c.Db, err = pgx.NewConnPool(connection)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func StartHTTP() http.Handler {
	api, err := CreateAPI()
	if err != nil {
		log.Println(err)
		return nil
	}

	api.initDB()

	r := mux.NewRouter()
	r.HandleFunc("/api/create-queue", api.CreateQueueHandler).Methods("POST")
	r.HandleFunc("/api/delete-queue", api.DeleteQueueHandler).Methods("POST")
	r.HandleFunc("/api/call-queue", api.CallQueueHanler).Methods("POST")
	//get customer list
	// r.HandleFunc("/api/customer/list", api.GetCustomerHanler).Methods("GET")
	return r
}
