package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		"Access-Control-Allow-Methods": "DELETE",
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

	req := types.DeletePointsAccountRequestBody{}
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       errors.New("invalid request. malformed request found").Error(),
		}, nil
	}

	if req.ID == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       errors.New("bad request. id not found").Error(),
		}, nil
	}
	log.Printf("UpdatePoints %s", *req.ID)

	deleted, err := DB.DeletePointsAccountByID(ctx, *req.ID)
	if !deleted {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	resp := fmt.Sprintf("Points account %s successfully deleted", *req.ID)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       resp,
	}, nil
}
