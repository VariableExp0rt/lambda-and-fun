package session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewSession returns a new initialized session to be passed to service creation for specific AWS services.
// You must have a session before interacting with AWS services which can be initialized using the current
// session as follows; svc := apigateway.New(sess)
func NewSession(region string) *session.Session {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(region),
		},
	}))

	return sess
}
