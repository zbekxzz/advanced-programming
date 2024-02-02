package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
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

type User struct {
	gorm.Model
	Username string
	Email    string
	Password string
}

type Recipe struct {
	gorm.Model
	Title             string
	Category          string
	RecipeText        string
	PublisherUsername string
	PublishedDate     time.Time
}

type UserRepository interface {
	GetUserByID(id uint) (*User, error)
	UpdateUserName(id uint, newName string) error
	DeleteUser(id uint) error
	CreateUser(username, email, password string) error
	GetAllUsers() ([]User, error)
}

type RecipeRepository interface {
	GetRecipeByID(id uint) (*Recipe, error)
	UpdateRecipeTitle(id uint, newTitle string) error
	DeleteRecipe(id uint) error
	CreateRecipe(title, category, recipeText, publisherUsername string, publishedDate time.Time) error
	GetAllRecipes(filter string, sort string, page int, limit int) ([]Recipe, error)
}

type DBUserRepository struct {
	DB *gorm.DB
}

type DBRecipeRepository struct {
	DB *gorm.DB
}

var log = getLogger()

func getLogger() *logrus.Logger {
	logger := logrus.New()

	// Set the formatter to JSONFormatter
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Create a file for logging
	file, err := os.OpenFile("logs.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// Set the output to the file
		logger.SetOutput(file)
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	return logger
}

var limiter = rate.NewLimiter(1, 3)

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
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	params := mux.Vars(r)
	userID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.WithError(err).Error("Invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.UserRepo.GetUserByID(uint(userID))
	if err != nil {
		log.WithError(err).Error("Error getting user by ID")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("userID", userID).Info("Retrieved user by ID")

	response := ResponseBody{
		Status:  "success",
		Message: user.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *UserHandler) UpdateUserNameHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	params := mux.Vars(r)
	userID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.WithError(err).Error("Invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewName string `json:"newName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.UpdateUserName(uint(userID), req.NewName); err != nil {
		log.WithError(err).Error("Error updating user name")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithFields(logrus.Fields{
		"userID":  userID,
		"newName": req.NewName,
	}).Info("User name updated")

	response := ResponseBody{
		Status:  "success",
		Message: "User successfully updated",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	params := mux.Vars(r)
	userID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.WithError(err).Error("Invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.DeleteUser(uint(userID)); err != nil {
		log.WithError(err).Error("Error deleting user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("userID", userID).Info("User deleted")

	response := ResponseBody{
		Status:  "success",
		Message: "User successfully deleted",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	var req RegisterRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" && req.Email == "" && req.Password == "" {
		log.Error("Invalid JSON message")
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	err := h.UserRepo.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		log.WithError(err).Error("Error creating user")
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	log.WithField("username", req.Username).Info("New user registered")

	response := ResponseBody{
		Status:  "success",
		Message: "New user successfully registered " + req.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	users, err := h.UserRepo.GetAllUsers()
	if err != nil {
		log.WithError(err).Error("Error getting all users")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("userCount", len(users)).Info("Retrieved all users")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		return
	}
}

func (r *DBRecipeRepository) GetRecipeByID(id uint) (*Recipe, error) {
	var recipe Recipe
	result := r.DB.First(&recipe, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &recipe, nil
}

func (r *DBRecipeRepository) UpdateRecipeTitle(id uint, newTitle string) error {
	var recipe Recipe
	result := r.DB.First(&recipe, id)
	if result.Error != nil {
		return result.Error
	}

	recipe.Title = newTitle
	result = r.DB.Save(&recipe)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *DBRecipeRepository) DeleteRecipe(id uint) error {
	result := r.DB.Delete(&Recipe{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *DBRecipeRepository) CreateRecipe(title, category, recipeText, publisherUsername string, publishedDate time.Time) error {
	recipe := Recipe{Title: title, Category: category, RecipeText: recipeText, PublisherUsername: publisherUsername, PublishedDate: publishedDate}
	result := r.DB.Create(&recipe)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *DBRecipeRepository) GetAllRecipes(filter string, sort string, page int, limit int) ([]Recipe, error) {
	var recipes []Recipe
	offset := (page - 1) * limit

	query := r.DB.Model(&Recipe{})
	if filter != "" {
		query = query.Where("category LIKE ?", "%"+filter+"%")
	}
	if sort != "" {
		query = query.Order(sort)
	}
	query = query.Offset(offset).Limit(limit).Find(&recipes)

	if query.Error != nil {
		return nil, query.Error
	}

	return recipes, nil
}

type RecipeHandler struct {
	RecipeRepo RecipeRepository
}

func (h *RecipeHandler) GetRecipeByIDHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	params := mux.Vars(r)
	recipeID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.WithError(err).Error("Invalid recipe ID")
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	recipe, err := h.RecipeRepo.GetRecipeByID(uint(recipeID))
	if err != nil {
		log.WithError(err).Error("Error getting recipe by ID")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("recipeID", recipeID).Info("Retrieved recipe by ID")

	response := ResponseBody{
		Status:  "success",
		Message: recipe.Title,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *RecipeHandler) UpdateRecipeTitleHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	params := mux.Vars(r)
	recipeID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.WithError(err).Error("Invalid recipe ID")
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewTitle string `json:"newTitle"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.RecipeRepo.UpdateRecipeTitle(uint(recipeID), req.NewTitle); err != nil {
		log.WithError(err).Error("Error updating recipe title")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithFields(logrus.Fields{
		"recipeID": recipeID,
		"newTitle": req.NewTitle,
	}).Info("Recipe title updated")

	response := ResponseBody{
		Status:  "success",
		Message: "Recipe title successfully updated",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *RecipeHandler) DeleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	params := mux.Vars(r)
	recipeID, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.WithError(err).Error("Invalid recipe ID")
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	if err := h.RecipeRepo.DeleteRecipe(uint(recipeID)); err != nil {
		log.WithError(err).Error("Error deleting recipe")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("recipeID", recipeID).Info("Recipe deleted")

	response := ResponseBody{
		Status:  "success",
		Message: "Recipe successfully deleted",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *RecipeHandler) CreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	var req struct {
		Title             string    `json:"title"`
		Category          string    `json:"category"`
		RecipeText        string    `json:"recipeText"`
		PublisherUsername string    `json:"publisherUsername"`
		PublishedDate     time.Time `json:"publishedDate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Category == "" || req.RecipeText == "" || req.PublisherUsername == "" {
		log.Error("Missing required fields")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.RecipeRepo.CreateRecipe(req.Title, req.Category, req.RecipeText, req.PublisherUsername, req.PublishedDate)
	if err != nil {
		log.WithError(err).Error("Error creating recipe")
		http.Error(w, "Error creating recipe", http.StatusInternalServerError)
		return
	}

	log.WithField("recipeTitle", req.Title).Info("Recipe created")

	response := ResponseBody{
		Status:  "success",
		Message: "Recipe successfully created",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *RecipeHandler) GetAllRecipesHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	filter := r.URL.Query().Get("filter")
	sort := r.URL.Query().Get("sort")
	pageStr := r.URL.Query().Get("page")

	// Set default values for pagination
	limit := 12
	page := 1

	// Convert pageStr to integer
	if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
		page = pageInt
	}

	// Call GetAllRecipes with filter, sort, page, and limit
	recipes, err := h.RecipeRepo.GetAllRecipes(filter, sort, page, limit)
	if err != nil {
		log.WithError(err).Error("Error getting all recipes")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("recipeCount", len(recipes)).Info("Retrieved all recipes")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(recipes)
	if err != nil {
		return
	}
}

func (h *UserHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	if r.Method != http.MethodPost {
		log.Error("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody LoginRequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		log.WithError(err).Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if requestBody.Username == "" && requestBody.Password == "" {
		log.Error("Invalid JSON message")
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	users, err := h.UserRepo.GetAllUsers()
	if err != nil {
		log.WithError(err).Error("Error finding user")
	}

	var responseMessage string
	for _, user := range users {
		if user.Username == requestBody.Username && user.Password == requestBody.Password {
			responseMessage = "You successfully logged in " + requestBody.Username
			break
		}
	}

	log.WithField("username", requestBody.Username).Info("User logged in")

	response := ResponseBody{
		Status:  "success",
		Message: responseMessage,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/errors/404.html")
}

func main() {
	dsn := "user=postgres password=Zbekxzz3 dbname=advanced sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.WithError(err).Error("Failed to perform database auto migration")
		return
	}

	userRepo := &DBUserRepository{
		DB: db,
	}

	userHandler := &UserHandler{
		UserRepo: userRepo,
	}

	recipeRepo := &DBRecipeRepository{
		DB: db,
	}

	recipeHandler := &RecipeHandler{
		RecipeRepo: recipeRepo,
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

	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.GetRecipeByIDHandler).Methods("GET")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.UpdateRecipeTitleHandler).Methods("PUT")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.DeleteRecipeHandler).Methods("DELETE")
	router.HandleFunc("/api/recipes", recipeHandler.CreateRecipeHandler).Methods("POST")
	router.HandleFunc("/api/recipes", recipeHandler.GetAllRecipesHandler).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	addr := ":6060"
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Server error: ", err)
	}
}
