package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var (
	fileVar string
	// TODO: Could have the below region as an argument or flag really
	region  = func(s string) *string { return &s }("eu-west-2")
	account *string
)

func main() {
	// arguments to be passed to the commandline when invoking the program
	var RoleName string

	flag.StringVar(&fileVar, "file-path", "", "Path to deployment package (zip file)")
	flag.StringVar(&RoleName, "role-name", "", "Name of the new role to create")
	flag.Bool("skip-role-creation", false, "Use this flag to skip role creation if you have an existing role")
	//flag.StringVar(&runtime, "runtime", "go1.x", "Specify runtime for Lambda, default is go1.x.")
	flag.Parse()

	//TODO: also a switch that exits if required flag is missing

	//Add this to remove some of the globally declared variables
	if envVar, res := os.LookupEnv("ACCOUNT"); res != false {
		account = &envVar
	} else if res == false {
		fmt.Println("must set account ID environment variable before proceeding")
		os.Exit(1)
	}

	//TODO: Change this, add a CLI flag and if not null string, execute the create role, otherwise, add flag for attach role
	if len(RoleName) == 0 {
		RoleName = "default-role" + string(rand.Int())
	}
	// Every interaction with the AWS SDK requires a session, this just establishes a basic session
	// to be able to pass to the services we want to interact with
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, Config: aws.Config{Region: aws.String("eu-west-2")}}))

	// Initialise the service we'd like to work with
	svc := iam.New(sess)

	//Create the config to be used below to CreateRole
	params := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String("{\"Version\": \"2012-10-17\",\"Statement\": [{\"Effect\": \"Allow\",\"Principal\": {\"Service\": \"lambda.amazonaws.com\"},\"Action\": \"sts:AssumeRole\"}]}"),
		Description:              aws.String("A Role to allow Lambda to perform it's basic functions and interact with CloudFormation"),
		RoleName:                 aws.String(RoleName),
	}

	//CreateRole will create a new managed IAM Role, and can only have one managed policy attached to
	//it in the above configuration. Multiple can be assigned by calling AttachRolePolicy as many
	//times as needed

	//TODO: switch statement using above skip role creation var as case
	resp, err := svc.CreateRole(params)
	if err != nil {
		fmt.Println(err.Error(), resp)
	}

	//Wait until the Role exists before attaching new policies, which would otherwise fail
	svc.WaitUntilRoleExists(&iam.GetRoleInput{
		RoleName: aws.String(RoleName),
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//This is the primary policy we want the Lambda to be able to use when 'Assuming' the Role
	res, err := svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AdministratorAccess"),
		RoleName:  aws.String(RoleName),
	})
	if err != nil {
		fmt.Println(err.Error(), res)
	}
	fmt.Println("Successfully attached policy to Role", res)

	time.Sleep(6 * time.Second)

	//TODO: Create the policy document, then attach it to the Lambda role created above

	// This is easier than putting into a function, as I had before!
	pkg, err := ioutil.ReadFile(fileVar)
	if err != nil {
		fmt.Println(err.Error(), "Unable to read deployment file:", fileVar)
		return
	}

	lmdsvc := lambda.New(sess)

	var FuncName = "my-example-function"

	//Generate the Function Input
	funcParams := &lambda.CreateFunctionInput{
		Code:         &lambda.FunctionCode{ZipFile: pkg},
		Description:  aws.String("This function manages the creation and deletion of Amazon EKS Clusters"),
		FunctionName: aws.String(FuncName),
		Handler:      aws.String("main"),
		Role:         resp.Role.Arn,
		Runtime:      aws.String("go1.x"),
	}

	//Create the function and wait until it exists before trying to create the API Gateway resource for it
	resp2, err := lmdsvc.CreateFunction(funcParams)
	if err != nil {
		fmt.Println(err.Error(), resp2)
		return
	}

	if err := lmdsvc.WaitUntilFunctionActive(&lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(FuncName),
	}); err != nil {
		fmt.Println(err.Error(), resp2)
		return
	}

	fmt.Println("Function was created successfully and is currently active: ", FuncName)

	time.Sleep(6 * time.Second)

	// A helper for referencing within the API Gateway PutIntegration below
	function, err := lmdsvc.GetFunctionConfiguration(&lambda.GetFunctionConfigurationInput{FunctionName: aws.String(FuncName)})

	//Initialise another service for the API Gateway
	apigwsvc := apigateway.New(sess)

	//Create the REST API named EKS-Setup
	resp3, err := apigwsvc.CreateRestApi(&apigateway.CreateRestApiInput{
		Name: aws.String("EKS-Setup"),
	})
	if err != nil {
		fmt.Println(err.Error(), resp3)
	}

	time.Sleep(6 * time.Second)

	//Grab the ID of the newly created REST API, this is needed below to find the root object id
	api := resp3.Id

	//add logic to get the root path or parent ID
	parentAPI, err := apigwsvc.GetResources(&apigateway.GetResourcesInput{RestApiId: api})
	if err != nil {
		fmt.Println(err.Error(), parentAPI)
	}

	var parentID = func(*apigateway.GetResourcesOutput) *string {
		var s *string
		for _, val := range parentAPI.Items {
			s = val.Id
			if aws.StringValue(val.Id) == "" {
				os.Exit(1)
			}
		}
		return s
	}(parentAPI)

	//var check = func(parentID *string) bool {
	//	var ns string
	//	ns = aws.StringValue(parentID)
	//	var chk bool
	//	if ns == "" {
	//		chk = false
	//	} else if ns != "" {
	//		chk = true
	//	}
	//	return chk
	//}(parentID)

	//if check == false {
	//	fmt.Println("Cannot get parent ID value")
	//	os.Exit(1)
	//}

	time.Sleep(6 * time.Second)

	//Create the Resource on top of the above REST API
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
