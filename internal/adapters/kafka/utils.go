package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// eventToMap converts an event struct to a map
func eventToMap(event interface{}) map[string]interface{} {
	data, _ := json.Marshal(event)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// getCorrelationID extracts correlation ID from context or generates one
func getCorrelationID(ctx context.Context) string {
	if corrID := ctx.Value("correlation_id"); corrID != nil {
		if id, ok := corrID.(string); ok {
			return id
		}
	}
	return generateEventID()
}
