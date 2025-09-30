package services

import (
	"alati_projekat/model"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("service-tracing-layer")

type TracingService struct {
	Next Service
}

func NewTracingService(next Service) *TracingService {
	return &TracingService{
		Next: next,
	}
}

// Pomoćna funkcija za postavljanje Error Statusa
func endSpan(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// --- CONFIGURATIONS ---

func (s *TracingService) AddConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) (err error) {
	ctx, span := tracer.Start(ctx, "AddConfigurationService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("config.name", config.Name), attribute.String("config.version", config.Version))
	return s.Next.AddConfiguration(ctx, config, idempotencyKey)
}

func (s *TracingService) GetConfiguration(ctx context.Context, name string, version string) (out model.Configuration, err error) {
	ctx, span := tracer.Start(ctx, "GetConfigurationService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("config.name", name), attribute.String("config.version", version))
	return s.Next.GetConfiguration(ctx, name, version)
}

func (s *TracingService) UpdateConfiguration(ctx context.Context, config model.Configuration, idempotencyKey string) (err error) {
	ctx, span := tracer.Start(ctx, "UpdateConfigurationService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("config.name", config.Name), attribute.String("config.version", config.Version))
	return s.Next.UpdateConfiguration(ctx, config, idempotencyKey)
}

func (s *TracingService) DeleteConfiguration(ctx context.Context, name string, version string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteConfigurationService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("config.name", name), attribute.String("config.version", version))
	return s.Next.DeleteConfiguration(ctx, name, version)
}

// --- CONFIGURATION GROUPS ---

func (s *TracingService) AddConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) (err error) {
	ctx, span := tracer.Start(ctx, "AddConfigurationGroupService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("group.name", group.Name), attribute.String("group.version", group.Version))
	return s.Next.AddConfigurationGroup(ctx, group, idempotencyKey)
}

func (s *TracingService) GetConfigurationGroup(ctx context.Context, name string, version string) (out model.ConfigurationGroup, err error) {
	ctx, span := tracer.Start(ctx, "GetConfigurationGroupService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("group.name", name), attribute.String("group.version", version))
	return s.Next.GetConfigurationGroup(ctx, name, version)
}

func (s *TracingService) UpdateConfigurationGroup(ctx context.Context, group model.ConfigurationGroup, idempotencyKey string) (err error) {
	ctx, span := tracer.Start(ctx, "UpdateConfigurationGroupService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("group.name", group.Name), attribute.String("group.version", group.Version))
	return s.Next.UpdateConfigurationGroup(ctx, group, idempotencyKey)
}

func (s *TracingService) DeleteConfigurationGroup(ctx context.Context, name string, version string) (err error) {
	ctx, span := tracer.Start(ctx, "DeleteConfigurationGroupService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("group.name", name), attribute.String("group.version", version))
	return s.Next.DeleteConfigurationGroup(ctx, name, version)
}

// --- IDEMPOTENCY ---

func (s *TracingService) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckIdempotencyKeyService")
	defer span.End() // Ne treba nam endSpan jer ova funkcija nema grešku u potpisu
	span.SetAttributes(attribute.String("idempotency.key", key))
	return s.Next.CheckIdempotencyKey(ctx, key)
}

func (s *TracingService) SaveIdempotencyKey(ctx context.Context, key string) {
	ctx, span := tracer.Start(ctx, "SaveIdempotencyKeyService")
	defer span.End()
	span.SetAttributes(attribute.String("idempotency.key", key))
	s.Next.SaveIdempotencyKey(ctx, key)
}

// --- LABELS ---

func (s *TracingService) FilterConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (out []model.Configuration, err error) {
	ctx, span := tracer.Start(ctx, "FilterConfigsByLabelsService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("group.name", name), attribute.String("group.version", version))
	return s.Next.FilterConfigsByLabels(ctx, name, version, want)
}

func (s *TracingService) DeleteConfigsByLabels(ctx context.Context, name, version string, want map[string]string) (deleted int, err error) {
	ctx, span := tracer.Start(ctx, "DeleteConfigsByLabelsService")
	defer endSpan(span, err)
	span.SetAttributes(attribute.String("group.name", name), attribute.String("group.version", version))
	return s.Next.DeleteConfigsByLabels(ctx, name, version, want)
}
