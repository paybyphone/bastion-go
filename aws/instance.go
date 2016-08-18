package aws

import (
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
