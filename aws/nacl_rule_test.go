package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// testNetworkACLRule provides a test network ACL rule.
func testNetworkACLRule() NetworkACLRule {
	return NetworkACLRule{
		CidrBlock:    "10.0.1.0/24",
		Created:      true,
		Egress:       false,
		NetworkAclID: "nacl-123456",
		StartPort:    22,
		EndPort:      22,
		PreExisting:  false,
		RuleNumber:   1,
	}
}

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

// createTestEC2NACLMock returns a mock EC2 service to use with the network
// ACL test functions.
func createTestEC2NACLMock() *ec2.EC2 {
	conn := ec2.New(session.New(), nil)
	conn.Handlers.Clear()

	conn.Handlers.Send.PushBack(func(r *request.Request) {
		switch p := r.Params.(type) {
		case *ec2.DescribeNetworkAclsInput:
			out, err := testDescribeNetworkAcls(p)
			if out != nil {
				*r.Data.(*ec2.DescribeNetworkAclsOutput) = *out
			}
			r.Error = err
		case *ec2.CreateNetworkAclEntryInput:
			out, err := testCreateNetworkAclEntry(p)
			if out != nil {
				*r.Data.(*ec2.CreateNetworkAclEntryOutput) = *out
			}
			r.Error = err
		case *ec2.DeleteNetworkAclEntryInput:
			out, err := testDeleteNetworkAclEntry(p)
			if out != nil {
				*r.Data.(*ec2.DeleteNetworkAclEntryOutput) = *out
			}
			r.Error = err
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

func TestFindPreExistingNetworkACLRule(t *testing.T) {
	conn := createTestEC2NACLMock()
	acl := "nacl-123456"
	cidr := "10.0.0.0/24"
	start := 22
	end := 22
	egress := false

	expected := 100
	actual, err := FindPreExistingNetworkACLRule(conn, acl, cidr, start, end, egress)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func TestCreateNetworkACLRule(t *testing.T) {
	conn := createTestEC2NACLMock()
	acl := "nacl-123456"
	cidr := "10.0.1.0/24"
	start := 22
	end := 22
	egress := false

	expectedRule := 1
	expectedCreated := true

	out, err := CreateNetworkACLRule(conn, acl, cidr, start, end, egress)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}
	actualRule := out.RuleNumber
	actualCreated := out.Created

	if expectedRule != actualRule {
		t.Fatalf("Expected rule to be %v, got %v", expectedRule, actualRule)
	}
	if expectedCreated != actualCreated {
		t.Fatalf("Expected created to be %v, got %v", expectedCreated, actualCreated)
	}
}

func TestDeleteNetworkACLRule(t *testing.T) {
	conn := createTestEC2NACLMock()
	acl := testNetworkACLRule()

	expectedCreated := false

	out, err := DeleteNetworkACLRule(conn, acl)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}

	actualCreated := out.Created
	if expectedCreated != actualCreated {
		t.Fatalf("Expected created to be %v, got %v", expectedCreated, actualCreated)
	}
}
