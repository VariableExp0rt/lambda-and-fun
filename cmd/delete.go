package cmd

import (
	"fmt"

	"github.com/VariableExp0rt/lambda-and-fun/config/helper"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdDelete)
	cmdDelete.AddCommand(cmdDeleteRole)
	cmdDelete.AddCommand(cmdDeleteLambda)
	cmdDelete.AddCommand(cmdDeleteGateway)
}

var cmdDelete = &cobra.Command{
	Use:   "delete [resources]",
	Short: "Use this to delete AWS resources",
	Long: `Use this command to delete the given AWS resources as a subcommand, with the
				required flags, in order to tear down infrastructure. If you want to delete all resources
				the subcommand is all with a value of 'true'`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Attempting to delete requested resource(s)")
	},
}

var cmdDeleteRole = &cobra.Command{
	Use:   "role [flags]",
	Short: "Delete given IAM role",
	Long: `Delete role will delete the given role, identified by flags, from your AWS environment.
				These individual commands exist to ensure if you make a configuration error, you don't have to
				tear everything down.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		role, err := helper.DeleteRole(RoleArgs.RoleName, sess)
		if err != nil {
			fmt.Printf(err.Error())
		} else {
			fmt.Println("Deleted role: ", role)
		}
	},
}

var cmdDeleteLambda = &cobra.Command{
	Use:   "lambda",
	Short: "Delete a Lambda function",
	Long: `This subcommand will delete a given Lambda resource from your AWS envrionment,
				supply the name of the function.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		lmb, err := helper.DeleteLambda(LambdaArgs.FunctionName, sess)
		if err != nil {
			fmt.Printf(err.Error())
		} else {
			fmt.Println("Function deleted: ", lmb)
		}
	},
}

var cmdDeleteGateway = &cobra.Command{
	Use:   "gateway [flags]",
	Short: "Gateway subcommand deletes an API Gateway service",
	Long: `The Gateway subcommand deletes an API Gateway of the given name from your AWS
			environment, supply the API Gateway name.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		gwy, err := helper.DeleteRestAPI(GatewayArgs.Name, sess)
		if err != nil {
			fmt.Printf(err.Error())
		} else {
			fmt.Println("Gateway (REST API) deleted: ", gwy)
		}
	},
}
