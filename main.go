package main

import (
	"e_meeting/models"
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, models.Response{
			Data:    "Welcome to the E-Meeting API",
			Message: "Success",
		})
	})
	e.Logger.Fatal(e.Start(":8080"))
}
