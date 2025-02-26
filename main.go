package main

import (
	"context"
	"fmt"
	"goauthDemo/internal/auth"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func main() {
	auth.NewAuth()
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		tmpl, _ := template.ParseFiles("index.html")
		tmpl.Execute(c.Writer, nil)
	})

	r.GET("/auth/:provider", func(c *gin.Context) {
		provider := c.Param("provider")
		if provider == "" {
			c.JSON(400, gin.H{"error": "Provider is required"})
			return
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	r.GET("/auth/:provider/callback", func(c *gin.Context) {
		provider := c.Param("provider")
		if provider == "" {
			c.JSON(400, gin.H{"error": "Provider is required"})
			return
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "provider", provider))
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		tmpl, _ := template.ParseFiles("hello.html")
		tmpl.Execute(c.Writer, user)
		fmt.Println("User: ", user)
		c.JSON(200, gin.H{"User": user})
	})

	r.GET("/logout/:provider", func(c *gin.Context) {
		gothic.Logout(c.Writer, c.Request)
		c.Writer.Header().Set("Location", "/")
		c.Writer.WriteHeader(http.StatusTemporaryRedirect)
	})
	r.Run()
}
