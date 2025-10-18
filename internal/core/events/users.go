package events

import "assets-service/internal/core/domain"

type UserCreatedEvent struct {
	UserID            string          `json:"user_id"`
	Name              string          `json:"name"`
	Email             *string         `json:"email"`
	Phone             *string         `json:"phone"`
	Locale            *string         `json:"locale"`
	Country           *string         `json:"country"`
	DeviceID          *string         `json:"device_id"`
	NotificationToken *string         `json:"notification_token,omitempty"`
	UserType          domain.UserType `json:"user_type"` // e.g., "customer", "driver"
}

type UserUpdatedEvent struct {
	UserID    string  `json:"user_id"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}
