package helper

import (
	"log"
	"net/url"
	"testing"

	"github.com/VariableExp0rt/lambda-and-fun/config/session"
	"github.com/aws/aws-sdk-go/aws"
)

func TestCreateRoles(t *testing.T) {
	s := session.NewSession("")

	r, err := CreateRole(Role{
		RoleName:    "Testing1",
		Description: "My testing role",
		Service:     "ec2.amazonaws.com",
	}, s)
	if err != nil {
		log.Printf(err.Error())
	}

	want := "{\"Version\": \"2012-10-17\",\"Statement\": [{\"Effect\": \"Allow\",\"Principal\": {\"Service\": \"" + "ec2.amazonaws.com" + "\"},\"Action\": \"sts:AssumeRole\"}]}"
	got, _ := url.QueryUnescape(aws.StringValue(r.AssumeRolePolicyDocument))
	if got != want {
		t.Errorf("CreateRoles failed, expected %v, got %v", want, got)
	} else {
		log.Printf("CreateRoles successful, expected %v, got %v", want, got)
	}
}

func TestDeleteRoles(t *testing.T) {
	s := session.NewSession("")

	r, _ := DeleteRole("Testing1", s)

	want := string(`{

}`)
	got := r.String()

	if got != want {
		t.Errorf("DeleteRoles failed, expected %v, got %v", want, got)
	} else {
		log.Printf("DeleteRoles successful, expected %v, got %v", want, got)
	}
}
