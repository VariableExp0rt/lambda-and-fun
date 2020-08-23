package main

import (
	"github.com/VariableExp0rt/lambda-and-fun/cmd"
)

// TODO: REWRITE THIS SECTION TO ONLY INCLUDE NESTED SWITCH STATEMENTS
// CASE "CREATE", NESTED SWITCH FOR SUBRESOURCES (LAMBDA, ROLES, GATEWAY) OR ALL
// CASE "DELETE", NESTED SWITCH FOR SUBRESOURCES (LAMBDA, ROLES, GATEWAY) OR ALL
// This will be used to handle command line subcommands
func main() {
	cmd.Execute()

	// TODO: Move this to a function in helper package
	//testResult, err := apigwsvc.TestInvokeMethod(&apigateway.TestInvokeMethodInput{
	//	Body:                aws.String(`{"operation": "create"}`),
	//	HttpMethod:          aws.String("POST"),
	//	PathWithQueryString: aws.String(""),
	//	ResourceId:          resourceID,
	//	RestApiId:           api,
	//})
	//if err != nil || testResult.Status != aws.Int64(200) {
	//	fmt.Println(err.Error(), testResult.Status)
	//}
}
