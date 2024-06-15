package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/smithy-go"
	"github.com/vynious/ascenda-lp-backend/db"
	aws_helpers "github.com/vynious/ascenda-lp-backend/functions/users/aws-helpers"
	"github.com/vynious/ascenda-lp-backend/types"
	"gorm.io/gorm"
)

var (
	DBService     *db.DBService
	RDSClient     *rds.Client
	cognitoClient *cognito.CognitoIdentityProvider
	err           error
	DB *db.DB
)

func init() {
	DBService, err = db.SpawnDBService()
	if err != nil {
		log.Fatalf(err.Error())
	}
	cognitoClient = aws_helpers.InitCognitoClient()
}
func cognitoDeleteUser(userRequestBody types.DeleteUserRequestBody) error {
	cognitoInput := &cognito.AdminDeleteUserInput{
		Username:   aws.String(userRequestBody.Email),
		UserPoolId: aws.String(os.Getenv("COGNITO_USER_POOL_ID")),
	}

	_, err = cognitoClient.AdminDeleteUser(cognitoInput)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			fmt.Println("API Error Code:", apiErr.ErrorCode())
			fmt.Println("API Error Message:", apiErr.ErrorMessage())
		} else {
			fmt.Println("Unknown error:", err)
		}
		log.Println(err)
		return err
	}
	return nil
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	DB = DBService.GetBanksDB(request.Headers["Authorization"])

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

	var userRequestBody types.DeleteUserRequestBody

	if err := json.Unmarshal([]byte(request.Body), &userRequestBody); err != nil {
		log.Printf("JSON unmarshal error: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "DELETE",
			},
			Body: "Invalid request format",
		}, nil
	}

	err := cognitoDeleteUser(userRequestBody)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: "Error deleting user. Please check that the user exist/id is correct",
		}, nil
	}

	err = db.DeleteUserWithDeleteUserRequestBody(ctx, DB, userRequestBody)
	if err != nil {
		log.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "DELETE",
				},
				Body: "User not found",
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "DELETE",
			},
			Body: "Internal server error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "DELETE",
		},
		Body: "User successfully deleted",
	}, nil
}

func main() {
	lambda.Start(handler)
	defer DBService.CloseConnections()
}
