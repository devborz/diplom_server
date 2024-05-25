package server

import (
	"authentication"
	"db"
	"handlers"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Port string
}

func (s *Server) Run() {
	db.DB.Connect()
	r := gin.Default()

	publicRoutes := r.Group("/auth")
	{
		publicRoutes.POST("/login", handlers.Login)
		publicRoutes.POST("/register", handlers.Register)
		publicRoutes.POST("/logout", handlers.Logout)
	}

	protectedRoutes := r.Group("/v1")
	protectedRoutes.Use(authentication.Authentication())
	{
		protectedRoutes.POST("/resources/:owner_id", handlers.AddObject)
		protectedRoutes.GET("/resources/:owner_id", handlers.GetResource)
		protectedRoutes.PUT("/resources/:owner_id", handlers.CreateDirectory)
		protectedRoutes.DELETE("/resources/:owner_id", handlers.DeleteResource)

		protectedRoutes.POST("/rights", handlers.ShareRights)
		protectedRoutes.DELETE("/rights", handlers.DeleteRights)
		protectedRoutes.GET("/resources/:owner_id/access", handlers.GetUsersWithAccess)
        protectedRoutes.GET("/sharedresources", handlers.GetUsersSharedResources)
	}

	r.Run(":8080")
}
