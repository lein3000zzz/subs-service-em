package initializers

import (
	"log"
	"online-subs/pkg/handlers"
	"online-subs/pkg/subs"
	"os"

	"go.uber.org/zap"
)

func RunSubsService() {
	startGetEnv()

	zapLogger := startLogger()
	defer func(zapLogger *zap.Logger) {
		err := zapLogger.Sync()
		if err != nil {
			log.Fatal("Error syncing zap logger:", err)
		}
	}(zapLogger)

	logger := zapLogger.Sugar()

	db := startPostgres()
	gormAutoMigrate(db)

	subsRepo := subs.NewSubscriptionsPgRepo(logger, db)

	subsHandler := handlers.NewSubsHandler(subsRepo, logger)

	r := initSubsRouter(subsHandler)

	logger.Fatal(r.Run(":" + os.Getenv("PORT")))
}
