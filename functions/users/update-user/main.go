package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go/aws"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
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

func cognitoUpdateUser(userRequestBody types.UpdateUserRequestBody) error {
	var userAttributes []*cognito.AttributeType

	if userRequestBody.NewFirstName != "" {
		userAttributes = append(userAttributes, &cognito.AttributeType{Name: aws.String("given_name"), Value: aws.String(userRequestBody.NewFirstName)})
	}
	if userRequestBody.NewLastName != "" {
		userAttributes = append(userAttributes, &cognito.AttributeType{Name: aws.String("family_name"), Value: aws.String(userRequestBody.NewLastName)})
	}
	if userRequestBody.NewRoleName != "" {
		userAttributes = append(userAttributes, &cognito.AttributeType{Name: aws.String("custom:role"), Value: aws.String(userRequestBody.NewRoleName)})
	}

	cognitoInput := &cognito.AdminUpdateUserAttributesInput{
		UserPoolId:     aws.String(os.Getenv("COGNITO_USER_POOL_ID")),
		Username:       aws.String(userRequestBody.Email),
		UserAttributes: userAttributes,
	}

	_, err = cognitoClient.AdminUpdateUserAttributes(cognitoInput)
	if err != nil {
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

	var userRequestBody types.UpdateUserRequestBody

	if err := json.Unmarshal([]byte(request.Body), &userRequestBody); err != nil {
		log.Printf("JSON unmarshal error: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "PUT",
			},
			Body: "Invalid request format",
		}, nil
	}

	err := cognitoUpdateUser(userRequestBody)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: "Error updating user",
		}, nil
	}

	updatedUser, err := db.UpdateUserWithUpdateUserRequestBody(ctx, DB, userRequestBody)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "PUT",
				},
				Body: "User not found",
			}, nil
		} else {
			log.Printf("Database error: %s", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers: map[string]string{
					"Access-Control-Allow-Headers": "Content-Type",
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "PUT",
				},
				Body: "Internal server error",
			}, nil
		}
	}

	responseBody, err := json.Marshal(updatedUser)
	if err != nil {
		log.Printf("JSON marshal error: %s", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "PUT",
			},
			Body: "Error marshaling user into JSON",
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "PUT",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	// we are simulating a lambda behind an ApiGatewayV2
	lambda.Start(handler)
	defer DBService.CloseConnections()
}
