package main

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/vynious/ascenda-lp-backend/types"
)

var svc *dynamodb.DynamoDB

func init() {
	log.Printf("INIT")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-southeast-1"),
		Credentials: credentials.NewStaticCredentials("", "", ""),
	})
	if err != nil {
		log.Printf("failed to connect to db: " + err.Error())
		print(err)
	}

	svc = dynamodb.New(sess)
}

func handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	const logsTable = "logs"

	ttlStr := request.QueryStringParameters["TTL"]
	if ttlStr != "" {
		ttl, err := strconv.Atoi(ttlStr)
		if err != nil {
			return events.APIGatewayProxyResponse{}, errors.New("invalid TTL value")
		}

		return setTTL(ttl, logsTable)
	}

	// Check if ID is provided in query parameters
	// id := request.QueryStringParameters["id"]
	// if len(id) > 0 {
	// 	return fetchLogByID(id, logsTable)
	// }

	// Fetch all logs if no ID provided
	return fetchLogs(request, logsTable)
}

// func fetchLogByID(id, tableName string) (events.APIGatewayProxyResponse, error) {
// 	// Get single log from DynamoDB
// 	input := &dynamodb.GetItemInput{
// 		Key: map[string]*dynamodb.AttributeValue{
// 			"log_id": {
// 				S: aws.String(id),
// 			},
// 		},
// 		TableName: aws.String(tableName),
// 	}

// 	result, err := svc.GetItem(input)
// 	if err != nil {
// 		return events.APIGatewayProxyResponse{}, errors.New("failed to fetch record")
// 	}

// 	if result.Item == nil {
// 		return events.APIGatewayProxyResponse{
// 			StatusCode: 404,
// 			Body:       "Log does not exist",
// 		}, nil
// 	}

// 	item := new(types.Log)
// 	err = dynamodbattribute.UnmarshalMap(result.Item, item)
// 	if err != nil {
// 		return events.APIGatewayProxyResponse{}, errors.New("failed to unmarshal record")
// 	}

// 	// Convert log item to JSON
// 	body, _ := json.Marshal(item)

// 	// Return JSON response
// 	return events.APIGatewayProxyResponse{
// 		Body:       string(body),
// 		StatusCode: 200,
// 		Headers: map[string]string{
// 			"Access-Control-Allow-Headers": "Content-Type",
// 			"Access-Control-Allow-Origin":  "*",
// 			"Access-Control-Allow-Methods": "GET",
// 		},
// 	}, nil
// }

func fetchLogs(request events.APIGatewayV2HTTPRequest, tableName string) (events.APIGatewayProxyResponse, error) {
	log.Printf("fetching logs")
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := svc.Scan(input)
	if err != nil {
		log.Printf("failed to fetch logs: " + err.Error())
		return events.APIGatewayProxyResponse{}, errors.New("failed to fetch record")
	}

	var logs []types.Log
	for _, i := range result.Items {
		logItem := new(types.Log)
		err := dynamodbattribute.UnmarshalMap(i, logItem)
		if err != nil {
			log.Printf("failed to unmarshal logs: " + err.Error())
			return events.APIGatewayProxyResponse{}, err
		}
		logItem.LogId = *i["log_id"].S // Assuming log_id is the attribute name in DynamoDB
		logs = append(logs, *logItem)
	}

	responseBody, err := json.Marshal(logs)
	if err != nil {
		log.Printf("Failed to parse logs: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET",
			},
			Body: "Internal server error, failed to execute",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(responseBody),
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET",
		},
	}, nil
}

func setTTL(ttl int, tableName string) (events.APIGatewayProxyResponse, error) {
	// Check if TTL is already enabled
	descOutput, err := svc.DescribeTimeToLive(&dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		log.Printf("failed to describe TTL: %s", err.Error())
		return events.APIGatewayProxyResponse{}, errors.New("failed to describe TTL")
	}

	// If TTL is already enabled, return an error
	if descOutput.TimeToLiveDescription.TimeToLiveStatus != nil &&
		*descOutput.TimeToLiveDescription.TimeToLiveStatus == "ENABLED" {
		log.Printf("TTL is already enabled for table %s", tableName)
		return events.APIGatewayProxyResponse{}, errors.New("TTL is already enabled")
	}

	input := &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(tableName),
		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
			AttributeName: aws.String("ttl"),
			Enabled:       aws.Bool(true),
		},
	}

	_, err = svc.UpdateTimeToLive(input)
	if err != nil {
		log.Printf("failed to update TTL: %s", err.Error())
		return events.APIGatewayProxyResponse{}, errors.New("failed to update TTL")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "TTL set successfully",
	}, nil
}

func main() {
	lambda.Start(handler)
}
