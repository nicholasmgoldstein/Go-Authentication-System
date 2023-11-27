package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/plato5-auth/controllers"
	"github.com/plato5-auth/initializers"
	"github.com/plato5-auth/middleware"
)

func init() {
	initializers.LoadEnvVariables()
	//connectDB() & syncDB() should go here during prod
}

func main() {
	r := gin.Default()
	initializers.ConnectDB()
	initializers.SyncDB()
	//connectDB() & syncDB() are meant to be in init() during prod

	//AuthN Controls
	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	r.POST("/logout", controllers.Logout)
	r.GET("/validate", middleware.ReqAuth, controllers.Validate)

	//User Admin Controls
	r.POST("/update-country-and-location", controllers.UpdateCountryAndLocation)
	r.GET("/get-users-in-location", controllers.GetUsersInLocation)
	r.DELETE("/delete-user", controllers.DeleteUser)

	//User Permissions Controls
	r.POST("/alter-user-deactivation", controllers.AlterUserDeactivation)
	r.POST("/alter-commenting-permissions", controllers.AlterCommentingPermissions)
	r.POST("/alter-posting-permissions", controllers.AlterPostingPermissions)
	r.POST("/alter-analytix-permissions", controllers.AlterAnalytixPermissions)
	r.GET("/check-user-permissions", controllers.CheckUserPermissions)

	//Referral Rank Controls
	r.POST("/add-referral-to-user", controllers.AddReferralToUser)
	r.GET("/determine-referral-rank", controllers.DetermineReferralRank)

	//Search
	r.GET("/search-users", controllers.SearchUsers)

	fmt.Println("AuthN & AuthZ Microservice v0.01")
	r.Run()
}
