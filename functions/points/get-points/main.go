package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/vynious/ascenda-lp-backend/types"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vynious/ascenda-lp-backend/db"
)

var (
	DBService *db.DBService
	err       error
	headers   = map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
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
	var pointsRecords []types.Points

	DB := DBService.GetBanksDB(request.Headers["Authorization"])

	params := request.QueryStringParameters

	if params["user_id"] != "" {
		log.Printf("GetPointsAccountsByUser %s", params["user_id"])
		pointsRecords, err = DB.GetPointsAccountsByUser(ctx, params["user_id"])
	} else if params["id"] != "" {
		log.Printf("GetPointsAccountById %s", params["id"])
		pointsRecords, err = DB.GetPointsAccountById(ctx, params["id"])
	} else {
		log.Printf("GetPoints")
		pointsRecords, err = DB.GetPoints(ctx)
	}

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	if pointsRecords == nil {
		// Return 404 response if no points records are found
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	obj, err := json.Marshal(pointsRecords)
	if err != nil {
		log.Printf("Failed to parse points records: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       "Internal server error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(obj),
	}, nil
}
