package initializer

import (
	log "nidavellir/pkg/logger"
)

func InitializeLogger(glb *Global) {
	logger := log.NewLogger()
	glb.Logger = logger
}
