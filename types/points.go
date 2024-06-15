package types

import "gorm.io/gorm"

type Points struct {
	gorm.Model
	ID      string `gorm:"type:uuid;primary_key;"`
	UserID  string
	User    User `gorm:"foreignKey:UserID;constraint:OnUpdate:SET NULL,OnDelete:CASCADE;"`
	Balance int32
}

type UpdatePointsRequestBody struct {
	ID         string `json:"id,omitempty"`
	NewBalance int32  `json:"new_balance,omitempty"`
}

type UpdatePointsResponseBody struct {
	Status bool
}

type CreatePointsAccountRequestBody struct {
	UserID     *string `json:"user_id,omitempty"`
	NewBalance *int32  `json:"new_balance,omitempty"`
}

type DeletePointsAccountRequestBody struct {
	ID *string `json:"id,omitempty"`
}
