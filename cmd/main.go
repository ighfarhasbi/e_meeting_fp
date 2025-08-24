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

// @host localhost:8085
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
import (
	"e_meeting/config"
	"e_meeting/internal/delivery"
	"e_meeting/internal/middlewareAuth"
	"e_meeting/internal/repository"
	"e_meeting/internal/usecase"
	"e_meeting/pkg/db"
	"log"
	"os"

	"github.com/labstack/echo/v4/middleware"

	_ "e_meeting/docs"

	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	// Jika APP_ENV=local, dijalankan manual saat run (APP_ENV=local go run cmd/server/main.go)
	if os.Getenv("APP_ENV") == "local" {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Fatalf("Error loading .env.local: %v", err)
		}
	}

	// load config
	cfg := config.New()

	// connect to database
	conn, err := db.NewPostgres(cfg.DBUrl) // connect to postgres
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer conn.Close()

	// initialize echo framework
	e := echo.New()
	// set up middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	// set up Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// apply JWT middleware
	group := e.Group("")
	group.Use(middlewareAuth.JwtMiddleware)

	// interface -> handler -> usecase -> entity
	userRepo := repository.NewDBUsersRepository(conn) // isinya query ke db
	userUC := usecase.NewUserUsecase(userRepo)
	delivery.NewUsersHandler(e, userUC)

	// start the server
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
