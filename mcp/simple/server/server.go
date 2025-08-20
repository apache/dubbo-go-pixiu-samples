/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

import (
	"github.com/gorilla/mux"
)

// User represents a user structure
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age,omitempty"`
	Profile  string `json:"profile,omitempty"`
	CreateAt string `json:"created_at"`
}

// Post represents a post structure
type Post struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Status   string `json:"status"`
	CreateAt string `json:"created_at"`
}

// SearchResult represents search result structure
type SearchResult struct {
	Users      []User `json:"users"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
}

// CreateUserRequest represents create user request structure
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age,omitempty"`
}

// Mock data
var users = []User{
	{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Age: 28, Profile: "Software Engineer at TechCorp", CreateAt: "2023-01-15T10:30:00Z"},
	{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Age: 32, Profile: "Product Manager at StartupXYZ", CreateAt: "2023-02-20T14:45:00Z"},
	{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com", Age: 25, Profile: "Designer at CreativeStudio", CreateAt: "2023-03-10T09:15:00Z"},
	{ID: 4, Name: "Diana Prince", Email: "diana@example.com", Age: 30, Profile: "Data Scientist at DataCorp", CreateAt: "2023-04-05T16:20:00Z"},
	{ID: 5, Name: "Eve Wilson", Email: "eve@example.com", Age: 27, Profile: "DevOps Engineer at CloudTech", CreateAt: "2023-05-12T11:10:00Z"},
}

var posts = []Post{
	{ID: 1, UserID: 1, Title: "Getting Started with Go", Content: "Go is a great language for backend development...", Status: "published", CreateAt: "2023-06-01T10:00:00Z"},
	{ID: 2, UserID: 1, Title: "Microservices Architecture", Content: "Building scalable microservices...", Status: "published", CreateAt: "2023-06-15T14:30:00Z"},
	{ID: 3, UserID: 2, Title: "Product Management 101", Content: "Essential skills for product managers...", Status: "published", CreateAt: "2023-06-20T09:45:00Z"},
	{ID: 4, UserID: 3, Title: "UI/UX Design Trends", Content: "Latest trends in user interface design...", Status: "draft", CreateAt: "2023-06-25T16:15:00Z"},
	{ID: 5, UserID: 4, Title: "Data Analysis with Python", Content: "Analyzing data using pandas and numpy...", Status: "published", CreateAt: "2023-07-01T12:00:00Z"},
}

var nextUserID = 6

func main() {
	r := mux.NewRouter()

	// Add CORS middleware
	r.Use(corsMiddleware)
	r.Use(loggingMiddleware)

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// User related routes
	api.HandleFunc("/users/{id:[0-9]+}", getUserHandler).Methods("GET")
	api.HandleFunc("/users/search", searchUsersHandler).Methods("GET")
	api.HandleFunc("/users", createUserHandler).Methods("POST")
	api.HandleFunc("/users/{id:[0-9]+}/posts", getUserPostsHandler).Methods("GET")

	// Health check
	api.HandleFunc("/health", healthHandler).Methods("GET")

	// Root path
	r.HandleFunc("/", rootHandler).Methods("GET")

	fmt.Println("🚀 Mock Backend Server starting on :8081")
	fmt.Println("📚 Available endpoints:")
	fmt.Println("  GET  /api/users/{id}        - Get user by ID")
	fmt.Println("  GET  /api/users/search      - Search users")
	fmt.Println("  POST /api/users             - Create user")
	fmt.Println("  GET  /api/users/{id}/posts  - Get user posts")
	fmt.Println("  GET  /api/health            - Health check")
	fmt.Println("  GET  /                      - Root endpoint")

	log.Fatal(http.ListenAndServe(":8081", r))
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// Get user handler
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Find user
	var user *User
	for _, u := range users {
		if u.ID == userID {
			user = &u
			break
		}
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if include profile details
	includeProfile := r.URL.Query().Get("include_profile")
	if includeProfile != "true" {
		user.Profile = ""
	}

	fmt.Println("user :", user)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Search users handler
func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Search users
	var filteredUsers []User
	query = strings.ToLower(query)
	for _, user := range users {
		if strings.Contains(strings.ToLower(user.Name), query) ||
			strings.Contains(strings.ToLower(user.Email), query) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	// Pagination
	total := len(filteredUsers)
	totalPages := (total + limit - 1) / limit
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		filteredUsers = []User{}
	} else {
		if end > total {
			end = total
		}
		filteredUsers = filteredUsers[start:end]
	}

	result := SearchResult{
		Users:      filteredUsers,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Create user handler
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	// Check if email already exists
	for _, user := range users {
		if user.Email == req.Email {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
	}

	// Create new user
	newUser := User{
		ID:       nextUserID,
		Name:     req.Name,
		Email:    req.Email,
		Age:      req.Age,
		CreateAt: time.Now().Format(time.RFC3339),
	}
	nextUserID++

	users = append(users, newUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

// Get user posts handler
func getUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "published"
	}

	// Find user posts
	var userPosts []Post
	for _, post := range posts {
		if post.UserID == userID && (status == "all" || post.Status == status) {
			userPosts = append(userPosts, post)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"status":  status,
		"posts":   userPosts,
		"count":   len(userPosts),
	})
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"users":     len(users),
		"posts":     len(posts),
	})
}

// Root path handler
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Mock Backend Server",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"users":       "/api/users/{id}",
			"search":      "/api/users/search?q={query}",
			"create_user": "/api/users",
			"user_posts":  "/api/users/{id}/posts",
			"health":      "/api/health",
		},
	})
}
