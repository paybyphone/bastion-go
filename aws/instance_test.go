package aws

import (
	"github.com/aws/aws-sdk-go/aws"
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
		},
	}
}
