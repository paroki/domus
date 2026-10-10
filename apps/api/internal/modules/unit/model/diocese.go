package model

import (
	"time"

	"github.com/paroki/domus/api/internal/core"
)

type DioceseResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedBy core.User `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedBy core.User `json:"updatedBy"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateDioceseRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateDioceseRequest struct {
	Name string `json:"name" validate:"required"`
}
