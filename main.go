package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
)

//Be sure to remove these if they aren't needed
var (
	fileVar, ACC string
	region       = &rPtr
	rPtr         = "eu-west-2"
	account      = &aPtr

	//Set account ID variable instead, safer
	aPtr = os.Getenv(ACC)
)

//Marshal the file contents into a slice of Byte
func getDeploymentPackage() []byte {

	file, err := os.Open(fileVar)
	if err != nil {
		fmt.Printf("Error reading file: %v", file)
	}

	ext := filepath.Ext(fileVar)
	if ext != "zip" {
		fmt.Printf("Must supply deployment package as zip: %v", ext)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)

	if !scanner.Scan() {
		log.Printf("Error reading from scanner: %v", scanner.Err())
		return nil
	}
	obj := scanner.Bytes()
	return obj
}

func main() {
	// arguments to be passed to the commandline when invoking the program
	// TODO
	flag.StringVar(&fileVar, "-file-path", "", "Path to deployment package (zip file)")
	//flag.StringVar(&runtime, "runtime", "go1.x", "Specify runtime for Lambda, default is go1.x.")
	flag.Parse()

	RoleName := "Amazon-Lambda-EKS-Role"

	if len(os.Args) == 2 {
		RoleName = os.Args[1]
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
		Description:              aws.String("A Role to allow Lambda to perform it's basic functions and interact with EKS"),
		PermissionsBoundary:      aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		RoleName:                 aws.String(RoleName),
	}

	//CreateRole will create a new managed IAM Role, and can only have one managed policy attached to
	//it in the above configuration. Multiple can be assigned by calling AttachRolePolicy as many
	//times as needed
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

	//attach additional policies which will allow Lambda to do extra things
	res, err := svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"),
		RoleName:  aws.String(RoleName),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Successfully attached policy to Role", res)

	//This is the primary policy we want the Lambda to be able to use when 'Assuming' the Role
	res1, err := svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"),
		RoleName:  aws.String(RoleName),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Successfully attached policy to Role", res1)

	//Quickly call the function to get the deployment package
	pkg := getDeploymentPackage()

	lmdsvc := lambda.New(sess)

	var FuncName = "my-example-function"
	if len(os.Args) == 2 {
		FuncName = os.Args[2]
	}

	//Generate the Function Input
	funcParams := &lambda.CreateFunctionInput{
		Code:         &lambda.FunctionCode{ZipFile: pkg},
		Description:  aws.String("This function manages the creation and deletion of Amazon EKS Clusters"),
		FunctionName: aws.String(FuncName),
		Role:         &RoleName,
		Runtime:      aws.String("go1.x"),
	}

	//Create the function and wait until it exists before trying to create the API Gateway resource for it
	resp2, err := lmdsvc.CreateFunction(funcParams)
	if err != nil {
		fmt.Println(err.Error())
	}

	if err := lmdsvc.WaitUntilFunctionActive(&lambda.GetFunctionConfigurationInput{
		FunctionName: &FuncName,
	}); err != nil {
		fmt.Println(err.Error(), resp2)
	}

	fmt.Println("Function was created successfully and is currently active", FuncName)

	// A helper for referencing within the API Gateway PutIntegration below
	function, err := lmdsvc.GetFunctionConfiguration(&lambda.GetFunctionConfigurationInput{FunctionName: aws.String(FuncName)})

	//Initialise another service for the API Gateway
	apigwsvc := apigateway.New(sess)

	//Create the REST API named EKS-Setup
	resp3, err := apigwsvc.CreateRestApi(&apigateway.CreateRestApiInput{
		Name: aws.String("EKS-Setup"),
	})
	if err != nil {
		fmt.Println("Error creating REST API")
	}

	//Grab the ID of the newly created REST API, this is needed below to find the root object id
	api := resp3.Id

	//add logic to get the root path or parent ID
	parentAPI, err := apigwsvc.GetResource(&apigateway.GetResourceInput{RestApiId: aws.String(*api)})
	if err != nil {
		fmt.Printf("Unable to get parent REST API ID: %v", err)
	}

	//Create the Resource on top of the above REST API
	resource, err := apigwsvc.CreateResource(&apigateway.CreateResourceInput{
		RestApiId: api,
		PathPart:  resp3.Name,
		ParentId:  parentAPI.ParentId,
	})
	if err != nil {
		fmt.Printf("Error assigning resource %v to API Gateway: %v", resource, resp3)
	}

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
		fmt.Printf("Unable to create put method %v for resource %v", method, resource)
	}

	//Enable the integration responsible for allowing the API Gateway to forward requests/trigger
	//the Lambda function
	integration, err := apigwsvc.PutIntegration(&apigateway.PutIntegrationInput{
		HttpMethod:            aws.String("POST"),
		IntegrationHttpMethod: aws.String("POST"),
		RestApiId:             api,
		ResourceId:            resourceID,
		Type:                  aws.String("AWS"),
		Uri:                   aws.String(*function.FunctionArn),
	})
	if err != nil {
		fmt.Printf("Unable to associate method with function for %v: %v", integration, err)
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
		fmt.Printf("Error putting method response %v", methodResp)
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
		fmt.Printf("Error deploying newly created API %v: %v", deploy, err)
	}

	var REGION, API, ACCOUNT string
	os.Setenv(REGION, *region)
	os.Setenv(API, *api)
	os.Setenv(ACCOUNT, *account)

	sArn := "arn:aws:execute-api:$REGION:$ACCOUNT:$API/*/POST/EKS-Setup"

	perms, err := lmdsvc.AddPermission(&lambda.AddPermissionInput{
		FunctionName: &FuncName,
		StatementId:  aws.String("apigateway-test-2"),
		Action:       aws.String("lambda:InvokeFunction"),
		SourceArn:    aws.String(sArn),
	})
	if err != nil {
		fmt.Printf("Error adding permission to API Gateway %v: %v", perms, err)
	}

	perms1, err := lmdsvc.AddPermission(&lambda.AddPermissionInput{
		FunctionName: &FuncName,
		StatementId:  aws.String("apigateway-prod-2"),
		Action:       aws.String("lambda:InvokeFunction"),
		SourceArn:    aws.String(sArn),
	})
	if err != nil {
		fmt.Printf("Error adding permission to API Gateway %v: %v", perms1, err)
	}

	//var json = []byte(`{"operation": "create"}`)
	//result, err := http.Post(`https://$API.execute-api.$REGION.amazonaws.com/prod/EKS-Setup`, "application/json", bytes.NewBuffer(json))

	//if result.StatusCode == 200 {
	//	fmt.Println("Successfully made request to API Gateway:", result)
	//} else if result.StatusCode != 200 {
	//	fmt.Printf("Error making request to API Gateway: %v", err)
	//}

	testResult, err := apigwsvc.TestInvokeMethod(&apigateway.TestInvokeMethodInput{
		Body:                aws.String(`{"operation": "create"}`),
		HttpMethod:          aws.String("POST"),
		PathWithQueryString: aws.String(""),
		ResourceId:          resourceID,
		RestApiId:           api,
	})
	if err != nil || testResult.Status != aws.Int64(200) {
		fmt.Printf("Error invoking test of method on API Gateway: %v", err)
	}

	fmt.Println("Successfully deployed API Gateway and tested with the following response: ", testResult.Body)
}
