package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"github.com/vynious/ascenda-lp-backend/db"
	"github.com/vynious/ascenda-lp-backend/types"
	"github.com/vynious/ascenda-lp-backend/util"
)

var (
	DBService *db.DBService
	DB *db.DB
)


func init() {
	var err error
	DBService, err = db.SpawnDBService()
	if err != nil {
		log.Fatalf("Failed to initialize database service: %v", err)
	}
}

func AuthorizerHandler(ctx context.Context, req events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {

	DB = DBService.GetBanksDB(req.Headers["Authorization"])

	token := req.Headers["Authorization"]
	method := req.HTTPMethod
	route := req.Path[1:]

	log.Printf("Authorizer %s %s", method, route)

	roleName, err := util.GetCustomAttributeWithCognito("custom:role", token)
	if err != nil || roleName == "" {
		return GenerateDenyPolicy(uuid.NewString(), req.MethodArn), nil
	}

	return GeneratePolicyBasedOnRole(ctx, roleName, uuid.NewString(), route, method, req.MethodArn), nil
}

func GeneratePolicyBasedOnRole(ctx context.Context, roleName, principalId, route, method, arn string) events.APIGatewayCustomAuthorizerResponse {
	log.Printf("GeneratePolicyBasedOnRole %s, %s", roleName, principalId)
	role, err := db.RetrieveRoleWithRoleName(ctx, DB, roleName)
	if err != nil {
		log.Printf("GenerateDenyPolicy %s", roleName)
		return GenerateDenyPolicy(principalId, arn)
	}

	permissions := role.Permissions
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}
	authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{Version: "2012-10-17", Statement: []events.IAMPolicyStatement{}}

	resource := determineResource(route)
	effect := "deny"
	log.Printf("resource: %v", resource)

	if resource == "maker_checker" {
		log.Printf("generating policy for maker-checker")
		effect := "allow"
		statement := generateStatement("execute-api:Invoke", effect, arn)
		authResponse.PolicyDocument.Statement = append(authResponse.PolicyDocument.Statement, statement)
	} else {
		for _, permission := range permissions {
			if permission.Resource == resource && checkPermission(permission, method) {
				log.Println("found resource and matching permissions")
				effect = "allow"
				statement := generateStatement("execute-api:Invoke", effect, arn)
				authResponse.PolicyDocument.Statement = append(authResponse.PolicyDocument.Statement, statement)
			}
		}
		// If no access permissions found, return deny policy
		if effect == "deny" {
			log.Println("resource not found or permission not allowed")
			return GenerateDenyPolicy(principalId, arn)
		}
	}

	log.Printf("%+v", authResponse)
	return authResponse
}

func determineResource(route string) string {
	switch route {
	case "user", "users", "role", "roles":
		return "user_storage"
	case "points":
		return "points_ledger"
	case "logs", "log":
		return "logs"
	case "maker-checker":
		return "maker_checker"
	default:
		return ""
	}
}

func checkPermission(permission types.RolePermission, method string) bool {
	switch method {
	case "GET":
		return permission.CanRead
	case "PUT":
		return permission.CanUpdate
	case "POST":
		return permission.CanCreate
	case "DELETE":
		return permission.CanDelete
	default:
		return false
	}
}

func generateStatement(action, effect, resource string) events.IAMPolicyStatement {
	return events.IAMPolicyStatement{
		Action:   []string{action},
		Effect:   effect,
		Resource: []string{resource},
	}
}

func GenerateDenyPolicy(principalId, arn string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalId,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Deny",
					Resource: []string{arn},
				},
			},
		},
	}
}

func main() {
	lambda.Start(AuthorizerHandler)
	defer DBService.CloseConnections()

}
