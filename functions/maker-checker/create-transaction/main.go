package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vynious/ascenda-lp-backend/db"
	"github.com/vynious/ascenda-lp-backend/types"
	"github.com/vynious/ascenda-lp-backend/util"
)

var (
	DBService    *db.DBService
	DB *db.DB
	requestBody  types.CreateTransactionBody
	responseBody types.TransactionResponseBody
	err          error
)

func init() {
	// Initialise global variable DBService tied to Aurora
	DBService, err = db.SpawnDBService()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func CreateTransactionHandler(ctx context.Context, req *events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {

	DB = DBService.GetBanksDB(req.Headers["Authorization"])


	if err := json.Unmarshal([]byte(req.Body), &requestBody); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: "Bad Request",
		}, nil
	}

	makerId := requestBody.MakerId

	switch requestBody.Action.ActionType {
	case "UpdatePoints":

		// to check if the request body matches UpdatePointsRequesBody struct
		var updatePointsRequestBody types.UpdatePointsRequestBody
		if err := json.Unmarshal(requestBody.Action.RequestBody, &updatePointsRequestBody); err != nil {
			log.Printf("Error unmarshalling UpdatePointsRequestBody: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "POST",
				},
				Body: `{"message": "Invalid request format for UpdatePoints"}`,
			}, nil
		}

		log.Printf("UpdatePointsRequestBody: %+v", updatePointsRequestBody)

		// convert to json.RawMessage to fit MakerAction struct
		rawJsonBody, _ := json.Marshal(updatePointsRequestBody)

		// recreate the MakerAction struct to store
		updatedMakerCheckerAction := types.MakerAction{
			ActionType:  requestBody.Action.ActionType,
			RequestBody: rawJsonBody,
		}

		txn, err := DB.CreateTransaction(ctx, updatedMakerCheckerAction, makerId)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "POST",
				},
				Body: "",
			}, nil
		}
		responseBody.Txn = *txn
	case "UpdateUser":
		var updateUserRequestBody types.UpdateUserRequestBody
		if err := json.Unmarshal(requestBody.Action.RequestBody, &updateUserRequestBody); err != nil {
			log.Printf("Error unmarshalling UpdateUserRequestBody: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "POST",
				},
				Body: `{"message": "Invalid request format for UpdateUser"}`,
			}, nil
		}

		log.Printf("UpdatePointsRequestBody: %+v", updateUserRequestBody)

		// convert to json.RawMessage to fit MakerAction struct
		rawJsonBody, _ := json.Marshal(updateUserRequestBody)

		updatedMakerCheckerAction := types.MakerAction{
			ActionType:  requestBody.Action.ActionType,
			RequestBody: rawJsonBody,
		}

		txn, err := DB.CreateTransaction(ctx, updatedMakerCheckerAction, makerId)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "POST",
				},
				Body: "",
			}, nil
		}
		responseBody.Txn = *txn

	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: `{"message": "Bad Request"}`,
		}, nil
	}

	// Send emails seek checker's approval (Async)
	log.Printf("starting to send email...")
	checkersEmail, err := DB.GetCheckers(ctx, makerId)
	if err != nil {
		log.Println(err.Error())
	}
	if err = util.EmailCheckers(ctx, requestBody.Action.ActionType,
		checkersEmail); err != nil {
		log.Println(err.Error())
	}

	respBod, err := json.Marshal(responseBody)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 201,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: err.Error(),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST",
		},
		Body: string(respBod),
	}, nil
}

func main() {
	lambda.Start(CreateTransactionHandler)
	defer DBService.CloseConnections()
}
