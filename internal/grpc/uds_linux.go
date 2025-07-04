//go:build linux

package grpc

import (
	"go.uber.org/zap"
	"net"
)

func (s *Server) ServeUDS(unixAddr string) error {
	addr, err := net.ResolveUnixAddr("unix", unixAddr)
	if err != nil {
		s.logger.Fatal("Failed to listen gRPC uds", zap.Error(err))
	}
	udsLis, err := net.ListenUnix("unix", addr)
	if err != nil {
		s.logger.Fatal("Failed to listen gRPC uds", zap.Error(err))
	}
	s.logger.Info("Starting gRPC UDS server", zap.String("addr", unixAddr))
	return s.grpcServer.Serve(udsLis)
}
