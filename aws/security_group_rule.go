package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// SecurityGroupRule describes an AWS VPC security group rule.
type SecurityGroupRule struct {
	_ struct{}

	// The network range to allow or deny, in CIDR notation (for example 172.16.0.0/24).
	CidrBlock string `json:"cidr_block"`

	// true if the security group rule has been created, or is accounted for (ie: the
	// PreExisting flag is set).
	Created bool `json:"created"`

	// Indicates whether this is an egress rule (rule is applied to traffic leaving
	// the subnet).
	Egress bool `json:"egress"`

	// The ID of the security group the rule is being inserted into.
	GroupID string `json:"security_group_id"`

	// The starting port in the range that this rule applies to.
	StartPort int `json:"start_port"`

	// The starting port in the range that this rule applies to.
	EndPort int `json:"end_port"`

	// "true" if the rule was pre-existing in the exact form that it was going
	// to be created in (ie: direction and port). This is necessary to prevent
	// API errors for duplicate rule entries. Pre-existing rules are not deleted.
	PreExisting bool `json:"pre_existing"`
}

// FindPreExistingSecurityGroupRule will check to see if a rule already exists in
// the security group for a specific direction and port range.
func FindPreExistingSecurityGroupRule(conn *ec2.EC2, group, cidr string, start, end int, egress bool) (bool, error) {
	params := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{group}),
	}

	resp, err := conn.DescribeSecurityGroups(params)
	if err != nil {
		return false, err
	}

	if len(resp.SecurityGroups) < 1 {
		return false, fmt.Errorf("Security group %s not found.", group)
	}

	if len(resp.SecurityGroups) > 1 {
		panic(fmt.Errorf("More than one security group found for security group search %s", group))
	}

	var rules []*ec2.IpPermission
	if egress == true {
		rules = resp.SecurityGroups[0].IpPermissionsEgress
	} else {
		rules = resp.SecurityGroups[0].IpPermissions
	}

	for _, v := range rules {
		for _, x := range v.IpRanges {
			if *x.CidrIp == cidr && int(*v.FromPort) == start && int(*v.ToPort) == end {
				return true, nil
			}
		}
	}

	return false, nil
}

// CreateSecurityGroupRule creates a network ACL rule, and returns a
// NetworkACLRule struct.
//
// If the rule already exists, the struct wiil still be populated, however the
// PreExisting flag will be set to true.
//
// Note that in the event of errors, SecurityGroupRule will be in an inconsistent
// state and should not be used.
func CreateSecurityGroupRule(conn *ec2.EC2, group, cidr string, start, end int, egress bool) (SecurityGroupRule, error) {
	rule := SecurityGroupRule{
		CidrBlock: cidr,
		Egress:    egress,
		GroupID:   group,
		StartPort: start,
		EndPort:   end,
	}

	// Check for pre-existing rules first
	exists, err := FindPreExistingSecurityGroupRule(conn, group, cidr, start, end, egress)
	if err != nil {
		return rule, err
	}
	if exists == true {
		rule.PreExisting = true
		rule.Created = true
		return rule, nil
	}

	if egress == true {
		req := &ec2.AuthorizeSecurityGroupEgressInput{
			CidrIp:     aws.String(cidr),
			FromPort:   aws.Int64(int64(start)),
			IpProtocol: aws.String("tcp"),
			ToPort:     aws.Int64(int64(end)),
			GroupId:    aws.String(group),
		}
		_, err = conn.AuthorizeSecurityGroupEgress(req)
		if err != nil {
			return rule, err
		}
	} else {
		req := &ec2.AuthorizeSecurityGroupIngressInput{
			CidrIp:     aws.String(cidr),
			FromPort:   aws.Int64(int64(start)),
			IpProtocol: aws.String("tcp"),
			ToPort:     aws.Int64(int64(end)),
			GroupId:    aws.String(group),
		}
		_, err = conn.AuthorizeSecurityGroupIngress(req)
		if err != nil {
			return rule, err
		}
	}

	rule.Created = true
	return rule, nil
}

// runSecurityGroupRuleDelete runs most of the logic for
// DeleteSecurityGroupRule, but does not set Created to false.
func runSecurityGroupRuleDelete(conn *ec2.EC2, rule SecurityGroupRule) error {
	// do nothing if the rule was pre-existing.
	if rule.PreExisting == true {
		return nil
	}

	if rule.Egress == true {
		req := &ec2.RevokeSecurityGroupEgressInput{
			CidrIp:     aws.String(rule.CidrBlock),
			FromPort:   aws.Int64(int64(rule.StartPort)),
			IpProtocol: aws.String("tcp"),
			ToPort:     aws.Int64(int64(rule.EndPort)),
			GroupId:    aws.String(rule.GroupID),
		}
		_, err := conn.RevokeSecurityGroupEgress(req)
		if err != nil {
			return err
		}
	} else {
		req := &ec2.RevokeSecurityGroupIngressInput{
			CidrIp:     aws.String(rule.CidrBlock),
			FromPort:   aws.Int64(int64(rule.StartPort)),
			IpProtocol: aws.String("tcp"),
			ToPort:     aws.Int64(int64(rule.EndPort)),
			GroupId:    aws.String(rule.GroupID),
		}
		_, err := conn.RevokeSecurityGroupIngress(req)
		if err != nil {
			return err
		}
	}

	rule.Created = false
	return nil
}

// DeleteSecurityGroupRule deletes a security group rule, if it was not pre-existing.
func DeleteSecurityGroupRule(conn *ec2.EC2, rule SecurityGroupRule) (SecurityGroupRule, error) {
	err := runSecurityGroupRuleDelete(conn, rule)
	if err != nil {
		return rule, err
	}

	rule.Created = false
	return rule, nil
}
