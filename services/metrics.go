package services

import (
	"alati_projekat/model"
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type MetricsService struct {
	Next           Service
	RequestCount   *stdprometheus.CounterVec
	RequestLatency *stdprometheus.HistogramVec
}

func NewMetricsService(next Service) *MetricsService {
	// Counter
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

func (s *MetricsService) CheckIdempotencyKey(key string) (bool, error) {
	return s.Next.CheckIdempotencyKey(key)
}

func (s *MetricsService) SaveIdempotencyKey(key string) {
	s.Next.SaveIdempotencyKey(key)
}
func (s *MetricsService) FilterConfigsByLabels(name, version string, want map[string]string) (out []model.Configuration, err error) {
	defer s.measure("FilterConfigsByLabels", time.Now())
	return s.Next.FilterConfigsByLabels(name, version, want)
}

func (s *MetricsService) DeleteConfigsByLabels(name, version string, want map[string]string) (deleted int, err error) {
	defer s.measure("DeleteConfigsByLabels", time.Now())
	return s.Next.DeleteConfigsByLabels(name, version, want)
}
