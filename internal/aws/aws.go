package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
)

func NewSession() *session.Session {

	sess, err := session.NewSession()
	if err != nil {
		log.Println(err.Error())
	}
	return sess
}
