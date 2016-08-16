package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// testDescribeNetworkAclsOutput provides test data for the stub
// DescribeNetworkAcls function.
func testDescribeNetworkAclsOutput() *ec2.DescribeNetworkAclsOutput {
	return &ec2.DescribeNetworkAclsOutput{
		NetworkAcls: []*ec2.NetworkAcl{
			&ec2.NetworkAcl{
				Associations: []*ec2.NetworkAclAssociation{
					&ec2.NetworkAclAssociation{
						NetworkAclAssociationId: aws.String("assid-123456"),
						NetworkAclId:            aws.String("nacl-123456"),
						SubnetId:                aws.String("subnet-123456"),
					},
				},
				Entries: []*ec2.NetworkAclEntry{
					&ec2.NetworkAclEntry{
						CidrBlock:    aws.String("172.16.0.0/24"),
						Egress:       aws.Bool(false),
						IcmpTypeCode: &ec2.IcmpTypeCode{},
						PortRange:    &ec2.PortRange{From: aws.Int64(22), To: aws.Int64(22)},
						Protocol:     aws.String("TCP"),
						RuleAction:   aws.String("allow"),
						RuleNumber:   aws.Int64(0),
					},
					&ec2.NetworkAclEntry{
						CidrBlock:    aws.String("172.16.0.0/24"),
						Egress:       aws.Bool(true),
						IcmpTypeCode: &ec2.IcmpTypeCode{},
						PortRange:    &ec2.PortRange{From: aws.Int64(1024), To: aws.Int64(65535)},
						Protocol:     aws.String("TCP"),
						RuleAction:   aws.String("allow"),
						RuleNumber:   aws.Int64(0),
					},
					&ec2.NetworkAclEntry{
						CidrBlock:    aws.String("10.0.0.0/24"),
						Egress:       aws.Bool(false),
						IcmpTypeCode: &ec2.IcmpTypeCode{},
						PortRange:    &ec2.PortRange{From: aws.Int64(22), To: aws.Int64(22)},
						Protocol:     aws.String("TCP"),
						RuleAction:   aws.String("allow"),
						RuleNumber:   aws.Int64(100),
					},
					&ec2.NetworkAclEntry{
						CidrBlock:    aws.String("10.0.0.0/24"),
						Egress:       aws.Bool(true),
						IcmpTypeCode: &ec2.IcmpTypeCode{},
						PortRange:    &ec2.PortRange{From: aws.Int64(1024), To: aws.Int64(65535)},
						Protocol:     aws.String("TCP"),
						RuleAction:   aws.String("allow"),
						RuleNumber:   aws.Int64(100),
					},
				},
				IsDefault:    aws.Bool(false),
				NetworkAclId: aws.String("nacl-123456"),
				VpcId:        aws.String("vpc-123456"),
			},
		},
	}
}

// testDescribeNetworkAcls is a stub function for testing the
// ec2.DescribeNetworkAcls function.
func testDescribeNetworkAcls(input *ec2.DescribeNetworkAclsInput) (*ec2.DescribeNetworkAclsOutput, error) {
	if *input.NetworkAclIds[0] == "bad" {
		return nil, fmt.Errorf("error")
	}
	return testDescribeNetworkAclsOutput(), nil
}

// testCreateNetworkAclEntry is a stub function for testing the
// *ec2.CreateNetworkAclEntry function.
func testCreateNetworkAclEntry(input *ec2.CreateNetworkAclEntryInput) (*ec2.CreateNetworkAclEntryOutput, error) {
	if *input.NetworkAclId == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.CreateNetworkAclEntryOutput{}, nil
}

// testDeleteNetworkAclEntry is a stub function for testing the
// *ec2.DeleteNetworkAclEntry function.
func testDeleteNetworkAclEntry(input *ec2.DeleteNetworkAclEntryInput) (*ec2.DeleteNetworkAclEntryOutput, error) {
	if *input.NetworkAclId == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.DeleteNetworkAclEntryOutput{}, nil
}

// createTestEC2NACLMock returns a testEC2NACLMock object, a mock "connection" to use with the network ACL functions.
func createTestEC2NACLMock() *ec2.EC2 {
	conn := ec2.New(session.New(), nil)
	conn.Handlers.Clear()

	conn.Handlers.Send.PushBack(func(r *request.Request) {
		switch p := r.Params.(type) {
		case *ec2.DescribeNetworkAclsInput:
			log.Println("[DEBUG] Executing testDescribeNetworkAcls")
			data := r.Data.(*ec2.DescribeNetworkAclsOutput)
			out, err := testDescribeNetworkAcls(p)
			*data = *out
			r.Error = err
		case *ec2.CreateNetworkAclEntryInput:
			log.Println("[DEBUG] Executing testCreateNetworkAclEntry")
			r.Data, r.Error = testCreateNetworkAclEntry(p)
		case *ec2.DeleteNetworkAclEntryInput:
			log.Println("[DEBUG] Executing testDeleteNetworkAclEntry")
			r.Data, r.Error = testDeleteNetworkAclEntry(p)
		default:
			panic(fmt.Errorf("Unsupported input type %T", p))
		}
	})
	return conn
}

func TestFindVacantNetworkACLRule(t *testing.T) {
	conn := createTestEC2NACLMock()
	acl := "nacl-123456"

	expected := 1
	actual, err := FindVacantNetworkACLRule(conn, acl)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}
