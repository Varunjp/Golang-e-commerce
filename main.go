package main

import (
	db "first-project/DB"
	"first-project/routes"
	"first-project/utils"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {


  err := godotenv.Load()
  if err != nil{
    log.Fatal("Error while loading env file")
  }



  db.DbInit()
  port := os.Getenv("PORT")


  if port == ""{
    port = "8080"
  }

  
  router := gin.Default()


  router.Use(gin.Logger(), gin.Recovery())

  // size constrain
  router.MaxMultipartMemory = 8 << 20

  // creating session
  store := cookie.NewStore([]byte("secret-key"))
  router.Use(sessions.Sessions("Mysession",store))

  // Connect helper function
  router.SetFuncMap(utils.TemplateFuncs())
  
  // Load static files
	router.Static("/static", "./static")
  router.Static("/upload", "./upload")
  router.Static("/uploads", "./uploads")

	// Load html
	router.LoadHTMLGlob("templates/**/*")

  routes.GetUrl(router)
  
  router.Run(":"+port)
  
}