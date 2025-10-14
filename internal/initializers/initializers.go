package initializers

import (
	"log"
	"online-subs/pkg/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func startGetEnv() {
	err := godotenv.Load(".env")

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
		log.Fatalf("Error initializing zap logger: %v", err)
	}

	return db
}

func initSubsRouter(handler *handlers.SubsHandler) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

	subsGroup := r.Group("/subscriptions")

	subsGroup.GET("/get/query", handler.GetByParams())
	subsGroup.GET("/get/:id", handler.GetSubByID())
	subsGroup.GET("/list", handler.List())
	subsGroup.GET("/total", handler.GetTotalCost())

	subsGroup.POST("/create", handler.CreateSub())
	subsGroup.PATCH("/update", handler.UpdateSub())

	subsGroup.DELETE("/delete/:id", handler.DeleteSub())

	return r
}
