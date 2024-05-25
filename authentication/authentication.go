package authentication

import (
	. "clouderrors"
	. "db"
	"net/http"
	"strings"
	"utils"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
            c.Abort()
            return
        }

        // The token should be prefixed with "Bearer "
        tokenParts := strings.Split(tokenString, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, ErrInvalidAuthenticationToken.Error())
            c.Abort()
            return
        }

        tokenString = tokenParts[1]

        claims, err := utils.VerifyToken(tokenString)

        if err != nil {
            c.JSON(http.StatusUnauthorized, ErrInvalidAuthenticationToken.Error())
            c.Abort()
            return
        }
        uid_ := claims["user_id"]
        if uid_, ok := uid_.(float64); ok {
            uid := int64(uid_)
            if err != nil {
                c.JSON(http.StatusUnauthorized, ErrInvalidAuthenticationToken.Error())
                c.Abort()
            }
            c.Set("user_id", uid)
            c.Set("token", tokenString)
            if DB.CheckToken(uid, tokenString) {
                c.Next()
            } else {
                c.JSON(http.StatusUnauthorized, ErrInvalidAuthenticationToken.Error())
                c.Abort()
            }
        } else {
            c.JSON(http.StatusUnauthorized, ErrInvalidAuthenticationToken.Error())
            c.Abort()
        }
    }
}