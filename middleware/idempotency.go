package middleware

import (
	"alati_projekat/services"
	"log"
	"net/http"
)

// IdempotencyMiddleware provides idempotency for POST and PUT requests
type IdempotencyMiddleware struct {
	Service services.Service
}

// NewIdempotencyMiddleware creates a new instance of the idempotency middleware
func NewIdempotencyMiddleware(service services.Service,
) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		Service: service,
	}
}

// Middleware returns the HTTP handler with idempotency logic
func (im *IdempotencyMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		idempotencyKey := r.Header.Get("X-Request-Id")

		if idempotencyKey == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error: X-Request-Id header (UUID) is necessary."))
			return
		}

		isProcessed, err := im.Service.CheckIdempotencyKey(idempotencyKey)
		if err != nil {
			log.Printf("IDEMPOTENCY ERROR: Consul check failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if isProcessed {
			log.Printf("IDEMPOTENCY HIT: Request with key %s already processed.", idempotencyKey)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Request already processed (Idempotent)."))
			return
		}

		log.Printf("IDEMPOTENCY MISS: Processing new request with key %s.", idempotencyKey)
		next.ServeHTTP(w, r)
	})
}
