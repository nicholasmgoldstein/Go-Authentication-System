package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/auth/initializers"
	"github.com/auth/models"
)

type AlterUserPermissionsRequest struct {
	Email                string `json:"email" binding:"required"`
	UserDeactivated      bool   `json:"user_deactivated"`
	BannedFromCommenting bool   `json:"banned_from_commenting"`
	BannedFromPosting    bool   `json:"banned_from_posting"`
	BannedFromAnalytix   bool   `json:"banned_from_analytix"`
}

func CheckUserPermissions(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read JSON request body",
		})
		return
	}

	var userPermissions models.UserPermissions

	err := initializers.DB.Where("email = ?", request.Email).First(&userPermissions).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User permissions not found",
		})
		return
	}

	c.JSON(http.StatusOK, userPermissions)
}

func AlterUserDeactivation(c *gin.Context) {
	alterUserPermissions(c, "UserDeactivated")
}

func AlterCommentingPermissions(c *gin.Context) {
	alterUserPermissions(c, "BannedFromCommenting")
}

func AlterPostingPermissions(c *gin.Context) {
	alterUserPermissions(c, "BannedFromPosting")
}

func AlterAnalytixPermissions(c *gin.Context) {
	alterUserPermissions(c, "BannedFromAnalytix")
}

func alterUserPermissions(c *gin.Context, fieldToUpdate string) {
	var request AlterUserPermissionsRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read JSON request body",
		})
		return
	}

	var userPermissions models.UserPermissions

	err := initializers.DB.Where("email = ?", request.Email).First(&userPermissions).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User permissions not found",
		})
		return
	}

	switch fieldToUpdate {
	case "UserDeactivated":
		userPermissions.UserDeactivated = request.UserDeactivated
	case "BannedFromCommenting":
		userPermissions.BannedFromCommenting = request.BannedFromCommenting
	case "BannedFromPosting":
		userPermissions.BannedFromPosting = request.BannedFromPosting
	case "BannedFromAnalytix":
		userPermissions.BannedFromAnalytix = request.BannedFromAnalytix
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid field to update",
		})
		return
	}

	err = initializers.DB.Save(&userPermissions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user permissions",
		})
		return
	}

	c.JSON(http.StatusOK, userPermissions)
}
