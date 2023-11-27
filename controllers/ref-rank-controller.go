package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/auth/initializers"
	"github.com/auth/models"
)

// AddReferralToUser increments the referral count for a user.
func AddReferralToUser(c *gin.Context) {
	// Define a struct to represent the request body for adding a referral to a user
	type AddReferralToUserRequest struct {
		UserID uint `json:"user_id"`
	}

	var requestBody AddReferralToUserRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Find the user by UserID
	var user models.User
	if err := initializers.DB.First(&user, requestBody.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Increment the referral count
	user.RefRank++

	// Update the user's referral rank in the database
	result := initializers.DB.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update referral count",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Referral added to user successfully",
		"user":    user,
	})
}

// DetermineReferralRank determines the referral rank for a user based on the number of referrals.
func DetermineReferralRank(c *gin.Context) {
	// Define a struct to represent the request body for determining referral rank
	type DetermineReferralRankRequest struct {
		UserID uint `json:"user_id"`
	}

	var requestBody DetermineReferralRankRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Find the user by UserID
	var user models.User
	if err := initializers.DB.First(&user, requestBody.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Determine the referral rank based on the number of referrals
	var rank string

	switch {
	case user.RefRank >= 500:
		rank = "Purple Rank"
	case user.RefRank >= 100:
		rank = "Gold Rank"
	case user.RefRank >= 50:
		rank = "Orange Rank"
	case user.RefRank >= 20:
		rank = "Blue Rank"
	case user.RefRank >= 5:
		rank = "Green Rank"
	default:
		rank = "No Rank"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Referral rank determined successfully",
		"rank":    rank,
	})
}
