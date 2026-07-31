// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Order struct {
	ID       int     `json:"id"`
	Customer string  `json:"customer"`
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"status"`
}

type orderInput struct {
	Customer string  `json:"customer"`
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"status"`
}

type orderStore struct {
	mu     sync.RWMutex
	nextID int
	orders map[int]Order
}

func newOrderStore() *orderStore {
	return &orderStore{
		nextID: 1,
		orders: make(map[int]Order),
	}
}

func (s *orderStore) create(input orderInput) Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	order := Order{
		ID:       s.nextID,
		Customer: input.Customer,
		Product:  input.Product,
		Quantity: input.Quantity,
		Amount:   input.Amount,
		Status:   input.Status,
	}
	s.orders[order.ID] = order
	s.nextID++
	return order
}

func (s *orderStore) list() []Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].ID < orders[j].ID
	})
	return orders
}

func (s *orderStore) get(id int) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[id]
	return order, ok
}

func (s *orderStore) update(id int, input orderInput) (Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orders[id]; !ok {
		return Order{}, false
	}
	order := Order{
		ID:       id,
		Customer: input.Customer,
		Product:  input.Product,
		Quantity: input.Quantity,
		Amount:   input.Amount,
		Status:   input.Status,
	}
	s.orders[id] = order
	return order, true
}

func (s *orderStore) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orders[id]; !ok {
		return false
	}
	delete(s.orders, id)
	return true
}

type orderAPI struct {
	store *orderStore
}

func NewServer() http.Handler {
	api := &orderAPI{store: newOrderStore()}
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", api.handleOrders)
	mux.HandleFunc("/orders/", api.handleOrder)
	return mux
}

func (api *orderAPI) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, api.store.list())
	case http.MethodPost:
		input, err := readOrderInput(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if input.Status == "" {
			input.Status = "pending"
		}
		writeJSON(w, http.StatusCreated, api.store.create(input))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (api *orderAPI) handleOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseOrderID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		order, ok := api.store.get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeJSON(w, http.StatusOK, order)
	case http.MethodPut:
		input, err := readOrderInput(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if input.Status == "" {
			input.Status = "pending"
		}
		order, ok := api.store.update(id, input)
		if !ok {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeJSON(w, http.StatusOK, order)
	case http.MethodDelete:
		if !api.store.delete(id) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func parseOrderID(path string) (int, error) {
	value := strings.TrimPrefix(path, "/orders/")
	if value == "" || strings.Contains(value, "/") {
		return 0, errors.New("invalid order id")
	}
	id, err := strconv.Atoi(value)
	if err != nil || id < 1 {
		return 0, errors.New("invalid order id")
	}
	return id, nil
}

func readOrderInput(r *http.Request) (orderInput, error) {
	defer r.Body.Close()

	var input orderInput
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return orderInput{}, errors.New("invalid JSON body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return orderInput{}, errors.New("request body must contain one JSON object")
	}

	input.Customer = strings.TrimSpace(input.Customer)
	input.Product = strings.TrimSpace(input.Product)
	input.Status = strings.TrimSpace(input.Status)
	if input.Customer == "" || input.Product == "" {
		return orderInput{}, errors.New("customer and product are required")
	}
	if input.Quantity < 1 {
		return orderInput{}, errors.New("quantity must be greater than zero")
	}
	if input.Amount < 0 {
		return orderInput{}, errors.New("amount must not be negative")
	}
	return input, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func main() {
	const address = ":8080"
	fmt.Printf("order demo listening on http://localhost%s\n", address)
	if err := http.ListenAndServe(address, NewServer()); err != nil {
		log.Fatal(err)
	}
}
