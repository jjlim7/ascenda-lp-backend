package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vynious/ascenda-lp-backend/db"
	"github.com/vynious/ascenda-lp-backend/types"
)

var (
	DBService *db.DBService
	err       error
	headers   = map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "POST",
	}
)

func init() {
	log.Printf("INIT")
	DBService, err = db.SpawnDBService()
}

func main() {
	// we are simulating a lambda behind an ApiGatewayV2
	lambda.Start(handler)

	defer DBService.CloseConnections()
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	DB := DBService.GetBanksDB(request.Headers["Authorization"])

	req := types.CreatePointsAccountRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       errors.New("invalid request. malformed request found").Error(),
		}, nil
	}

	if req.UserID == nil || req.NewBalance == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       errors.New("bad request. user_id or new_balance not found").Error(),
		}, nil
	}
	log.Printf("CreatePointsAccount %s", *req.UserID)

	pointsRecord, err := DB.CreatePointsAccount(ctx, req)
	if pointsRecord == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	obj, _ := json.Marshal(pointsRecord)
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers:    headers,
		Body:       string(obj),
	}, nil
}
