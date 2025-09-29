package services

import (
	"alati_projekat/model"
	"time"

	kitmetrics "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type MetricsService struct {
	Next           Service
	RequestCount   kitmetrics.Counter
	RequestLatency kitmetrics.Histogram
}

func NewMetricsService(next Service) *MetricsService {
	// Counter
	requestCount := prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "config_service",
		Subsystem: "configuration_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, []string{"method"})

	// Histogram
	requestLatency := prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: "config_service",
		Subsystem: "configuration_service",
		Name:      "request_latency_seconds",
		Help:      "Total duration of requests in seconds.",
	}, []string{"method"})

	return &MetricsService{
		Next:           next,
		RequestCount:   requestCount,
		RequestLatency: requestLatency,
	}
}

// Generalna defer funkcija za merenje i brojanje
func (s *MetricsService) measure(method string, start time.Time) {
	s.RequestCount.With("method", method).Add(1)
	s.RequestLatency.With("method", method).Observe(time.Since(start).Seconds())
}

// Konfiguracije
func (s *MetricsService) AddConfiguration(config model.Configuration, idempotencyKey string) (err error) {
	defer s.measure("AddConfiguration", time.Now())
	return s.Next.AddConfiguration(config, idempotencyKey)
}

func (s *MetricsService) GetConfiguration(name string, version string) (out model.Configuration, err error) {
	defer s.measure("GetConfiguration", time.Now())
	return s.Next.GetConfiguration(name, version)
}

func (s *MetricsService) UpdateConfiguration(config model.Configuration, idempotencyKey string) (err error) {
	defer s.measure("UpdateConfiguration", time.Now())
	return s.Next.UpdateConfiguration(config, idempotencyKey)
}

func (s *MetricsService) DeleteConfiguration(name string, version string) (err error) {
	defer s.measure("DeleteConfiguration", time.Now())
	return s.Next.DeleteConfiguration(name, version)
}

// Grupe Konfiguracija
func (s *MetricsService) AddConfigurationGroup(group model.ConfigurationGroup, idempotencyKey string) (err error) {
	defer s.measure("AddConfigurationGroup", time.Now())
	return s.Next.AddConfigurationGroup(group, idempotencyKey)
}

func (s *MetricsService) GetConfigurationGroup(name string, version string) (out model.ConfigurationGroup, err error) {
	defer s.measure("GetConfigurationGroup", time.Now())
	return s.Next.GetConfigurationGroup(name, version)
}

func (s *MetricsService) UpdateConfigurationGroup(group model.ConfigurationGroup, idempotencyKey string) (err error) {
	defer s.measure("UpdateConfigurationGroup", time.Now())
	return s.Next.UpdateConfigurationGroup(group, idempotencyKey)
}

func (s *MetricsService) DeleteConfigurationGroup(name string, version string) (err error) {
	defer s.measure("DeleteConfigurationGroup", time.Now())
	return s.Next.DeleteConfigurationGroup(name, version)
}

// Idempotentnost
func (s *MetricsService) CheckIdempotencyKey(key string) (bool, error) {
	return s.Next.CheckIdempotencyKey(key)
}

func (s *MetricsService) SaveIdempotencyKey(key string) {
	s.Next.SaveIdempotencyKey(key)
}
