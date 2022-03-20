package service

import (
	"context"

	health "google.golang.org/grpc/health/grpc_health_v1"
)

type healthServer struct {
	health.HealthServer
}

func NewHealthServer() health.HealthServer {
	return &healthServer{}
}
func (s *healthServer) Check(_ context.Context, _ *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *health.HealthCheckRequest, server health.Health_WatchServer) error {
	// Example of how to register both methods but only implement the Check method.
	return server.Send(&health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	})
}
