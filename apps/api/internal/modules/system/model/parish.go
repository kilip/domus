package model

type ParishResponse struct {
	ID      int              `json:"id"`
	Name    string           `json:"name"`
	Active  bool             `json:"active"`
	Church  *string          `json:"church,omitempty"`
	Address *string          `json:"address,omitempty"`
	Phone   *string          `json:"phone,omitempty"`
	Email   *string          `json:"email,omitempty"`
	Diocese *DioceseResponse `json:"diocese,omitempty"`
}

type CreateParishRequest struct {
	Name      string  `json:"name" validate:"required"`
	Active    *bool   `json:"active,omitempty"`
	Church    *string `json:"church,omitempty"`
	Address   *string `json:"address,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	DioceseID *int    `json:"dioceseId,omitempty"`
}

type UpdateParishRequest struct {
	Name      string  `json:"name" validate:"required"`
	Active    *bool   `json:"active,omitempty"`
	Church    *string `json:"church,omitempty"`
	Address   *string `json:"address,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	DioceseID *int    `json:"dioceseId,omitempty"`
}
