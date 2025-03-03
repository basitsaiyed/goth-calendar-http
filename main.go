package main

import (
	"goauthDemo/internal/auth"
	"goauthDemo/middleware"
	"goauthDemo/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	auth.NewAuth()
	r := gin.Default()

	r.GET("/", routes.Home)
	r.GET("/auth/:provider", routes.AuthProvider)
	r.GET("/auth/:provider/callback", routes.AuthCallback)
	r.GET("/schedule-meeting", routes.ScheduleMeeting)
	r.POST("/create-meeting", middleware.JWTAuthMiddleware(), routes.CreateMeeting)
	r.GET("/upcoming-meetings",middleware.JWTAuthMiddleware(), routes.GetUpcomingMeetings)

	r.Run()
}
