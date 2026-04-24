package router

import (
	"net/http"

	"JumysTab/internal/handler"
	"JumysTab/internal/middleware"
)

func New(auth *handler.AuthHandler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	// Auth routes (public)
	mux.HandleFunc("POST /api/auth/register", auth.Register)
	mux.HandleFunc("POST /api/auth/login/request", auth.RequestOTP)
	mux.HandleFunc("POST /api/auth/login/verify", auth.VerifyOTP)
	mux.HandleFunc("GET /api/auth/status", auth.ActivationStatus)

	// Protected routes
	protected := middleware.Auth(jwtSecret)
	mux.Handle("GET /api/profile", protected(http.HandlerFunc(auth.GetProfile)))

	// Wrap entire mux with CORS
	return corsMiddleware(mux)
}

// corsMiddleware adds CORS headers for frontend interaction.
// In production, restrict AllowedOrigins to your actual domain.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
