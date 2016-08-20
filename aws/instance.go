package aws

import (
	"fmt"
	"log"
	"sort"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// The instance start timeout, in seconds.
const startTimeout = 300

// The instance type to launch.
const instanceType = "t2.nano"

// The SSH user that is used to log into the default image.
const sshUser = "ec2-user"

// amiSearchParameters returns a DescribeImagesInput struct with the details
// necessary to locate the image that the bastion host will launch. The code
// describes an Amazon Linux AMI, which is the default that gets launched.
func amiSearchParameters() *ec2.DescribeImagesInput {
	return &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("owner-id"),
				Values: aws.StringSlice([]string{"137112412989"}),
			},
			&ec2.Filter{
				Name:   aws.String("owner-alias"),
				Values: aws.StringSlice([]string{"amazon"}),
			},
			&ec2.Filter{
				Name:   aws.String("name"),
				Values: aws.StringSlice([]string{"amzn-ami-hvm-*.x86_64-gp2"}),
			},
			&ec2.Filter{
				Name:   aws.String("description"),
				Values: aws.StringSlice([]string{"Amazon Linux AMI * x86_64 HVM GP2"}),
			},
		},
	}
}

// Instance describes an AWS EC2 instance.
type Instance struct {
	_ struct{}

	// true if the instance has been created.
	Created bool `json:"created"`

	// The ID of the AMI used to launch the instance.
	ImageID string `json:"image_id"`

	// The ID of the instance.
	InstanceID string `json:"instance_id"`

	// The instance type.
	InstanceType string `json:"instance_type"`

	// The subnet for the instance.
	SubnetID string `json:"subnet_id"`

	// The key pair name for SSH access.
	KeyPairName string `json:"key_pair_name"`

	// The security group ID the instance is being launched in.
	SecurityGroupID string `json:"security_group_id"`

	// The public IP address.
	PublicIPAddress string `json:"public_ip_address"`

	// The private IP address.
	PrivateIPAddress string `json:"private_ip_address"`

	// The SSH user to connect to the instance with.
	SSHUser string `json:"ssh_user"`
}

// imageSort is an alias type for []*ec2.Image, used for sorting.
type imageSort []*ec2.Image

// Len is the sort.Interface.Len() implementation for imageSort.
func (a imageSort) Len() int { return len(a) }

// Swap is the sort.Interface.Swap() implementation for imageSort.
func (a imageSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less is the sort.Interface.Less() implementation for imageSort.
func (a imageSort) Less(i, j int) bool {
	itime, _ := time.Parse(time.RFC3339, *a[i].CreationDate)
	jtime, _ := time.Parse(time.RFC3339, *a[j].CreationDate)
	return itime.Unix() < jtime.Unix()
}

// mostRecentAmi returns the most recent AMI out of a slice of images.
func mostRecentAmi(images []*ec2.Image) *ec2.Image {
	sortedImages := images
	sort.Sort(imageSort(sortedImages))
	return sortedImages[len(sortedImages)-1]
}

// waitForInstanceStart waits for the instance to start, and returns the
// properly updated *ec2.Instance object.
func waitForInstanceStart(conn *ec2.EC2, instanceID string, timeout int) (*ec2.Instance, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-id"),
				Values: aws.StringSlice([]string{instanceID}),
			},
		},
	}

	start := time.Now()
	d := time.Duration(timeout) * time.Second
	max := start.Add(d)

	for time.Now().After(max) == false {
		resp, err := conn.DescribeInstances(params)
		if err != nil {
			return nil, err
		}

		if len(resp.Reservations) != 1 {
			panic(fmt.Errorf("Expected a single reservation, got %d", len(resp.Reservations)))
		}

		if len(resp.Reservations[0].Instances) < 1 {
			return nil, fmt.Errorf("No instances were found.")
		}

		if len(resp.Reservations[0].Instances) > 1 {
			panic("More than one instance was found when only one was requested")
		}

		instance := resp.Reservations[0].Instances[0]
		if *instance.State.Name == "running" {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("Instance was not started after %d seconds", timeout)
}

// waitForSSH waits not only for SSH to be running and open, but also ensures
// that the IP address can be reached via the configured SSH user.
func waitForSSH(addr, user string, key KeyPair, timeout int) error {
	signer, err := ssh.ParsePrivateKey([]byte(key.PrivateKeyPEM))
	if err != nil {
		log.Fatalf("Unable to parse private key: %s", err.Error())
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	start := time.Now()
	d := time.Duration(timeout) * time.Second
	max := start.Add(d)

	for time.Now().After(max) == false {
		_, err := ssh.Dial("tcp", addr, config)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("SSH could not be connected after %d seconds", timeout)
}

// LocateImage searches for a suitable AMI to launch, based off the
// filters supplied by amiSearchParameters().
func LocateImage(conn *ec2.EC2) (string, error) {
	params := amiSearchParameters()

	resp, err := conn.DescribeImages(params)
	if err != nil {
		return "", err
	}

	if len(resp.Images) < 1 {
		return "", fmt.Errorf("No default image found. You may need to update bastion.")
	}

	// Sort the images and return the most recent AMI found
	image := mostRecentAmi(resp.Images)

	return *image.ImageId, nil
}

// CreateInstance creates an Amazon EC2 insatnce, and returns an Instance
// struct.
func CreateInstance(conn *ec2.EC2, subnet, securityGroup string, keyPair KeyPair) (Instance, error) {
	instance := Instance{
		SubnetID:        subnet,
		KeyPairName:     keyPair.KeyName,
		SecurityGroupID: securityGroup,
		InstanceType:    instanceType,
		SSHUser:         sshUser,
	}
	// Locate an AMI for the instance
	ami, err := LocateImage(conn)
	if err != nil {
		return instance, err
	}

	// Attempt to launch the instance.
	params := &ec2.RunInstancesInput{
		ImageId:      aws.String(ami),
		InstanceType: aws.String(instanceType),
		KeyName:      aws.String(keyPair.KeyName),
		MaxCount:     aws.Int64(1),
		MinCount:     aws.Int64(1),
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			&ec2.InstanceNetworkInterfaceSpecification{
				AssociatePublicIpAddress: aws.Bool(true),
				DeleteOnTermination:      aws.Bool(true),
				DeviceIndex:              aws.Int64(0),
				Groups:                   aws.StringSlice([]string{securityGroup}),
				SubnetId:                 aws.String(subnet),
			},
		},
	}

	resp, err := conn.RunInstances(params)
	if err != nil {
		return instance, err
	}

	if len(resp.Instances) < 1 {
		return instance, fmt.Errorf("No instances were launched.")
	}

	if len(resp.Instances) > 1 {
		panic("More than one instance was launched when only one was requested")
	}

	// Wait for the instance to be started.
	newInstance, err := waitForInstanceStart(conn, *resp.Instances[0].InstanceId, startTimeout)
	if err != nil {
		return instance, err
	}

	// Wait for SSH off the new instance public IP address
	if newInstance.PublicIpAddress == nil {
		return instance, fmt.Errorf("Instance ID %s does not have a public IP address.", *newInstance.InstanceId)
	}

	err = waitForSSH(*newInstance.PublicIpAddress, sshUser, keyPair, startTimeout)
	if err != nil {
		return instance, err
	}

	// Done
	instance.InstanceID = *newInstance.InstanceId
	instance.PublicIPAddress = *newInstance.PublicIpAddress
	instance.PrivateIPAddress = *newInstance.PrivateIpAddress
	instance.Created = true

	return instance, nil
}

// DeleteInstance terminates an Amazon EC2 instance.
func DeleteInstance(conn *ec2.EC2, instance Instance) (Instance, error) {
	params := &ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{instance.InstanceID}),
	}

	_, err := conn.TerminateInstances(params)
	if err != nil {
		return instance, err
	}

	instance.Created = false
	return instance, nil
}
