package model

type ParishResponse struct {
	ID      int              `json:"id"`
	Name    string           `json:"name"`
	Diocese *DioceseResponse `json:"diocese,omitempty"`
}

type CreateParishRequest struct {
	Name      string `json:"name" validate:"required"`
	DioceseID *int   `json:"dioceseId,omitempty"`
}

type UpdateParishRequest struct {
	Name      string `json:"name" validate:"required"`
	DioceseID *int   `json:"dioceseId,omitempty"`
}

