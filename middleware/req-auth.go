package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/plato5-auth/initializers"
	"github.com/plato5-auth/models"
)

func ReqAuth(c *gin.Context) {
	//	Get the cookie
	tokenString, err := c.Cookie("AuthZ")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	//	Decode
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		//	Check exp
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		//	Find the user
		var user models.User
		initializers.DB.First(&user, claims["sub"])

		if user.ID == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		//	Attach to req
		c.Set("user", user)

		//	Continue
		c.Next()
		fmt.Println(claims["foo"], claims["nbh"])
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

}
