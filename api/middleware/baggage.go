package middleware

import (
	"context"
	"net/http"
)

type baggageKey struct{}

func BaggageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baggage := r.Header.Get("baggage")
		if baggage == "" {
			baggage = r.Header.Get("Baggage")
		}

		if baggage != "" {
			ctx := context.WithValue(r.Context(), baggageKey{}, baggage)
			r = r.WithContext(ctx)

			// Add baggage to response headers for downstream services
			w.Header().Set("baggage", baggage)
		}

		next.ServeHTTP(w, r)
	})
}

func WithBaggage(ctx context.Context, baggage string) context.Context {
	return context.WithValue(ctx, baggageKey{}, baggage)
}

func GetBaggageFromContext(ctx context.Context) string {
	if baggage, ok := ctx.Value(baggageKey{}).(string); ok {
		return baggage
	}
	return ""
}
