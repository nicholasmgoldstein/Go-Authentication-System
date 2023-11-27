package controllers

import (
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/auth/initializers"
	"github.com/auth/models"
	"golang.org/x/crypto/bcrypt"
)

const AllLocations = "All"

func Signup(c *gin.Context) {
	var body struct {
		Email    string
		Name     string
		Password string
		Pic      string
		Intro    string
		RefRank  uint
		DOB      string
		Country  string
		Location string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	// Validate email format
	if !isValidEmail(body.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	// Parse the DOB string into a time.Time
	dob, err := time.Parse("2006-01-02", body.DOB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date of birth format",
		})
		return
	}

	//	Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// Create User
	user := models.User{
		Email:    body.Email,
		Name:     body.Name,
		Password: string(hash),
		RefRank:  0,
		DOB:      dob, // Assign the parsed date of birth
		Country:  body.Country,
		Location: body.Location,
	}

	// Create UserPermissions
	userPermissions := models.UserPermissions{
		Email:                user.Email,
		UserDeactivated:      false,
		BannedFromCommenting: false,
		BannedFromPosting:    false,
		BannedFromAnalytix:   false,
	}

	// Start a database transaction
	tx := initializers.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the user and user permissions within a transaction
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	if err := tx.Create(&userPermissions).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user permissions",
		})
		return
	}

	// Commit the transaction
	tx.Commit()

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{})
}

// isValidEmail checks if the provided email is in a valid format
func isValidEmail(email string) bool {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(emailPattern).MatchString(email)
}

func DeleteUser(c *gin.Context) {
	// Define a request body struct
	var deleteUserRequest struct {
		Email string `json:"email"`
	}

	// Bind the request body to the deleteUserRequest struct
	if err := c.BindJSON(&deleteUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Now you can access the email from deleteUserRequest.Email
	email := deleteUserRequest.Email

	// Delete the user from the database based on the email
	if err := initializers.DB.Where("email = ?", email).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	// Delete the user's permissions from the database if they exist
	var userPermissions models.UserPermissions
	result := initializers.DB.Where("email = ?", email).First(&userPermissions)
	if result.Error == nil {
		if err := initializers.DB.Delete(&userPermissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete user permissions",
			})
			return
		}
	}

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "User and permissions deleted successfully",
	})
}

func Login(c *gin.Context) {
	//	Get email and password off of req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	var user models.User
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email and/or password",
		})
		return
	}

	// Lookup user by email, including records where deleted_at is NULL
	if err := initializers.DB.Where("email = ? AND deleted_at IS NULL", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email and/or password",
		})
		return
	}

	//	Compare passwords
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email and/or password",
		})
		return
	}

	//	Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	//	Send back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("AuthZ", tokenString, 3600*24*30, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{})
}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})

}

func Logout(c *gin.Context) {
	// Clear the AuthZ cookie to log out the user
	c.SetCookie("AuthZ", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// A global map to track searched users per location
var searchedUsers = make(map[string]map[int][]models.User)

func GetUsersInLocation(c *gin.Context) {
	// Define a request body struct locally
	var getUsersInLocationRequest struct {
		Location string `json:"location"`
		Page     int    `json:"page"`
	}

	// Bind the request body to the getUsersInLocationRequest struct
	if err := c.BindJSON(&getUsersInLocationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Extract the location and page number from the request body
	location := getUsersInLocationRequest.Location
	page := getUsersInLocationRequest.Page

	// Apply default value for page if not provided
	if page < 1 {
		page = 1
	}

	pageSize := 5
	offset := (page - 1) * pageSize

	// Check if we have already searched for users in this location
	if _, ok := searchedUsers[location]; !ok {
		searchedUsers[location] = make(map[int][]models.User)
	}

	// Check if we already have results for this page in the cache
	if cachedUsers, ok := searchedUsers[location][page]; ok {
		// Return the cached results
		c.JSON(http.StatusOK, cachedUsers)
		return
	}

	var users []models.User
	if err := initializers.DB.Where("location = ?", location).Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}

	// Store the fetched users in the cache
	searchedUsers[location][page] = users

	// Return the list of users as a JSON response
	c.JSON(http.StatusOK, users)
}

// MaxCommonCharacterSeries calculates the length of the longest common character series
func MaxCommonCharacterSeries(s1, s2 string) int {
	lenS1, lenS2 := len(s1), len(s2)
	maxLen := 0

	for i := 0; i < lenS1; i++ {
		for j := 0; j < lenS2; j++ {
			k := 0
			for i+k < lenS1 && j+k < lenS2 && s1[i+k] == s2[j+k] {
				k++
			}
			if k > maxLen {
				maxLen = k
			}
		}
	}

	return maxLen
}

// SearchUsers is the controller function for searching users
func SearchUsers(c *gin.Context) {
	// SearchRequest represents the JSON request body for the search operation
	type SearchRequest struct {
		Query    string `json:"query"`
		Location string `json:"location"`
	}

	var request SearchRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Retrieve users from the database based on the location (if specified)
	var users []models.User
	if request.Location != "" {
		initializers.DB.Where("location = ?", request.Location).Find(&users)
	} else {
		initializers.DB.Find(&users)
	}

	// Perform similarity matching and sort the results in-place
	sort.Slice(users, func(i, j int) bool {
		maxSeriesLenI := MaxCommonCharacterSeries(strings.ToLower(request.Query), strings.ToLower(users[i].Name))
		maxSeriesLenJ := MaxCommonCharacterSeries(strings.ToLower(request.Query), strings.ToLower(users[j].Name))
		return maxSeriesLenI > maxSeriesLenJ
	})

	c.JSON(http.StatusOK, users)
}

func UpdateCountryAndLocation(c *gin.Context) {
	// Define a request body struct locally
	var updateCountryAndLocationRequest struct {
		Country  string `json:"country"`
		Location string `json:"location"`
		UserID   string `json:"userID"`
	}

	// Bind the request body to the updateCountryAndLocationRequest struct
	if err := c.BindJSON(&updateCountryAndLocationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Extract the user ID from the request body
	userID := updateCountryAndLocationRequest.UserID

	// Update the user's country and location in the database
	var user models.User
	if err := initializers.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	user.Country = updateCountryAndLocationRequest.Country
	user.Location = updateCountryAndLocationRequest.Location

	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user information",
		})
		return
	}

	// Respond with a success message or updated user object
	c.JSON(http.StatusOK, user)
}
