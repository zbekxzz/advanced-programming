package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

type RegisterRequestBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ResponseUser struct {
	Status  string `json:"status"`
	Message []User `json:"message"`
}

type User struct {
	gorm.Model
	Username string
	Email    string
	Password string
}

type UserRepository interface {
	GetUserByID(id uint) (*User, error)
	UpdateUserName(id uint, newName string) error
	DeleteUser(id uint) error
	CreateUser(username, email, password string) error
	GetAllUsers() ([]User, error)
}

type DBUserRepository struct {
	DB *gorm.DB
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (r *DBUserRepository) GetUserByID(id uint) (*User, error) {
	var user User
	result := r.DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *DBUserRepository) UpdateUserName(id uint, newName string) error {
	var user User
	result := r.DB.First(&user, id)
	if result.Error != nil {
		return result.Error
	}

	user.Username = newName
	result = r.DB.Save(&user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *DBUserRepository) DeleteUser(id uint) error {
	result := r.DB.Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *DBUserRepository) CreateUser(username, email, password string) error {
	user := User{Username: username, Email: email, Password: password}
	result := r.DB.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *DBUserRepository) GetAllUsers() ([]User, error) {
	var users []User
	result := r.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

type UserHandler struct {
	UserRepo UserRepository
}

func (h *UserHandler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.UserRepo.GetUserByID(uint(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ResponseBody{
		Status:  "success",
		Message: user.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) UpdateUserNameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewName string `json:"newName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.UpdateUserName(uint(userID), req.NewName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ResponseBody{
		Status:  "success",
		Message: "User successfully updated",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.DeleteUser(uint(userID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ResponseBody{
		Status:  "success",
		Message: "User successfully deleted",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" && req.Email == "" && req.Password == "" {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	err := h.UserRepo.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		log.Println("Error creating user:", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received new user: %s\n", req.Username)

	response := ResponseBody{
		Status:  "success",
		Message: "New user successfully registered " + req.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserRepo.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody LoginRequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if requestBody.Username == "" && requestBody.Password == "" {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	responseMessage := "Invalid login or password"

	users, err := h.UserRepo.GetAllUsers()
	for _, user := range users {
		if user.Username == requestBody.Username && user.Password == requestBody.Password {
			responseMessage = "You successfully logged in " + requestBody.Username
			break
		}
	}

	if err != nil {
		log.Println("Error finding user:", err)
	}

	fmt.Printf("User %s successfully logged in\n", requestBody.Username)

	response := ResponseBody{
		Status:  "success",
		Message: responseMessage,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/errors/404.html")
}

func main() {
	dsn := "user=postgres password=Zbekxzz3 dbname=advanced sslmode=disable"
	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&User{})

	userRepo := &DBUserRepository{
		DB: db,
	}

	userHandler := &UserHandler{
		UserRepo: userRepo,
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/index.html")
	})

	router.HandleFunc("/help", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/help.html")
	})

	router.HandleFunc("/recipes", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/recipes.html")
	})

	router.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/errors/404.html")
	})

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/auth/login.html")
	})

	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/auth/register.html")
	})

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/crud-test/users.html")
	})

	router.HandleFunc("/api/register", userHandler.CreateUserHandler).Methods("POST")
	router.HandleFunc("/api/login", userHandler.LoginUserHandler).Methods("POST")

	router.HandleFunc("/api/users/{id:[0-9]+}", userHandler.GetUserByIDHandler).Methods("GET")
	router.HandleFunc("/api/users/{id:[0-9]+}", userHandler.UpdateUserNameHandler).Methods("PUT")
	router.HandleFunc("/api/users/{id:[0-9]+}", userHandler.DeleteUserHandler).Methods("DELETE")
	router.HandleFunc("/api/users", userHandler.GetAllUsersHandler).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	fmt.Println("Server listening on http://localhost:8080")
	http.ListenAndServe(":8080", router)
}
