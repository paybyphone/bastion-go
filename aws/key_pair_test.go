package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// testKeyPair provides a test AWS EC2 key pair.
func testKeyPair() KeyPair {
	return KeyPair{
		Created:       true,
		Fingerprint:   "Fingerprint",
		KeyName:       "bastion-abcdef0123456789",
		PrivateKeyPEM: "PrivateKeyPEM",
	}
}

// testCreateKeyPairOutput provides test data for the stub
// CreateKeyPair function.
func testCreateKeyPairOutput() *ec2.CreateKeyPairOutput {
	return &ec2.CreateKeyPairOutput{
		KeyFingerprint: aws.String("Fingerprint"),
		KeyMaterial:    aws.String("PrivateKeyPEM"),
		KeyName:        aws.String("bastion-abcdef0123456789"),
	}
}

// testCreateKeyPair is a stub function for testing the
// ec2.CreateKeyPair function.
func testCreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	if *input.KeyName == "bad" {
		return nil, fmt.Errorf("error")
	}
	return testCreateKeyPairOutput(), nil
}

// testDeleteKeyPair is a stub function for testing the
// ec2.DeleteKeyPair function.
func testDeleteKeyPair(input *ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error) {
	if *input.KeyName == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.DeleteKeyPairOutput{}, nil
}

// createTestEC2KPMock returns a mock EC2 service to use with the key pair
// test functions.
func createTestEC2KPMock() *ec2.EC2 {
	conn := ec2.New(session.New(), nil)
	conn.Handlers.Clear()

	conn.Handlers.Send.PushBack(func(r *request.Request) {
		switch p := r.Params.(type) {
		case *ec2.CreateKeyPairInput:
			out, err := testCreateKeyPair(p)
			if out != nil {
				*r.Data.(*ec2.CreateKeyPairOutput) = *out
			}
			r.Error = err
		case *ec2.DeleteKeyPairInput:
			out, err := testDeleteKeyPair(p)
			if out != nil {
				*r.Data.(*ec2.DeleteKeyPairOutput) = *out
			}
			r.Error = err
		default:
			panic(fmt.Errorf("Unsupported input type %T", p))
		}
	})
	return conn
}

func TestCreateKeyPair(t *testing.T) {
	conn := createTestEC2KPMock()

	expectedFingerprint := "Fingerprint"
	expectedCreated := true
	expectedPrivateKeyPEM := "PrivateKeyPEM"
	expectedKeyNameStart := "bastion-"

	out, err := CreateKeyPair(conn)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}

	actualFingerprint := out.Fingerprint
	actualCreated := out.Created
	actualPrivateKeyPEM := out.PrivateKeyPEM
	actualKeyName := out.KeyName

	if expectedFingerprint != actualFingerprint {
		t.Fatalf("Expected fingerprint to be %v, got %v", expectedFingerprint, actualFingerprint)
	}
	if expectedPrivateKeyPEM != actualPrivateKeyPEM {
		t.Fatalf("Expected fingerprint to be %v, got %v", expectedPrivateKeyPEM, actualPrivateKeyPEM)
	}
	if expectedCreated != actualCreated {
		t.Fatalf("Expected created flag to be %v, got %v", expectedCreated, actualCreated)
	}
	matched, _ := regexp.MatchString("^"+expectedKeyNameStart, actualKeyName)
	if matched != true {
		t.Fatalf("Expected name to start with %v, but name is %v", expectedKeyNameStart, actualKeyName)
	}
}

func TestDeleteKeyPair(t *testing.T) {
	conn := createTestEC2KPMock()
	kp := testKeyPair()

	expectedCreated := false

	out, err := DeleteKeyPair(conn, kp)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}

	actualCreated := out.Created

	if expectedCreated != actualCreated {
		t.Fatalf("Expected created flag to be %v, got %v", expectedCreated, actualCreated)
	}
}
