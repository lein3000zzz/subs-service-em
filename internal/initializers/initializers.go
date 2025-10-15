package initializers

import (
	"log"
	"online-subs/docs"
	"online-subs/pkg/handlers"
	"online-subs/pkg/subs"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func startGetEnv() {
	if os.Getenv("ENVIRONMENT") == "PROD" {
		return
	}

	err := godotenv.Load("local.env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func startLogger() *zap.Logger {
	zapLogger, err := zap.NewProduction()

	if err != nil {
		log.Fatalf("Error initializing zap logger: %v", err)
	}

	return zapLogger
}

func startPostgres() *gorm.DB {
	dsn := os.Getenv("PG_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error initializing postgres: %v", err)
	}

	return db
}

func initSubsRouter(handler *handlers.SubsHandler) *gin.Engine {
	r := gin.Default()

	host := "localhost:" + os.Getenv("PORT")
	docs.SwaggerInfo.Host = host
	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	subsGroup := r.Group("/subscriptions/v1")

	subsGroup.GET("/get/query", handler.GetByParams)
	subsGroup.GET("/get/:id", handler.GetSubByID)
	subsGroup.GET("/list", handler.List)
	subsGroup.GET("/total", handler.GetTotalCost)

	subsGroup.POST("/create", handler.CreateSub)
	subsGroup.PATCH("/update/:id", handler.UpdateSub)

	subsGroup.DELETE("/delete/:id", handler.DeleteSub)

	return r
}

func gormAutoMigrate(db *gorm.DB) {
	if os.Getenv("ENVIRONMENT") != "LOCAL" {
		return
	}

	if errAuto := db.AutoMigrate(
		&subs.Subscription{},
	); errAuto != nil {
		log.Fatalf("AutoMigrate failed: %v", errAuto)
		return
	}
}
