package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// testSecurityGroupRule provides a test network ACL rule.
func testSecurityGroupRule() SecurityGroupRule {
	return SecurityGroupRule{
		CidrBlock:   "10.0.1.0/24",
		Created:     true,
		Egress:      false,
		GroupID:     "sg-123456",
		StartPort:   22,
		EndPort:     22,
		PreExisting: false,
	}
}

// testDescribeNetworkAclsOutput provides test data for the stub
// DescribeNetworkAcls function.
func testDescribeSecurityGroupsOutput() *ec2.DescribeSecurityGroupsOutput {
	return &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: []*ec2.SecurityGroup{
			&ec2.SecurityGroup{
				GroupId:   aws.String("sg-123456"),
				GroupName: aws.String("sg-test"),
				IpPermissions: []*ec2.IpPermission{
					&ec2.IpPermission{
						FromPort:   aws.Int64(22),
						IpProtocol: aws.String("tcp"),
						IpRanges:   []*ec2.IpRange{&ec2.IpRange{CidrIp: aws.String("172.16.0.0/24")}},
						ToPort:     aws.Int64(22),
					},
					&ec2.IpPermission{
						FromPort:   aws.Int64(22),
						IpProtocol: aws.String("tcp"),
						IpRanges:   []*ec2.IpRange{&ec2.IpRange{CidrIp: aws.String("10.0.0.0/24")}},
						ToPort:     aws.Int64(22),
					},
				},
				IpPermissionsEgress: []*ec2.IpPermission{
					&ec2.IpPermission{
						FromPort:   aws.Int64(22),
						IpProtocol: aws.String("tcp"),
						IpRanges:   []*ec2.IpRange{&ec2.IpRange{CidrIp: aws.String("10.0.1.0/24")}},
						ToPort:     aws.Int64(22),
					},
				},
				OwnerId: aws.String("123456789012"),
				VpcId:   aws.String("vpc-123456"),
			},
		},
	}
}

// testDescribeSecurityGroups is a stub function for testing the
// ec2.DescribeSecurityGroups function.
func testDescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	if *input.GroupIds[0] == "bad" {
		return nil, fmt.Errorf("error")
	}
	return testDescribeSecurityGroupsOutput(), nil
}

// testAuthorizeSecurityGroupEgress is a stub function for testing the
// *ec2.AuthorizeSecurityGroupEgress function.
func testAuthorizeSecurityGroupEgress(input *ec2.AuthorizeSecurityGroupEgressInput) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	if *input.GroupId == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.AuthorizeSecurityGroupEgressOutput{}, nil
}

// testAuthorizeSecurityGroupIngress is a stub function for testing the
// *ec2.AuthorizeSecurityGroupIngress function.
func testAuthorizeSecurityGroupIngress(input *ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	if *input.GroupId == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.AuthorizeSecurityGroupIngressOutput{}, nil
}

// testRevokeSecurityGroupEgress is a stub function for testing the
// *ec2.RevokeSecurityGroupEgress function.
func testRevokeSecurityGroupEgress(input *ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	if *input.GroupId == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.RevokeSecurityGroupEgressOutput{}, nil
}

// testRevokeSecurityGroupIngress is a stub function for testing the
// *ec2.RevokeSecurityGroupIngress function.
func testRevokeSecurityGroupIngress(input *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	if *input.GroupId == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.RevokeSecurityGroupIngressOutput{}, nil
}

// createTestEC2SGRMock returns a mock EC2 service to use with the security
// group rule functions.
func createTestEC2SGRMock() *ec2.EC2 {
	conn := ec2.New(session.New(), nil)
	conn.Handlers.Clear()

	conn.Handlers.Send.PushBack(func(r *request.Request) {
		switch p := r.Params.(type) {
		case *ec2.DescribeSecurityGroupsInput:
			out, err := testDescribeSecurityGroups(p)
			if out != nil {
				*r.Data.(*ec2.DescribeSecurityGroupsOutput) = *out
			}
			r.Error = err
		case *ec2.AuthorizeSecurityGroupEgressInput:
			out, err := testAuthorizeSecurityGroupEgress(p)
			if out != nil {
				*r.Data.(*ec2.AuthorizeSecurityGroupEgressOutput) = *out
			}
			r.Error = err
		case *ec2.AuthorizeSecurityGroupIngressInput:
			out, err := testAuthorizeSecurityGroupIngress(p)
			if out != nil {
				*r.Data.(*ec2.AuthorizeSecurityGroupIngressOutput) = *out
			}
			r.Error = err
		case *ec2.RevokeSecurityGroupEgressInput:
			out, err := testRevokeSecurityGroupEgress(p)
			if out != nil {
				*r.Data.(*ec2.RevokeSecurityGroupEgressOutput) = *out
			}
			r.Error = err
		case *ec2.RevokeSecurityGroupIngressInput:
			out, err := testRevokeSecurityGroupIngress(p)
			if out != nil {
				*r.Data.(*ec2.RevokeSecurityGroupIngressOutput) = *out
			}
			r.Error = err
		default:
			panic(fmt.Errorf("Unsupported input type %T", p))
		}
	})
	return conn
}

func TestFindPreExistingSecurityGroupRule(t *testing.T) {
	conn := createTestEC2SGRMock()
	group := "sg-123456"
	cidr := "10.0.0.0/24"
	start := 22
	end := 22
	egress := false

	expected := true
	actual, err := FindPreExistingSecurityGroupRule(conn, group, cidr, start, end, egress)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func TestCreateSecurityGroupRule(t *testing.T) {
	conn := createTestEC2SGRMock()
	expected := testSecurityGroupRule()
	group := expected.GroupID
	cidr := expected.CidrBlock
	start := expected.StartPort
	end := expected.EndPort
	egress := expected.Egress

	actual, err := CreateSecurityGroupRule(conn, group, cidr, start, end, egress)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}

	if reflect.DeepEqual(expected, actual) == false {
		t.Fatalf("Expected %#v, got %#v", expected, actual)
	}
}

func TestDeleteSecurityGroupRule(t *testing.T) {
	conn := createTestEC2SGRMock()
	expected := testSecurityGroupRule()

	actual, err := DeleteSecurityGroupRule(conn, expected)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}

	expected.Created = false

	if reflect.DeepEqual(expected, actual) == false {
		t.Fatalf("Expected %#v, got %#v", expected, actual)
	}
}
