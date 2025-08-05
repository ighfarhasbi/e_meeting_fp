package main

// @title E-Meeting API
// @version 1.0
// @description API for Booking Meeting Room application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email uD2lG@example.com
// @license.name Apache 2.0

// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @schemes http

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
import (
	"e_meeting/config"
	"e_meeting/internal/handlers"
	"e_meeting/internal/middleware"
	"e_meeting/pkg/db"
	"e_meeting/pkg/utils"
	"net/http"

	_ "e_meeting/docs"

	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
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
		return c.JSON(http.StatusOK, utils.SuccessResponse{
			Data:    "Welcome to the E-Meeting API",
			Message: "Success",
		})
	})
	// set up CORS middleware
	e.Use(echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
			c.Response().Header().Set(echo.HeaderAccessControlAllowMethods, "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set(echo.HeaderAccessControlAllowHeaders, "Content-Type, Authorization")
			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	}))

	// set up Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// apply JWT middleware
	group := e.Group("")
	group.Use(middleware.JwtMiddleware)
	// initialize handlers
	handlers.InitRoomHandler(group, conn)   // initialize room handler
	handlers.InitSnacksHandler(group, conn) // initialize snacks handler
	handlers.InitUserHandler(group, conn)   // initialize user handler
	handlers.InitUserAuthHandler(e, conn)   // initialize user auth handler

	// start the server
	e.Logger.Fatal(e.Start(":" + cgf.Port))
}
