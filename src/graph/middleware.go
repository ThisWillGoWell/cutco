package graph

import (
	"context"
	"net/http"
	"stock-simulator-serverless/src/models"
)

func AttachContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add a requestID uuid
		ctx := context.WithValue(r.Context(), "request-id", models.NewUUID())

		header := r.Header.Get("Authorization")
		// if auth is not available then proceed to resolver
		if header == "" {
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// merge token onto the ctx
			ctx = context.WithValue(ctx, "token", header)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
