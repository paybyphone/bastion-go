package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// testDescribeImagesOutput supplies a real-world DescribeImagesOutput example
// for use in testing.
//
// The image returned is an Amazon linux AMI from us-west-2.
func testDescribeImagesOutput() *ec2.DescribeImagesOutput {
	return &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			&ec2.Image{
				Architecture: aws.String("x86_64"),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					&ec2.BlockDeviceMapping{
						DeviceName: aws.String("/dev/xvda"),
						Ebs: &ec2.EbsBlockDevice{
							DeleteOnTermination: aws.Bool(true),
							Encrypted:           aws.Bool(false),
							SnapshotId:          aws.String("snap-6d465049"),
							VolumeSize:          aws.Int64(8),
							VolumeType:          aws.String("gp2"),
						},
					},
				},
				CreationDate:       aws.String("2014-06-22T09:19:44.000Z"),
				Description:        aws.String("Amazon Linux AMI 2014.03.3 x86_64 HVM GP2"),
				Hypervisor:         aws.String("xen"),
				ImageId:            aws.String("ami-8172b616"),
				ImageLocation:      aws.String("amazon/amzn-ami-hvm-2014.03.3.x86_64-gp2"),
				ImageOwnerAlias:    aws.String("amazon"),
				ImageType:          aws.String("machine"),
				Name:               aws.String("amzn-ami-hvm-2014.03.3.x86_64-gp2"),
				OwnerId:            aws.String("137112412989"),
				Public:             aws.Bool(true),
				RootDeviceName:     aws.String("/dev/xvda"),
				RootDeviceType:     aws.String("ebs"),
				SriovNetSupport:    aws.String("simple"),
				State:              aws.String("available"),
				VirtualizationType: aws.String("hvm"),
			},
			&ec2.Image{
				Architecture: aws.String("x86_64"),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					&ec2.BlockDeviceMapping{
						DeviceName: aws.String("/dev/xvda"),
						Ebs: &ec2.EbsBlockDevice{
							DeleteOnTermination: aws.Bool(true),
							Encrypted:           aws.Bool(false),
							SnapshotId:          aws.String("snap-d465048a"),
							VolumeSize:          aws.Int64(8),
							VolumeType:          aws.String("gp2"),
						},
					},
				},
				CreationDate:       aws.String("2016-06-22T09:19:44.000Z"),
				Description:        aws.String("Amazon Linux AMI 2016.03.3 x86_64 HVM GP2"),
				Hypervisor:         aws.String("xen"),
				ImageId:            aws.String("ami-7172b611"),
				ImageLocation:      aws.String("amazon/amzn-ami-hvm-2016.03.3.x86_64-gp2"),
				ImageOwnerAlias:    aws.String("amazon"),
				ImageType:          aws.String("machine"),
				Name:               aws.String("amzn-ami-hvm-2016.03.3.x86_64-gp2"),
				OwnerId:            aws.String("137112412989"),
				Public:             aws.Bool(true),
				RootDeviceName:     aws.String("/dev/xvda"),
				RootDeviceType:     aws.String("ebs"),
				SriovNetSupport:    aws.String("simple"),
				State:              aws.String("available"),
				VirtualizationType: aws.String("hvm"),
			},
			&ec2.Image{
				Architecture: aws.String("x86_64"),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					&ec2.BlockDeviceMapping{
						DeviceName: aws.String("/dev/xvda"),
						Ebs: &ec2.EbsBlockDevice{
							DeleteOnTermination: aws.Bool(true),
							Encrypted:           aws.Bool(false),
							SnapshotId:          aws.String("snap-5d465048"),
							VolumeSize:          aws.Int64(8),
							VolumeType:          aws.String("gp2"),
						},
					},
				},
				CreationDate:       aws.String("2015-06-22T09:19:44.000Z"),
				Description:        aws.String("Amazon Linux AMI 2015.03.3 x86_64 HVM GP2"),
				Hypervisor:         aws.String("xen"),
				ImageId:            aws.String("ami-7172b612"),
				ImageLocation:      aws.String("amazon/amzn-ami-hvm-2015.03.3.x86_64-gp2"),
				ImageOwnerAlias:    aws.String("amazon"),
				ImageType:          aws.String("machine"),
				Name:               aws.String("amzn-ami-hvm-2015.03.3.x86_64-gp2"),
				OwnerId:            aws.String("137112412989"),
				Public:             aws.Bool(true),
				RootDeviceName:     aws.String("/dev/xvda"),
				RootDeviceType:     aws.String("ebs"),
				SriovNetSupport:    aws.String("simple"),
				State:              aws.String("available"),
				VirtualizationType: aws.String("hvm"),
			},
		},
	}
}

// testEC2Reservation provides a test ec2.Reservation struct.
//
// This type is used in the ec2.DescribeInstances() and ec2.RunInstances()
// functions.
func testEC2Reservation() *ec2.Reservation {
	return &ec2.Reservation{
		Instances: []*ec2.Instance{
			&ec2.Instance{
				State: &ec2.InstanceState{
					Code: aws.Int64(16),
					Name: aws.String("running"),
				},
				InstanceId:       aws.String("i-1234567890abcdef0"),
				PrivateIpAddress: aws.String("10.0.0.1"),
				PublicIpAddress:  aws.String("54.0.0.1"),
			},
		},
	}
}

// testDescribeInstancesOutput provides a test ec2.DescribeInstancesOutput
// object.
func testDescribeInstancesOutput() *ec2.DescribeInstancesOutput {
	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			testEC2Reservation(),
		},
	}
}

// testInstance provides a test Instance struct.
func testInstance() Instance {
	return Instance{
		Created:          true,
		ImageID:          "ami-7172b611",
		InstanceID:       "i-1234567890abcdef0",
		InstanceType:     "t2.nano",
		SubnetID:         "subnet-1234567890abcdef0",
		KeyPairName:      "bastion-test",
		SecurityGroupID:  "sg-1234567890abcdef0",
		PublicIPAddress:  "8.8.8.8",
		PrivateIPAddress: "10.0.0.1",
		SSHUser:          "ec2-user",
	}
}

// testDescribeImages is a stub function for testing the
// ec2.DescribeImages function.
func testDescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	if *input.ImageIds[0] == "bad" {
		return nil, fmt.Errorf("error")
	}
	return testDescribeImagesOutput(), nil
}

// testDescribeInstances is a stub function for testing the
// ec2.DescribeInstances function.
func testDescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	if *input.InstanceIds[0] == "bad" {
		return nil, fmt.Errorf("error")
	}
	return testDescribeInstancesOutput(), nil
}

// testRunInstances is a stub function for testing the
// ec2.RunInstances function.
func testRunInstances(input *ec2.RunInstancesInput) (*ec2.Reservation, error) {
	if *input.PrivateIpAddress == "bad" {
		return nil, fmt.Errorf("error")
	}
	return testEC2Reservation(), nil
}

// testTerminateInstances is a stub function for testing the
// ec2.TerminateInstances function.
func testTerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	if *input.InstanceIds[0] == "bad" {
		return nil, fmt.Errorf("error")
	}
	return &ec2.TerminateInstancesOutput{}, nil
}

// createTestEC2InstanceMock returns a mock EC2 service to use with the
// instance test functions.
func createTestEC2InstanceMock() *ec2.EC2 {
	conn := ec2.New(session.New(), nil)
	conn.Handlers.Clear()

	conn.Handlers.Send.PushBack(func(r *request.Request) {
		switch p := r.Params.(type) {
		case *ec2.DescribeImagesInput:
			out, err := testDescribeImages(p)
			if out != nil {
				*r.Data.(*ec2.DescribeImagesOutput) = *out
			}
			r.Error = err
		case *ec2.DescribeInstancesInput:
			out, err := testDescribeInstances(p)
			if out != nil {
				*r.Data.(*ec2.DescribeInstancesOutput) = *out
			}
			r.Error = err
		case *ec2.RunInstancesInput:
			out, err := testRunInstances(p)
			if out != nil {
				*r.Data.(*ec2.Reservation) = *out
			}
			r.Error = err
		case *ec2.TerminateInstancesInput:
			out, err := testTerminateInstances(p)
			if out != nil {
				*r.Data.(*ec2.TerminateInstancesOutput) = *out
			}
			r.Error = err
		default:
			panic(fmt.Errorf("Unsupported input type %T", p))
		}
	})
	return conn
}
