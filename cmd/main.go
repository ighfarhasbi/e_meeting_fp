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
	delivery "e_meeting/internal/delivery/users"
	"e_meeting/internal/handlers"
	"e_meeting/internal/middlewareAuth"
	repository "e_meeting/internal/repository/users"
	usecase "e_meeting/internal/usecase/users"
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
	redisConn, err := db.NewRedis(cfg.RedisUrl) // connect to redis
	if err != nil {
		panic("Failed to connect to redis: " + err.Error())
	}
	defer redisConn.Close()

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

	// DELIVERY -> HANDLER -> USECASE -> ENTITY
	// login & register
	authRepo := repository.NewDBUsersRepository(conn) // isinya query ke db
	userUC := usecase.NewUserUsecase(authRepo)
	delivery.NewUsersHandler(e, userUC)
	// reset password
	resetPassRepo := repository.NewDBResetPassRepository(conn)
	resetPassUsecase := usecase.NewResetPassUsecase(resetPassRepo)
	delivery.NewResetPassHandler(e, resetPassUsecase)

	// belum implementasi clean architecture
	handlers.InitDashboardHandler(group, conn)              // initialize dashboard handler
	handlers.InitUploadHandler(group)                       // initialize upload handler
	handlers.InitReservationHandler(group, conn, redisConn) // initialize reservation handler
	handlers.InitRoomHandler(group, conn)                   // initialize room handler
	handlers.InitSnacksHandler(group, conn)                 // initialize snacks handler
	handlers.InitUserHandler(group, conn)                   // initialize user handler
	// handlers.InitUserAuthHandler(e, conn)                   // initialize user auth handler

	// start the server
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
