package domain

type BaseModel struct {
	ID          string  `json:"id"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
	DeletedByID *string `json:"deleted_by_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Active      bool    `json:"active"`
	CreatedByID *string `json:"created_by_id,omitempty"`
	UpdatedByID *string `json:"updated_by_id,omitempty"`
}
