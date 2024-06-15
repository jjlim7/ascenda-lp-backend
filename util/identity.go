package util

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

func GetRoleWithCognito(token string) (string, error) {
	
	var roleName string

	sess := session.Must(session.NewSession())
	cogClient := cognitoidentityprovider.New(sess, aws.NewConfig().WithRegion("ap-southeast-1"))

	result, err := cogClient.GetUser(&cognitoidentityprovider.GetUserInput{
		AccessToken: &token,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get user from token %v", err)
	}

	for i := 0; i < len(result.UserAttributes); i++ {
		if *result.UserAttributes[i].Name == "custom:role" {
			roleName = *result.UserAttributes[i].Value
			break
		}
	}
	
	return roleName, nil
}

func GetCustomAttributeWithCognito(attribute, token string) (string, error) {
	

	var res string 

	sess := session.Must(session.NewSession())
	cogClient := cognitoidentityprovider.New(sess, aws.NewConfig().WithRegion("ap-southeast-1"))

	result, err := cogClient.GetUser(&cognitoidentityprovider.GetUserInput{
		AccessToken: &token,
	})

	if err != nil {
		return "", fmt.Errorf("failed to get user from token %v", err)
	}

	for i := 0; i < len(result.UserAttributes); i++ {
		if *result.UserAttributes[i].Name == attribute {
			res = *result.UserAttributes[i].Value
			break
		}
	}


	return res, nil
}