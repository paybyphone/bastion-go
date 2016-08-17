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

// testSecurityGroup provides a test security group.
func testSecurityGroup() SecurityGroup {
	return SecurityGroup{
		Created:   true,
		GroupID:   "sg-123456",
		GroupName: "bastion-abcdefgh0123456789",
		VpcID:     "vpc-123456",
	}
}

// testDescribeSubnetsOutput provides test data for the stub
// DescribeNetworkAcls function.
func testDescribeSubnetsOutput() *ec2.DescribeSubnetsOutput {
	return &ec2.DescribeSubnetsOutput{
		Subnets: []*ec2.Subnet{
			&ec2.Subnet{
				AvailabilityZone:        aws.String("us-west-2"),
				AvailableIpAddressCount: aws.Int64(7),
				CidrBlock:               aws.String("10.0.0.0/24"),
				DefaultForAz:            aws.Bool(true),
				MapPublicIpOnLaunch:     aws.Bool(false),
				State:                   aws.String("available"),
				SubnetId:                aws.String("subnet-123456"),
				VpcId:                   aws.String("vpc-123456"),
			},
		},
	}
}

// testDescribeSubnets is a stub function for testing the
// ec2.DescribeSubnets function.
func testDescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	if *input.SubnetIds[0] == "bad" {
		return nil, fmt.Errorf("error")
	}
	return testDescribeSubnetsOutput(), nil
}

// testCreateSecurityGroup is a stub function for testing the
// *ec2.CreateSecurityGroup function.
func testCreateSecurityGroup(input *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {
	if *input.GroupName == "bad" {
		return nil, fmt.Errorf("error")
	}
	out := &ec2.CreateSecurityGroupOutput{
		GroupId: aws.String("sg-123456"),
	}
	return out, nil
}

// testDeleteSecurityGroup is a stub function for testing the
// *ec2.DeleteSecurityGroup function.
func testDeleteSecurityGroup(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	if *input.GroupId == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.DeleteSecurityGroupOutput{}, nil
}

// createTestEC2SGMock returns a mock EC2 service to use with the security group
// test functions.
func createTestEC2SGMock() *ec2.EC2 {
	conn := ec2.New(session.New(), nil)
	conn.Handlers.Clear()

	conn.Handlers.Send.PushBack(func(r *request.Request) {
		switch p := r.Params.(type) {
		case *ec2.DescribeSubnetsInput:
			out, err := testDescribeSubnets(p)
			if out != nil {
				*r.Data.(*ec2.DescribeSubnetsOutput) = *out
			}
			r.Error = err
		case *ec2.CreateSecurityGroupInput:
			out, err := testCreateSecurityGroup(p)
			if out != nil {
				*r.Data.(*ec2.CreateSecurityGroupOutput) = *out
			}
			r.Error = err
		case *ec2.DeleteSecurityGroupInput:
			out, err := testDeleteSecurityGroup(p)
			if out != nil {
				*r.Data.(*ec2.DeleteSecurityGroupOutput) = *out
			}
			r.Error = err
		default:
			panic(fmt.Errorf("Unsupported input type %T", p))
		}
	})
	return conn
}

func TestFindVpcIDFromSubnet(t *testing.T) {
	conn := createTestEC2SGMock()
	subnet := "subnet-123456"

	expected := "vpc-123456"
	actual, err := findVpcIDFromSubnet(conn, subnet)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func TestCreateSecurityGroup(t *testing.T) {
	conn := createTestEC2SGMock()
	subnet := "subnet-123456"

	expectedSgID := "sg-123456"
	expectedCreated := true
	expectedSgNameStart := "bastion-"

	out, err := CreateSecurityGroup(conn, subnet)
	if err != nil {
		t.Fatalf("Bad: %s", err.Error())
	}

	actualSgID := out.GroupID
	actualCreated := out.Created
	actualSgName := out.GroupName

	if expectedSgID != actualSgID {
		t.Fatalf("Expected security group ID to be %v, got %v", expectedSgID, actualSgID)
	}
	if expectedCreated != actualCreated {
		t.Fatalf("Expected created flag to be %v, got %v", expectedCreated, actualCreated)
	}
	matched, _ := regexp.MatchString("^"+expectedSgNameStart, actualSgName)
	if matched != true {
		t.Fatalf("Expected name to start with %v, but name is %v", expectedSgNameStart, actualSgName)
	}
}
