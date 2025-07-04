//go:build windows

package grpc

import (
	"go.uber.org/zap"
)

func (s *Server) ServeUDS(unixAddr string) error {
	s.logger.Warn("Unsupported OS, gRPC UDS server will not run", zap.String("addr", unixAddr))
	return nil
}
