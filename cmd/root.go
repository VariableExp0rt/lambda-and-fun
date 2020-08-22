package cmd

import (
	"fmt"
	"os"

	"github.com/VariableExp0rt/lambda-and-fun/config/helper"
	"github.com/spf13/cobra"
)

const (
	Version = "0.0.1"
)

var (
	account       string
	region        string
	attachPolArgs helper.AttachPolicyInput
	roleArgs      helper.Role
	lambdaArgs    helper.Lambda
	gatewayArgs   helper.Gateway
)

var rootCmd = &cobra.Command{
	Use:   "service-up",
	Short: "Service Up is a program to configure a number of AWS Services",
	Long: `Service Up is a binary which allows you to build a very quick
				Lambda-Over-HTTPS function to be able to trigger CloudFormation templates with
				as simple cURL command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %v", Version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "", "Specify the AWS Region to use.")
	rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "", "Account ID to be used")

	cmdCreateRole.Flags().StringVar(&roleArgs.RoleName, "name", "default-role"+helper.R(6, "abcdefghi"+"123456789"), "Define role name.")
	cmdCreateRole.Flags().StringVar(&roleArgs.Service, "service", "", "Service linked to role, if needed.")
	cmdCreateRole.Flags().StringVar(&roleArgs.Description, "desc", "A new IAM role for "+roleArgs.Service, "A short description of the role.")
	cmdCreateRole.MarkFlagRequired("name")
	cmdCreateRole.MarkFlagRequired("service")

	cmdCreateLambda.Flags().StringVar(&lambdaArgs.Description, "desc", "", "Short description of function")
	cmdCreateLambda.Flags().StringVar(&lambdaArgs.FunctionName, "name", "default-function"+helper.R(6, "abcdefghi"+"123456789"), "Name for function")
	cmdCreateLambda.Flags().StringVar(&lambdaArgs.Handler, "handler", "main", "Entrypoint of function")
	cmdCreateLambda.Flags().StringVar(&lambdaArgs.Role, "role", roleArgs.RoleName, "Link to the role for the service")
	cmdCreateLambda.Flags().StringVar(&lambdaArgs.Runtime, "runtime", "go1.x", "Lambda runtime to use")
	cmdCreateLambda.Flags().StringVar(&lambdaArgs.Code, "code-path", "", "Path to zip file/deployment package")
	cmdCreateLambda.MarkFlagRequired("name")
	cmdCreateLambda.MarkFlagRequired("handler")
	cmdCreateLambda.MarkFlagRequired("role")
	cmdCreateLambda.MarkFlagRequired("runtime")
	cmdCreateLambda.MarkFlagRequired("code-path")

	cmdCreateGateway.Flags().StringVar(&gatewayArgs.Name, "name", "default-gateway"+helper.R(6, "abcdefghi"+"123456789"), "Name of API Gateway")
	cmdCreateGateway.Flags().StringVar(&gatewayArgs.Description, "desc", "", "A description for the API Gateway")
	cmdCreateGateway.MarkFlagRequired("name")

}
