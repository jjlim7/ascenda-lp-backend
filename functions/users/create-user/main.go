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
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/google/uuid"
	"github.com/vynious/ascenda-lp-backend/db"
	aws_helpers "github.com/vynious/ascenda-lp-backend/functions/users/aws-helpers"
	"github.com/vynious/ascenda-lp-backend/types"
	"github.com/vynious/ascenda-lp-backend/util"
	"gorm.io/gorm"
)

var (
	DBService     *db.DBService
	RDSClient     *rds.Client
	cognitoClient *cognitoidentityprovider.CognitoIdentityProvider
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

func cognitoCreateUser(userRequestBody types.CreateUserRequestBody, newUUID, bankName string) error {
	userPoolID := os.Getenv("COGNITO_USER_POOL_ID")
	log.Println(userPoolID)
	cognitoInput := &cognitoidentityprovider.AdminCreateUserInput{
		ForceAliasCreation: aws.Bool(true),
		DesiredDeliveryMediums: []*string{
			aws.String("EMAIL"),
		},
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(userRequestBody.Email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
			{
				Name:  aws.String("custom:userID"),
				Value: aws.String(newUUID),
			},
			{
				Name:  aws.String("custom:role"),
				Value: aws.String(userRequestBody.RoleName),
			},
			{
				Name: aws.String("custom:bank"),
				Value: aws.String(bankName),
			},
		},
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(userRequestBody.Email),
	}
	log.Println("try create user in user pool")
	_, err := cognitoClient.AdminCreateUser(cognitoInput)
	log.Println("tried")
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("User created in user pool")
	return nil
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {



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

	var userRequestBody types.CreateUserRequestBody


	DB = DBService.GetBanksDB(request.Headers["Authorization"])

	bankName, err := util.GetCustomAttributeWithCognito("custom:bank", request.Headers["Authorization"])

	if err != nil {
		log.Printf("failed to get user attribute: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: "Invalid request format",
		}, nil
	}

	if err := json.Unmarshal([]byte(request.Body), &userRequestBody); err != nil {
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

	if !util.CheckEmailValidity(userRequestBody.Email) {
		log.Printf("Invalid email")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST",
			},
			Body: "Invalid email",
		}, nil
	}

	_, err = db.RetrieveUserWithEmail(ctx, DB, userRequestBody.Email)
	if err != nil {
		log.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newUUID := uuid.NewString()
			err := cognitoCreateUser(userRequestBody, newUUID, bankName)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: 404,
					Headers: map[string]string{
						"Access-Control-Allow-Headers": "Content-Type",
						"Access-Control-Allow-Origin":  "*",
						"Access-Control-Allow-Methods": "POST",
					},
					Body: "Error creating user",
				}, nil
			}
			user, err := db.CreateUserWithCreateUserRequestBody(ctx, DB, userRequestBody, newUUID)
			if err != nil {
				log.Printf("Database error: %s", err)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return events.APIGatewayProxyResponse{
						StatusCode: 404,
						Headers: map[string]string{
							"Access-Control-Allow-Headers": "Content-Type",
							"Access-Control-Allow-Origin":  "*",
							"Access-Control-Allow-Methods": "POST",
						},
						Body: "Role not found. Please create a role first (if user have a role)",
					}, nil
				}
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

			// Send email to new users to verify their email to receive notifications.
			if err := util.SendEmailVerification(ctx, user.Email); err != nil {
				log.Printf("failed to send email: %v", err)
			}

			responseBody := fmt.Sprintf("{\"email\": \"%s\", \"id\": \"%s\"}", user.Email, user.Id)
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
		Body: "User already exist. Please use a different email",
	}, nil
}

func main() {
	lambda.Start(handler)
	defer DBService.CloseConnections()
}
