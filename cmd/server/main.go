package main

import (
	"e_meeting/config"
	"e_meeting/internal/handlers"
	"e_meeting/pkg/db"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
)

func main() {
	// load .env file
	if err := godotenv.Load(); err != nil {
		panic("Failed to load .env file")
	}

	// load config
	cgf := config.New()

	// connect to database
	conn, err := db.NewPostgres(cgf.DBUrl)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer conn.Close()

	// initialize echo framework
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, utils.Response{
			Data:    "Welcome to the E-Meeting API",
			Message: "Success",
		})
	})

	// initialize handlers
	handlers.InitRoomHandler(e, conn)   // initialize room handler
	handlers.InitSnacksHandler(e, conn) // initialize snacks handler
	handlers.InitUserHandler(e, conn)   // initialize user handler

	// start the server
	e.Logger.Fatal(e.Start(":8080"))
}
