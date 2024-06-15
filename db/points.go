package db

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/vynious/ascenda-lp-backend/types"
	"github.com/vynious/ascenda-lp-backend/util"
)

func (dbs *DB) GetPoints(ctx context.Context) ([]types.Points, error) {
	var pointsRecords []types.Points
	res := dbs.Conn.Find(&pointsRecords)
	if res.Error != nil {
		logEntry := types.Log{
			Type:   "Points",
			Action: fmt.Sprintf("Failed to query points: %s", res.Error),
		}
		if err := util.CreateLogEntry(logEntry); err != nil {
			log.Printf("Error creating log entry: %v", err)
		}
		return nil, fmt.Errorf("database error: %s", res.Error)
	}

	logEntry := types.Log{
		Type:   "Points",
		Action: "Queried all points",
	}
	if err := util.CreateLogEntry(logEntry); err != nil {
		log.Printf("Error creating log entry: %v", err)
	}

	return pointsRecords, nil
}

func (dbs *DB) GetPointsAccountById(ctx context.Context, accId string) ([]types.Points, error) {

	var pointsRecords []types.Points
	res := dbs.Conn.Where("id = ?", accId).First(&pointsRecords)
	if res.Error != nil {
		return nil, fmt.Errorf("database error %s", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("account %s does not exist", accId)
	}

	return pointsRecords, nil
}

func (dbs *DB) GetPointsAccountsByUser(ctx context.Context, userId string) ([]types.Points, error) {

	log.Printf("Test %s", userId)
	var pointsRecords []types.Points
	res := dbs.Conn.Where("user_id = ?", userId).Find(&pointsRecords)
	if res.Error != nil {
		return nil, fmt.Errorf("database error %s", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("user %s does not exist", userId)
	}

	return pointsRecords, nil
}

func (dbs *DB) UpdatePoints(ctx context.Context, req types.UpdatePointsRequestBody) (*types.Points, error) {

	var pointsRecords []types.Points
	pointsRecords, err := dbs.GetPointsAccountById(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	res := dbs.Conn.Model(pointsRecords).Update("balance", req.NewBalance).First(&pointsRecords)
	if res.RowsAffected == 0 {
		return nil, res.Error
	}

	return &pointsRecords[0], nil
}

func (dbs *DB) CreatePointsAccount(ctx context.Context, req types.CreatePointsAccountRequestBody) (*types.Points, error) {

	pointsRecord := types.Points{
		ID:      uuid.NewString(),
		UserID:  *req.UserID,
		Balance: *req.NewBalance,
	}
	res := dbs.Conn.Model(types.Points{}).Create(&pointsRecord)
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("database error %s", res.Error)
	}

	return &pointsRecord, nil
}

func (dbs *DB) DeletePointsAccountByUser(ctx context.Context, userId string) (bool, error) {
	res := dbs.Conn.Where("user_id = ?", &userId).Delete(&types.Points{})
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("database error %s", res.Error)
	}

	return true, nil
}

func (dbs *DB) DeletePointsAccountByID(ctx context.Context, accId string) (bool, error) {
	res := dbs.Conn.Delete(&types.Points{}, &accId)
	if res.RowsAffected == 0 {
		return false, fmt.Errorf("database error %v", res.Error)
	}

	return true, nil
}
