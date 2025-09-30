package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware je middleware koji prima kontekst praćenja iz dolaznog zahteva
// i kreira glavni span za taj zahtev.
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Izdvoji kontekst praćenja iz headera zahteva
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// 2. Kreiraj glavni Span
		opts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPTargetKey.String(r.URL.Path),
				semconv.NetHostNameKey.String(r.Host),
				attribute.String("client.ip", r.RemoteAddr),
			),
		}

		ctx, span := otel.Tracer("http-server-router").Start(ctx, r.URL.Path, opts...)
		defer span.End()

		// 3. Ubaci novi kontekst (sa span-om) u request
		r = r.WithContext(ctx)

		// 4. Prosledi zahtev dalje u lanac handlera
		next.ServeHTTP(w, r)
	})
}
