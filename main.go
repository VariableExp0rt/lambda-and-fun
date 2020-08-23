package main

import (
	"fmt"
	"time"

	"github.com/VariableExp0rt/lambda-and-fun/cmd"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/lambda"
	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/service/apigateway"
	//"github.com/aws/aws-sdk-go/service/lambda"
)

// TODO: REWRITE THIS SECTION TO ONLY INCLUDE NESTED SWITCH STATEMENTS
// CASE "CREATE", NESTED SWITCH FOR SUBRESOURCES (LAMBDA, ROLES, GATEWAY) OR ALL
// CASE "DELETE", NESTED SWITCH FOR SUBRESOURCES (LAMBDA, ROLES, GATEWAY) OR ALL
// This will be used to handle command line subcommands
func main() {
	cmd.Execute()

	resource, err := apigwsvc.CreateResource(&apigateway.CreateResourceInput{
		RestApiId: api,
		PathPart:  resp3.Name,
		ParentId:  parentID,
	})
	if err != nil {
		fmt.Println(err.Error(), resource, resp3)
	}

	time.Sleep(6 * time.Second)

	resourceID := resource.Id

	//PutMethod so that we are able to sent POST requests to the API Gateway to trigger
	//the Lambda function
	method, err := apigwsvc.PutMethod(&apigateway.PutMethodInput{
		AuthorizationType: aws.String("None"),
		HttpMethod:        aws.String("POST"),
		RestApiId:         api,
		ResourceId:        resourceID,
	})
	if err != nil {
		fmt.Println(err.Error(), method, resource)
	}

	//Enable the integration responsible for allowing the API Gateway to forward requests/trigger
	//the Lambda function
	fmt.Println(aws.StringValue(function.FunctionArn))
	integration, err := apigwsvc.PutIntegration(&apigateway.PutIntegrationInput{
		HttpMethod:            aws.String("POST"),
		IntegrationHttpMethod: aws.String("POST"),
		RestApiId:             api,
		ResourceId:            resourceID,
		Type:                  aws.String("AWS"),
		Uri:                   aws.String("arn:aws:apigateway:" + aws.StringValue(region) + ":lambda:path/2015-03-31/functions/" + aws.StringValue(function.FunctionArn) + "/invocations"),
	})
	if err != nil {
		fmt.Println(err.Error(), integration, err)
	}

	//There is absolutely, surely, an easier way to get a pointer to a string than doing this

	var str = "Empty"

	respModel := make(map[string]*string, 1)
	respModel["application/json"] = &str

	methodResp, err := apigwsvc.PutMethodResponse(&apigateway.PutMethodResponseInput{
		HttpMethod:     aws.String("POST"),
		RestApiId:      api,
		ResourceId:     resourceID,
		ResponseModels: respModel,
		StatusCode:     aws.String("200"),
	})
	if err != nil {
		fmt.Println(err.Error(), methodResp)
	}

	str = ""
	respModel["application/json"] = &str

	intResp, err := apigwsvc.PutIntegrationResponse(&apigateway.PutIntegrationResponseInput{
		HttpMethod:        aws.String("POST"),
		RestApiId:         api,
		ResourceId:        resourceID,
		ResponseTemplates: respModel,
		StatusCode:        aws.String("200"),
	})
	if err != nil {
		fmt.Printf("Error putting integration response %v", intResp)
	}

	deploy, err := apigwsvc.CreateDeployment(&apigateway.CreateDeploymentInput{
		RestApiId: api,
		StageName: aws.String("prod"),
	})
	if err != nil {
		fmt.Println(err.Error(), deploy, err)
	}

	time.Sleep(6 * time.Second)

	// Could really do with wrapping this into a variadic function because it looks untidy
	var (
		pathPrefix = "arn:aws:execute-api:"
		pathSuffix = "/*/POST/EKS-Setup"
	)

	//Anonymous helper function to concat strings that are a string pointer
	//var srcArn = func(strings ...*string) string {
	//	var s string
	//	for _, str := range strings {
	//		s = s + aws.StringValue(str)
	//		return s
	//	}
	//	return s
	//}()

	perms, err := lmdsvc.AddPermission(&lambda.AddPermissionInput{
		FunctionName: &FuncName,
		StatementId:  aws.String("apigateway-test-2"),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("apigateway.amazonaws.com"),
		SourceArn:    aws.String(pathPrefix + aws.StringValue(region) + ":" + aws.StringValue(account) + ":" + aws.StringValue(api) + pathSuffix),
	})
	if err != nil {
		fmt.Println(err.Error(), perms.Statement)
	}

	perms1, err := lmdsvc.AddPermission(&lambda.AddPermissionInput{
		FunctionName: &FuncName,
		StatementId:  aws.String("apigateway-prod-2"),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("apigateway.amazonaws.com"),
		SourceArn:    aws.String(pathPrefix + aws.StringValue(region) + ":" + aws.StringValue(account) + ":" + aws.StringValue(api) + pathSuffix),
	})
	if err != nil {
		fmt.Println(err.Error(), perms1.Statement)
	}

	testResult, err := apigwsvc.TestInvokeMethod(&apigateway.TestInvokeMethodInput{
		Body:                aws.String(`{"operation": "create"}`),
		HttpMethod:          aws.String("POST"),
		PathWithQueryString: aws.String(""),
		ResourceId:          resourceID,
		RestApiId:           api,
	})
	if err != nil || testResult.Status != aws.Int64(200) {
		fmt.Println(err.Error(), testResult.Status)
	}

	fmt.Println("Successfully deployed API Gateway and tested with the following status code: ", testResult.Status)
}
