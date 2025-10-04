package services

import (
	"alati_projekat/model"
	"context"
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type MetricsService struct {
	Next           Service
	RequestCount   *stdprometheus.CounterVec
	RequestLatency *stdprometheus.HistogramVec
}

func NewMetricsService(next Service) *MetricsService {
	requestCount := stdprometheus.NewCounterVec(stdprometheus.CounterOpts{
		Namespace: "app",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Number of requests received.",
	}, []string{"method"})

	requestLatency := stdprometheus.NewHistogramVec(stdprometheus.HistogramOpts{
		Namespace: "config_service",
		Subsystem: "configuration_service",
		Name:      "request_latency_seconds",
		Help:      "Total duration of requests in seconds.",
	}, []string{"method"})

	stdprometheus.MustRegister(requestCount)
	stdprometheus.MustRegister(requestLatency)

	return &MetricsService{
		Next:           next,
		RequestCount:   requestCount,
		RequestLatency: requestLatency,
	}
}

func (s *MetricsService) measure(method string, start time.Time) {
	s.RequestCount.WithLabelValues(method).Inc()
	s.RequestLatency.WithLabelValues(method).Observe(time.Since(start).Seconds())
}

func (s *MetricsService) AddConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) (err error) {
	defer s.measure("AddConfiguration", time.Now())
	return s.Next.AddConfiguration(ctx, config, idempotencyKey)
}

func (s *MetricsService) GetConfiguration(ctx context.Context, name string, version string) (out model.Configuration, err error) {
	defer s.measure("GetConfiguration", time.Now())
	return s.Next.GetConfiguration(ctx, name, version)
}

func (s *MetricsService) UpdateConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) (out model.Configuration, err error) {
	defer s.measure("UpdateConfiguration", time.Now())
	return s.Next.UpdateConfiguration(ctx, config, idempotencyKey)
}

func (s *MetricsService) DeleteConfiguration(ctx context.Context, name string, version string) (err error) {
	defer s.measure("DeleteConfiguration", time.Now())
	return s.Next.DeleteConfiguration(ctx, name, version)
}

func (s *MetricsService) AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) (err error) {
	defer s.measure("AddConfigurationGroup", time.Now())
	return s.Next.AddConfigurationGroup(ctx, group, idempotencyKey)
}

func (s *MetricsService) GetConfigurationGroup(ctx context.Context, name string, version string) (out model.ConfigurationGroup, err error) {
	defer s.measure("GetConfigurationGroup", time.Now())
	return s.Next.GetConfigurationGroup(ctx, name, version)
}

func (s *MetricsService) UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) (out model.ConfigurationGroup, err error) {
	defer s.measure("UpdateConfigurationGroup", time.Now())
	out, err = s.Next.UpdateConfigurationGroup(ctx, group, idempotencyKey)
	return out, err
}

func (s *MetricsService) DeleteConfigurationGroup(ctx context.Context, name string, version string) (err error) {
	defer s.measure("DeleteConfigurationGroup", time.Now())
	return s.Next.DeleteConfigurationGroup(ctx, name, version)
}

func (s *MetricsService) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return s.Next.CheckIdempotencyKey(ctx, key)
}

func (s *MetricsService) SaveIdempotencyKey(ctx context.Context, key string) {
	s.Next.SaveIdempotencyKey(ctx, key)
}

func (s *MetricsService) FilterConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (out []model.Configuration, err error) {
	defer s.measure("FilterConfigsByLabels", time.Now())
	return s.Next.FilterConfigsByLabels(ctx, name, version, want)
}

func (s *MetricsService) DeleteConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (deleted int, err error) {
	defer s.measure("DeleteConfigsByLabels", time.Now())
	return s.Next.DeleteConfigsByLabels(ctx, name, version, want)
}
