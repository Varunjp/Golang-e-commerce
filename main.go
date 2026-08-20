package main

import (
	db "first-project/DB"
	"first-project/routes"
	"first-project/utils"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {

	_ = godotenv.Load()

	db.DbInit()

	// delete
	fmt.Println("db added")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	logFile := &lumberjack.Logger{
		Filename:   "logs/gin.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	multiWriter := io.MultiWriter(logFile, os.Stdout)
	gin.DefaultWriter = multiWriter // Gin's own logs
	log.SetOutput(multiWriter)

	router := gin.New()

	router.Use(gin.Logger(), gin.Recovery())

	// size constrain
	router.MaxMultipartMemory = 8 << 20

	// creating session
	store := cookie.NewStore([]byte("secret-key"))
	router.Use(sessions.Sessions("Mysession", store))

	// Connect helper function
	router.SetFuncMap(utils.TemplateFuncs())

	// Load static files
	router.Static("/static", "./static")
	router.Static("/upload", "./upload")
	router.Static("/uploads", "./uploads")

	// Load html
	router.LoadHTMLGlob("templates/**/*")

	routes.GetUrl(router)

	router.Run(":" + port)

}
