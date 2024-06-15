package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/vynious/ascenda-lp-backend/db"
	"github.com/vynious/ascenda-lp-backend/types"
	"gorm.io/gorm"
)

var (
	DBService *db.DBService
	RDSClient *rds.Client
	err       error
)

func init() {
	DBService, err = db.SpawnDBService()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	DB := DBService.GetBanksDB(request.Headers["Authorization"])

	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type",
			},
			Body: "{}",
		}, nil
	}

	var roleRequestBody types.CreateRoleRequestBody

	if err := json.Unmarshal([]byte(request.Body), &roleRequestBody); err != nil {
		log.Printf("JSON unmarshal error: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: "Invalid request format",
		}, nil
	}

	_, err := db.RetrieveRoleWithRoleName(ctx, DB, roleRequestBody.RoleName)
	if err != nil {
		log.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			roleName, err := db.CreateRoleWithCreateRoleRequestBody(ctx, DB, roleRequestBody)
			if err != nil {
				log.Printf("Database error: %s", err)
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
					Headers: map[string]string{
						"Access-Control-Allow-Headers": "Content-Type",
						"Access-Control-Allow-Origin":  "*",
						"Access-Control-Allow-Methods": "POST",
					},
					Body: "Internal server error",
				}, nil
			}

			responseBody := fmt.Sprintf("{\"role_name\": \"%s\"}", roleName)
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "POST",
				},
				Body: responseBody,
			}, nil
		} else {
			log.Printf("Database error: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "POST",
				},
				Body: "Internal server error",
			}, nil
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 409,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST",
		},
		Body: "Role already exist. Please use a different name",
	}, nil
}

func main() {
	// we are simulating a lambda behind an ApiGatewayV2
	lambda.Start(handler)
	defer DBService.CloseConnections()
}
