package session

import (
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestNewSession(t *testing.T) {
	ns := NewSession("eu-west-2")

	var r = aws.StringValue(ns.Config.Region)
	if r != "eu-west-2" {
		t.Errorf("NewSession(\"eu-west-2\") failed, expected %v, got %v", "eu-west-2", r)
	} else {
		log.Printf("NewSession(\"eu-west-2\") succeeded, expected %v, got %v", "eu-west-2", r)
	}
}
