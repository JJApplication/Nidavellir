package initializer

import (
	"fmt"
	"go.uber.org/zap"
	log "nidavellir/pkg/logger"
)

func InitializeLogger(glb *Global) {
	logger := log.NewLogger()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			fmt.Printf("sync logger error: %v\n", err)
		}
	}(logger)

	glb.Logger = logger
}
