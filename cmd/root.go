package cmd

import (
	"fmt"
	"os"

	"github.com/VariableExp0rt/lambda-and-fun/config/helper"
	"github.com/VariableExp0rt/lambda-and-fun/config/session"
	awssess "github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
)

const (
	version = "0.0.1"
)

var (
	// Account is exported to use in helper package
	Account string
	// Region is exported to use in helper package
	Region string
	sess   *awssess.Session
	// AttachPolArgs is exported to use in helper package
	AttachPolArgs helper.AttachPolicyInput
	// RoleArgs is exported to use in helper package
	RoleArgs helper.Role
	// LambdaArgs is exported to use in helper package
	LambdaArgs helper.Lambda
	// GatewayArgs is exported to use in helper package
	GatewayArgs helper.Gateway
)

var rootCmd = &cobra.Command{
	Use:   "my-app",
	Short: "My App is a program to configure a number of AWS Services",
	Long: `My App is a binary which allows you to build a very quick
				Lambda-Over-HTTPS function to be able to trigger CloudFormation templates with
				as simple cURL command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %v", version)
	},
}

// Execute ensures the root command is executed and read
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&Region, "region", "r", "", "Specify the AWS Region to use.")
	rootCmd.PersistentFlags().StringVarP(&Account, "account", "a", "", "Account ID to be used")

	cmdCreateRole.Flags().StringVar(&RoleArgs.RoleName, "name", "default-role"+helper.R(6, "abcdefghi"+"123456789"), "Define role name.")
	cmdCreateRole.Flags().StringVar(&RoleArgs.Service, "service", "", "Service linked to role, if needed.")
	cmdCreateRole.Flags().StringVar(&RoleArgs.Description, "desc", "A new IAM role for "+RoleArgs.Service, "A short description of the role.")
	cmdCreateRole.MarkFlagRequired("name")
	cmdCreateRole.MarkFlagRequired("service")

	cmdCreateLambda.Flags().StringVar(&LambdaArgs.Description, "desc", "", "Short description of function")
	cmdCreateLambda.Flags().StringVar(&LambdaArgs.FunctionName, "name", "default-function"+helper.R(6, "abcdefghi"+"123456789"), "Name for function")
	cmdCreateLambda.Flags().StringVar(&LambdaArgs.Handler, "handler", "main", "Entrypoint of function")
	cmdCreateLambda.Flags().StringVar(&LambdaArgs.Role, "role", RoleArgs.RoleName, "Link to the role for the service")
	cmdCreateLambda.Flags().StringVar(&LambdaArgs.Runtime, "runtime", "go1.x", "Lambda runtime to use")
	cmdCreateLambda.Flags().StringVar(&LambdaArgs.Code, "code-path", "", "Path to zip file/deployment package")
	cmdCreateLambda.MarkFlagRequired("name")
	cmdCreateLambda.MarkFlagRequired("handler")
	cmdCreateLambda.MarkFlagRequired("role")
	cmdCreateLambda.MarkFlagRequired("runtime")
	cmdCreateLambda.MarkFlagRequired("code-path")

	cmdCreateGateway.Flags().StringVar(&GatewayArgs.Name, "name", "default-gateway"+helper.R(6, "abcdefghi"+"123456789"), "Name of API Gateway")
	cmdCreateGateway.Flags().StringVar(&GatewayArgs.Description, "desc", "", "A description for the API Gateway")
	cmdCreateGateway.Flags().StringVar(&GatewayArgs.FunctionName, "func-name", LambdaArgs.FunctionName, "Supply function name to allow invocation")
	cmdCreateGateway.MarkFlagRequired("name")
	cmdCreateGateway.MarkFlagRequired("func-name")

	cmdDeleteRole.Flags().StringVar(&RoleArgs.RoleName, "name", "", "The name of the Role to be deleted")
	cmdDeleteLambda.Flags().StringVar(&LambdaArgs.FunctionName, "name", "", "The name of the Function to be deleted")
	cmdDeleteGateway.Flags().StringVar(&GatewayArgs.Name, "name", "", "The name of the Gateway to be deleted")
	cmdDeleteRole.MarkFlagRequired("name")
	cmdDeleteLambda.MarkFlagRequired("name")
	cmdDeleteGateway.MarkFlagRequired("name")

	sess = session.NewSession(Region)

}
