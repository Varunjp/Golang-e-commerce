package main

import (
	db "first-project/DB"
	"first-project/routes"
	"first-project/utils"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {

  db.DbInit()
  port := os.Getenv("PORT")

  if port == ""{
    port = "8080"
  }

  
  router := gin.Default()

  // creating session
  store := cookie.NewStore([]byte("secret-key"))
  router.Use(sessions.Sessions("Mysession",store))

  // Connect helper function
  router.SetFuncMap(utils.TemplateFuncs())
  
  // Load static files
	router.Static("/static", "./static")
  router.Static("/upload", "./upload")

	// Load html
	router.LoadHTMLGlob("templates/**/*")

  routes.GetUrl(router)
  
  router.Run(":"+port)
  
}