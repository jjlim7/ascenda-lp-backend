package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vynious/ascenda-lp-backend/db"
	"github.com/vynious/ascenda-lp-backend/types"
)

var (
	DBService *db.DBService
	DB *db.DB
	err       error
	headers   = map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
	}
)

func init() {

	// Initialise global variable DBService tied to Aurora
	DBService, err = db.SpawnDBService()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func GetTransactionsHandler(ctx context.Context, req *events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	
	var transactions *[]types.Transaction

	params := req.QueryStringParameters

	DB = DBService.GetBanksDB(req.Headers["Authorization"])


	switch {
	case params["transaction_id"] != "":
		// get one
		transactions, err = DB.GetTransaction(ctx, params["transaction_id"])

	// Checker
	case params["checker_id"] != "" && params["status"] != "":
		if params["status"] == "pending" {
			transactions, err = DB.GetPendingTransactionsForChecker(ctx, params["checker_id"])
		} else if params["status"] == "completed" {
			transactions, err = DB.GetCompletedTransactionsByCheckerId(ctx, params["checker_id"])
		} else {
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Headers:    headers,
				Body:       `{"message":"wtf is this"}`,
			}, nil
		}

	// Maker
	case params["maker_id"] != "" && params["status"] != "":
		if params["status"] == "pending" {
			transactions, err = DB.GetTransactionsByMakerIdByStatus(ctx, params["maker_id"], params["status"])
		} else if params["status"] == "completed" {
			transactions, err = DB.GetTransactionsByMakerIdByStatus(ctx, params["maker_id"], params["status"])
		} else {
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Headers:    headers,
				Body:       `{"message":"wtf is this"}`,
			}, nil
		}
	case len(params) == 0:
		transactions, err = DB.GetTransactions(ctx)
	default:
		// get all
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       `{"message":"wtf is this"}`,
		}, nil
	}

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       err.Error(),
		}, nil

	}

	if transactions == nil {
		// Return 404 response if no points records are found
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	obj, err := json.Marshal(transactions)
	if err != nil {
		log.Printf("Failed to parse transactions records: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       `{"message":"internal service error"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(obj),
	}, nil

}

func main() {
	lambda.Start(GetTransactionsHandler)
	defer DBService.CloseConnections()

}
