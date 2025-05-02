package main

import (
	db "first-project/DB"
	"first-project/routes"
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

  routes.GetUrl(router)
  
  router.Run(":"+port)
  
}