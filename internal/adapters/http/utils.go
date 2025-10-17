package http

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	domain "assets-service/internal/core/domain"
)

// Helper methods
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *HTTPHandler) customError(w http.ResponseWriter, status int, domainErr domain.DomainError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": domainErr.Message, "code": string(domainErr.Code)})
}

func (h *HTTPHandler) writeValidationError(w http.ResponseWriter, status int, errors domain.ValidationErrors) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"errors": errors})
}

func (h *HTTPHandler) responseWithError(w http.ResponseWriter, status int, err error) {
	if domainErr, ok := err.(*domain.DomainError); ok {
		h.customError(w, http.StatusBadRequest, *domainErr)
	} else {
		h.writeError(w, status, err.Error())
	}
}

// Helper method to get client IP
func (h *HTTPHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP if multiple are present
		if idx := strings.Index(forwarded, ","); idx > 0 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

func (h *HTTPHandler) getUserAgent(r *http.Request) string {
	if r.UserAgent() != "" {
		return r.UserAgent()
	}
	return "unknown"
}

func (h *HTTPHandler) getDeviceID(r *http.Request) string {
	if deviceID := r.Header.Get("Device-ID"); deviceID != "" {
		return deviceID
	}
	return "unknown"
}

func (h *HTTPHandler) logError(err error, msg string, r *http.Request) {
	h.logger.Error(msg,
		"error", err.Error(),
		"ip", h.getClientIP(r),
		"user_agent", h.getUserAgent(r),
		"device_id", h.getDeviceID(r),
	)
}

