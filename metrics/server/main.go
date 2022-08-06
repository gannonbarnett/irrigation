package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	api "github.com/gbarnett/irrigation/metrics/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
    SERVER_PORT = flag.Int("server_port", 50051, "Server port")
    PROM_PORT = flag.Int("prom_port", 9100, "Prometheus port")

    tempGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "irrigation_temp_celsius_gauge",
        Help: "Temperature gauge.",
    })

    soilMoistureGauge = promauto.NewCounter(prometheus.CounterOpts{
        Name: "soil_moisture_gauge",
        Help: "Soil moisture gauge.",
    })
)

type MetricsServer struct {
    api.UnimplementedMetricsServer
}

func (s *MetricsServer) PostMetrics(ctx context.Context, req *api.PostMetricsRequest) (*api.PostMetricsResponse, error) {
    tempGauge.Add(float64(req.GetTempCelsius()))
    soilMoistureGauge.Add(float64(req.GetSoilMoisture()))

    return &api.PostMetricsResponse{}, nil
}

func main() {
    flag.Parse()
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *SERVER_PORT))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    s := grpc.NewServer()
    api.RegisterMetricsServer(s, &MetricsServer{})
    go func() {
        log.Printf("Starting server at %v", lis.Addr())
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    go func() {
        log.Printf("Starting prometheus at %v", *PROM_PORT)
        http.Handle("/metrics", promhttp.Handler())
        if err := http.ListenAndServe(fmt.Sprintf(":%d", *PROM_PORT), nil); err != nil {
            log.Fatalf("Failed to host prom metrics: %v", err)
        }
    }()

    select {}
}
