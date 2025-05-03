package main

import (
	db "first-project/DB"
	"first-project/routes"
	"first-project/utils"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

  db.DbInit()
  port := os.Getenv("PORT")

  if port == ""{
    port = "8080"
  }

  
  router := gin.Default()

  // Connect helper function
  router.SetFuncMap(utils.TemplateFuncs())
  
  // Load static files
	router.Static("/static", "./static")

	// Load html
	router.LoadHTMLGlob("templates/**/*")

  routes.GetUrl(router)
  
  router.Run(":"+port)
  
}