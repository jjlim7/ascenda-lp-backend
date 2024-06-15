package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/vynious/ascenda-lp-backend/types"
	"gorm.io/gorm/clause"
)

// CreateTransaction creates a maker-checker transaction
func (dbs *DB) CreateTransaction(ctx context.Context, action types.MakerAction, makerId string) (*types.Transaction, error) {

	var maker types.User
	var approvalMap types.ApprovalChainMap

	tx := dbs.Conn.WithContext(ctx).Begin()

	// Ensure the MakerId corresponds to an existing user
	if err := tx.First(&maker, "id = ?", makerId).Error; err != nil {
		return nil, fmt.Errorf("maker with ID %s not found: %w", makerId, err)
	}

	// check if the maker-role-id is inside the approvalchain
	result := tx.Where("maker_role_id = ?", maker.RoleID).First(&approvalMap)
	if result.Error != nil {
		return nil, fmt.Errorf("user is not allowed to create a transaction: %v", result.Error.Error())
	}

	jsonMsgAction, _ := json.Marshal(action)

	txn := &types.Transaction{
		TransactionId: uuid.NewString(),
		MakerId:       makerId,
		Action:        jsonMsgAction,
	}

	if err := tx.Create(&txn).Error; err != nil {
		return nil, err
	}

	return txn, tx.Commit().Error
}

func (dbs *DB) GetTransaction(ctx context.Context, txnId string) (*[]types.Transaction, error) {
	var transaction types.Transaction

	tx := dbs.Conn.WithContext(ctx)

	if err := tx.Preload("Maker").Where("transaction_id = ?", txnId).First(&transaction).Error; err != nil {
		return nil, err
	}

	return &[]types.Transaction{transaction}, nil
}

func (dbs *DB) GetTransactions(ctx context.Context) (*[]types.Transaction, error) {
	var transactions []types.Transaction

	tx := dbs.Conn.WithContext(ctx)
	if result := tx.Preload("Maker").Find(&transactions); result.Error != nil {
		return nil, fmt.Errorf("failed to get all transactions: %v", result.Error.Error())
	}
	return &transactions, nil
}

// GetTransactionsByMakerIdByStatus Gets the transactions by maker_id and the status
func (dbs *DB) GetTransactionsByMakerIdByStatus(ctx context.Context, makerId string, status string) (*[]types.Transaction, error) {
	var transactions []types.Transaction

	tx := dbs.Conn.WithContext(ctx)
	if result := tx.
		Preload("Maker").
		Where("maker_id = ?", makerId).
		Where("status = ?", status).
		Find(&transactions); result.Error != nil {
		return nil, fmt.Errorf("failed to get all transactions by maker_id: %v", result.Error.Error())
	}
	return &transactions, nil
}

func (dbs *DB) GetPendingTransactionsForChecker(ctx context.Context, checkerId string) (*[]types.Transaction, error) {
	var transactions *[]types.Transaction
	var checkerRoleId uint

	// Start a transaction
	tx := dbs.Conn.WithContext(ctx)

	// Get the role ID of the checker
	err := tx.Table("users").
		Select("roles.id").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("users.id = ?", checkerId).
		Pluck("roles.id", &checkerRoleId).Error
	if err != nil {
		return nil, err
	}

	// Fetch transactions that are pending and where the maker's role is in the approval chain for the checker's role
	err = tx.Model(&types.Transaction{}).
		Preload("Maker").
		Joins("JOIN users AS makers ON makers.id = transactions.maker_id").
		Joins("JOIN roles AS maker_roles ON maker_roles.id = makers.role_id").
		Joins("JOIN approval_chain_maps ON approval_chain_maps.maker_role_id = maker_roles.id").
		Where("transactions.status = ?", "pending").
		Where("approval_chain_maps.checker_role_id = ?", checkerRoleId).
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetCompletedTransactionsByCheckerId This function assumes that all transactions with a value checker_id has been completed.
func (dbs *DB) GetCompletedTransactionsByCheckerId(ctx context.Context, checkerId string) (*[]types.Transaction, error) {
	var transactions []types.Transaction

	tx := dbs.Conn.WithContext(ctx)

	if result := tx.
		Preload("Maker").
		Preload("Checker").
		Where("checker_id = ?", checkerId).
		Find(&transactions); result.Error != nil {
		return nil, fmt.Errorf("failed to get completed transactions by checkerid: %v", result.Error.Error())
	}

	return &transactions, nil
}

func (dbs *DB) UpdateTransaction(ctx context.Context, txnId string, checkerId string, approval bool) (*types.Transaction, error) {

	tx := dbs.Conn.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Define the update map
	decision := map[string]interface{}{
		"CheckerId": checkerId,
		"Approval":  approval,
		"Status":    "completed",
	}

	// Update the transaction and locks the current entry
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&types.Transaction{}).
		Where("transaction_id = ?", txnId).
		Updates(decision).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Retrieve the updated transaction
	var updatedTransaction types.Transaction

	if err := tx.Where("transaction_id = ?", txnId).First(&updatedTransaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the transaction

	if approval {
		// if you cannot process the transaction, rollback updates for maker-checker transaction
		var makerAction types.MakerAction
		if err := json.Unmarshal(updatedTransaction.Action, &makerAction); err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := dbs.ProcessTransaction(ctx, &makerAction); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &updatedTransaction, nil
}

func (dbs *DB) GetCheckers(ctx context.Context, makerId string) ([]string, error) {
	var checkersEmails []string

	// First, get the maker's role name using the makerId
	var makerRoleName string
	err := dbs.Conn.WithContext(ctx).
		Table("users").
		Select("roles.role_name").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("users.id = ?", makerId).
		Pluck("roles.role_name", &makerRoleName).Error
	if err != nil {
		return nil, err
	}

	// Now, use the maker's role name to find the corresponding checkers' emails
	if err := dbs.Conn.WithContext(ctx).
		Table("users").
		Select("users.email").
		Joins("JOIN roles ON roles.id = users.role_id").
		Joins("JOIN approval_chain_maps ON approval_chain_maps.checker_role_id = roles.id").
		Joins("JOIN roles as maker_roles ON approval_chain_maps.maker_role_id = maker_roles.id").
		Where("maker_roles.role_name = ?", makerRoleName).
		Pluck("users.email", &checkersEmails).
		Error; err != nil {
		return nil, err
	}
	return checkersEmails, nil
}

func (dbs *DB) ProcessTransaction(ctx context.Context, action *types.MakerAction) error {
	switch action.ActionType {
	case "UpdatePoints":
		var updatePointsRequestBody types.UpdatePointsRequestBody
		if err := json.Unmarshal(action.RequestBody, &updatePointsRequestBody); err != nil {
			return err
		}
		if _, err := dbs.UpdatePoints(ctx, updatePointsRequestBody); err != nil {
			return err
		}
	case "UpdateUser":
		var updateUserRequestBody types.UpdateUserRequestBody
		if err := json.Unmarshal(action.RequestBody, &updateUserRequestBody); err != nil {
			return err
		}
		// calls update user.
		if _, err := UpdateUserWithUpdateUserRequestBody(ctx, dbs, updateUserRequestBody); err != nil {
			return err
		}
	case "":

	default:
		return fmt.Errorf("action method does not exist")
	}
	return nil
}
