package aws

import (
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// keyPairNamePrefix is the prefix that is applied to auto-generated
// key pairs.
const keyPairNamePrefix = "bastion-"

// KeyPair describes an AWS EC2 key pair.
type KeyPair struct {
	_ struct{}

	// true if the network ACL rule has been created, or is accounted for (ie: the
	// PreExisting flag is set).
	Created bool `json:"created"`

	// The SHA-1 digest of the DER encoded private key.
	Fingerprint string `json:"fingerprint"`

	// The unique name for the key pair.
	KeyName string `json:"key_name"`

	// The private key, in PEM format.
	PrivateKeyPEM string `json:"private_key_pem"`
}

// generateKeyPairName creates an randomly-generated key pair name.
func generateKeyPairName() string {
	id := fmt.Sprintf("%x", rand.Int())
	return securityGroupNamePrefix + id
}

// CreateKeyPair creates an AWS EC2 key pair.
//
// Note that in the event of errors, KeyPair will be in an inconsistent
// state and should not be used.
func CreateKeyPair(conn *ec2.EC2) (KeyPair, error) {
	name := generateKeyPairName()
	var kp KeyPair
	kp.KeyName = name

	params := &ec2.CreateKeyPairInput{
		KeyName: aws.String(name),
	}

	resp, err := conn.CreateKeyPair(params)
	if err != nil {
		return kp, err
	}

	kp.Fingerprint = *resp.KeyFingerprint
	kp.PrivateKeyPEM = *resp.KeyMaterial
	kp.Created = true

	return kp, nil
}

// DeleteKeyPair deletes an AWS EC2 key pair.
func DeleteKeyPair(conn *ec2.EC2, kp KeyPair) (KeyPair, error) {
	params := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(kp.KeyName),
	}

	_, err := conn.DeleteKeyPair(params)
	if err != nil {
		return kp, err
	}

	kp.Created = false
	return kp, nil
}
