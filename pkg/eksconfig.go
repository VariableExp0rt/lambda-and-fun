package eksconfig

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/awslabs/goformation/cloudformation"
	cldfmt "github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/eks"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
)

var (
	resource   = "ExampleCluster"
	securityID = "test"
	vpcId      = os.Getenv("VPC_ID")
	sess       = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, Config: aws.Config{Region: aws.String("eu-west-2")}}))
)

func getVPCInfo() (*ec2.DescribeVpcsOutput, error) {

	svc := ec2.New(sess)

	vpc, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{vpcId}),
	})
	if err != nil {
		fmt.Printf("Error getting VPC information: %v", err)
	}

	return vpc, err
}

func createConfigurationTemplate() *cldfmt.Template {

	template := cldfmt.NewTemplate()

	//TODO: Add template resource for VPC and subnets to be used in Cluster and NodeGroup configuration
	template.Resources["eksClusterRole"] = &iam.Role{
		AssumeRolePolicyDocument: "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Effect\": \"Allow\",\"Principal\": {\"Service\": \"eks.amazonaws.com\"},\"Action\": \"sts:AssumeRole\"}]}",
		ManagedPolicyArns:        []string{"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"},
		RoleName:                 "default-role" + strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}

	var role, roleArn string
	for _, k := range template.Resources {
		if k.AWSCloudFormationType() == "Amazon::IAM::Role" {
			role = cldfmt.Ref("eksClusterRole")
			roleArn = cldfmt.GetAtt(role, "RoleArn")
			if roleArn == "" {
				fmt.Printf("Error fetching Role ARN from resource: %v", role)
			}
		}
	}

	template.Resources[resource] = &eks.Cluster{
		AWSCloudFormationDeletionPolicy: "Delete",
		Name:                            "my-example-cluster" + strconv.FormatInt(time.Now().Unix(), 10),
		ResourcesVpcConfig: &eks.Cluster_ResourcesVpcConfig{
			SubnetIds: []string{"subnet-123a"},
		},
		//Needs to be dynamically created as part of the stack to ensure this can be dynamically assigned
		RoleArn: roleArn,
	}

	var ref string
	for _, k := range template.Resources {
		if k.AWSCloudFormationType() == "AWS::EKS::Cluster" {
			ref = cloudformation.Ref(resource)
		}
	}

	template.Resources["ExampleNodeGroup"] = &eks.Nodegroup{
		AmiType:       "ubuntu",
		ClusterName:   cloudformation.GetAtt(ref, "Name"),
		InstanceTypes: []string{"t3.large"},
	}

	return template
}
