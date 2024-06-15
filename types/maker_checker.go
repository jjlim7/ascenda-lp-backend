package types

import (
	"encoding/json"
	"time"
)

// Transaction => Database Model
type Transaction struct {
	TransactionId string          `gorm:"type:uuid;primary_key;"`
	Action        json.RawMessage `gorm:"type:json"`
	MakerId       string          `gorm:"type:string;index"`
	Maker         User            `gorm:"foreignKey:MakerId"`
	CheckerId     string          `gorm:"type:string;default:null;index"`
	Checker       User            `gorm:"foreignKey:CheckerId"`
	Status        string          `gorm:"type:string;default:pending"`
	Approval      bool            `gorm:"type:boolean;default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ApprovalChainMap struct {
	ID            uint `gorm:"primaryKey"`
	MakerRoleID   uint
	MakerRole     Role `gorm:"foreignKey:MakerRoleID"`
	CheckerRoleID uint
	CheckerRole   Role `gorm:"foreignKey:CheckerRoleID"`
}

// Others
type CreateTransactionBody struct {
	MakerId string      `json:"maker_id"`
	Action  MakerAction `json:"action"`
}

type MakerAction struct {
	ActionType  string          `json:"action_type"`
	RequestBody json.RawMessage `json:"request_body"` // based off other function's request body
}

type UpdateTransactionRequestBody struct {
	CheckerId     string `json:"checker_id"`
	TransactionId string `json:"transaction_id"`
	Approval      bool   `json:"approval"`
}

type TransactionResponseBody struct {
	Txn Transaction
}

type MultipleTransactionsResponseBody struct {
	Txns []Transaction
}

type GetAllTransactionsRequestBody struct{}

type GetTransactionRequestBody struct {
	TransactionId string `json:"transaction_id"`
}
