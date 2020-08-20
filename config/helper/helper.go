package helper

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
)

const (
	policyArnPrefix            = "arn:aws:iam::aws:policy/"
	policyArnPrefixServiceRole = "arn:aws:iam::aws:policy/service-role/"
)

type RequiredArgs struct {
	Account      string
	FunctionName string
	GatewayName  string
	Operation    string
	Region       string
	RoleName     string
	RoleService  string
	Zipfile      []byte
}

// Role is a data structure to hold the number of roles the function should create, the names it should use
// to do so, and a
type Role struct {
	RoleName    string
	Description string
	Service     string
}

// Lambda provides the necessary required fields to create a minimal Lambda function for our purposes, creates a
// service instance using the session created in session.go and then our input as the CreateFunctionInput arguments
// , it returns a *lambda.Function
type Lambda struct {
	Code         []byte
	Description  string
	FunctionName string
	Handler      string
	Runtime      string
	Role         string
}

// Gateway provides the configuration data for creating a REST API, HTTP API, or another kind of gateway
// it uses the session to create a service and then invoke the creation with the parameters supplied by this
// data structure
type Gateway struct {
	Name        string
	Type        string
	Description string
}

//func SetArgs() {
//
//}

// CreateRole creates a given number of IAM roles with the required parameters only as input to the function
// Singular would be easy to express, multiple roles can be created by running this function multiple times
func CreateRole(args Role, sess *session.Session) (*iam.Role, error) {

	svc := iam.New(sess)

	role, err := svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(args.RoleName),
		AssumeRolePolicyDocument: aws.String("{\"Version\": \"2012-10-17\",\"Statement\": [{\"Effect\": \"Allow\",\"Principal\": {\"Service\": \"" + args.Service + "\"},\"Action\": \"sts:AssumeRole\"}]}"),
		Description:              aws.String(args.Description),
	})
	return role.Role, err
}

// AttachPolicy is used to attach policies to roles that have previously been created
// supply managed ARN of the policy e.g. AWSLambdaBasicExecutionRole or AWSEKSClusterPolicy
func AttachPolicy(policy string, roleName string, r *Role, sess *session.Session) (*iam.AttachRolePolicyOutput, error) {
	svc := iam.New(sess)

	var err error
	var res *iam.AttachRolePolicyOutput

	// If is managed policy specifically linked to a role which is linked to a specific service, this applies
	// else use the normal policy prefix instead of a service-role prefix
	var s = func(*Role) string { newStr := strings.Split(r.Service, "."); return newStr[0] }
	if strings.Contains(policy, s(r)) {
		res, err = svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: aws.String(policyArnPrefixServiceRole + policy),
			RoleName:  aws.String(roleName),
		})
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		res, err = svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: aws.String(policyArnPrefix + policy),
			RoleName:  aws.String(roleName),
		})
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return res, err
}

// CreateLambda creates a new Lambda function where Lambda is the input - only the required fields have
// been included for ease
func CreateLambda(l Lambda, sess *session.Session) (*lambda.FunctionConfiguration, error) {
	svc := lambda.New(sess)

	res, err := svc.CreateFunction(&lambda.CreateFunctionInput{
		Code:         &lambda.FunctionCode{ZipFile: l.Code},
		Description:  aws.String(l.Description),
		FunctionName: aws.String(l.FunctionName),
		Handler:      aws.String(l.Handler),
		Role:         aws.String(l.Role),
		Runtime:      aws.String(l.Runtime),
	})
	if err != nil {
		fmt.Printf(err.Error())
	}
	return res, err
}

// CreateGateway creates a new API Gateway (REST of HTTP) with Gateway as the required input
func CreateGateway(g Gateway, sess *session.Session) (*apigateway.RestApi, *string, error) {
	svc := apigateway.New(sess)

	res, err := svc.CreateRestApi(&apigateway.CreateRestApiInput{
		Description: aws.String(g.Description),
		Name:        aws.String(g.Name),
	})
	if err != nil {
		fmt.Printf(err.Error())
	}

	time.Sleep(6 * time.Second)

	rootID := GetAPIParentID(res.Id, sess)

	return res, rootID, err
}

// DeleteRole will delete the given Role resources from the IAM console
func DeleteRole(roleName string, sess *session.Session) (*iam.DeleteRoleOutput, error) {
	svc := iam.New(sess)

	res, err := svc.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	return res, err
}

// DeleteAttachedPolicy will delete the given policies from the nominated role
func DeleteAttachedPolicy(policy string, roleName string, sess *session.Session) (*iam.DeleteRolePolicyOutput, error) {
	svc := iam.New(sess)

	res, err := svc.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
		RoleName:   aws.String(policy),
		PolicyName: aws.String(roleName),
	})
	if err != nil {
		fmt.Printf(err.Error())
	}
	return res, err
}

// DeleteLambda deletes the given function by name
func DeleteLambda(funcName string, sess *session.Session) (*lambda.DeleteFunctionOutput, error) {
	svc := lambda.New(sess)

	res, err := svc.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: aws.String(funcName),
	})
	if err != nil {
		fmt.Printf(err.Error())
	}
	return res, err
}

// DeleteRestAPI deletes the given Rest API, use GetRestApi to see available APIs for deletion
func DeleteRestAPI(name string, sess *session.Session) (*apigateway.DeleteRestApiOutput, error) {
	svc := apigateway.New(sess)

	res, err := svc.DeleteRestApi(&apigateway.DeleteRestApiInput{
		RestApiId: aws.String(name),
	})
	if err != nil {
		fmt.Printf(err.Error())
	}
	return res, err
}

// GetAPIParentID gets the Parent ID of the newly created rest api in order to create the new resource
func GetAPIParentID(apiID *string, sess *session.Session) *string {
	svc := apigateway.New(sess)

	res, err := svc.GetResources(&apigateway.GetResourcesInput{
		RestApiId: apiID,
	})
	if err != nil {
		fmt.Printf(err.Error())
	}
	var ID *string
	for _, item := range res.Items {
		ID = item.Id
		if aws.StringValue(ID) == "" {
			os.Exit(1)
		}
	}
	return ID
}

// ConfigureAPIEndpoint conducts the necessary steps to make the API reachable
func ConfigureAPIEndpoint()

// CreateAllResources is used to make all of the above resources, similarly to a stack, rather than having to
// have a switch statement to trigger each, more logic to be added in the lmabda itself for this
func CreateAllResources()

// DeleteAllResources is the same as the above, but a teardown instead of setting up
func DeleteAllResources()
