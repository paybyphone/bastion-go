package aws

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// NetworkACLRule describes an AWS VPC network ACL rule.
type NetworkACLRule struct {
	_ struct{}

	// The network range to allow or deny, in CIDR notation (for example 172.16.0.0/24).
	CidrBlock string `json:"cidr_block"`

	// true if the network ACL rule has been created, or is accounted for (ie: the
	// PreExisting flag is set).
	Created bool `json:"created"`

	// Indicates whether this is an egress rule (rule is applied to traffic leaving
	// the subnet).
	Egress bool `json:"egress"`

	// The ID of the network ACL the rule is being inserted into.
	NetworkAclID string `json:"network_acl_id"`

	// The starting port in the range that this rule applies to. Normally this
	// will be the same as EndPort, with the exception of ephemeral rules.
	StartPort int `json:"start_port"`

	// The starting port in the range that this rule applies to. Normally this
	// will be the same as StartPort, with the exception of ephemeral rules.
	EndPort int `json:"end_port"`

	// "true" if the rule was pre-existing in the exact form that it was going
	// to be created in (ie: direction and port). This is necessary to prevent
	// API errors for duplicate ACL entries. Pre-existing rules are not deleted.
	PreExisting bool `json:"pre_existing"`

	// The rule number for the entry (for example, 100). ACL entries are processed
	// in ascending order by rule number.
	//
	// Constraints: Positive integer from 1 to 32766. The range 32767 to 65535
	// is reserved for internal use.
	RuleNumber int `json:"rule_number"`
}

// FindVacantNetworkACLRule will find the highest priority entry (that is,
// the lowest rule number) available in a network ACL to use to add the
// bastion allow rule to.
func FindVacantNetworkACLRule(conn *ec2.EC2, acl string) (int, error) {
	req := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{acl}),
	}

	resp, err := conn.DescribeNetworkAcls(req)
	if err != nil {
		return 0, err
	}

	if len(resp.NetworkAcls) < 1 {
		return 0, fmt.Errorf("Network ACL %s not found.", acl)
	}

	if len(resp.NetworkAcls) > 1 {
		panic(fmt.Errorf("More than one network ACL found for newtork ACL search %s", acl))
	}

	nums := []int{}
	for _, v := range resp.NetworkAcls[0].Entries {
		nums = append(nums, int(*v.RuleNumber))
	}
	sort.Ints(nums)

	n := 0
	for _, v := range nums {
		if v != n {
			break
		}
		n++
	}

	return n, nil
}

// FindPreExistingNetworkACLRule will check to see if a rule already exists in
// an ACL for a specific direction and port range. If the rule exists, the
// rule number is returned, otherwise the result is -1.
//
// Note that error needs to be checked for errors, as the zero value returned
// during errors could be interpreted as rule number 0 as well.
func FindPreExistingNetworkACLRule(conn *ec2.EC2, acl, cidr string, start, end int, egress bool) (int, error) {
	req := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{acl}),
	}

	resp, err := conn.DescribeNetworkAcls(req)
	if err != nil {
		return 0, err
	}

	if len(resp.NetworkAcls) < 1 {
		return 0, fmt.Errorf("Network ACL %s not found.", acl)
	}

	if len(resp.NetworkAcls) > 1 {
		panic(fmt.Errorf("More than one network ACL found for newtork ACL search %s", acl))
	}

	for _, v := range resp.NetworkAcls[0].Entries {
		if *v.CidrBlock == cidr && int(*v.PortRange.From) == start && int(*v.PortRange.To) == end && *v.Egress == egress {
			return int(*v.RuleNumber), nil
		}
	}

	return -1, nil
}

// CreateNetworkACLRule creates a network ACL rule, and returns a
// NetworkACLRule struct.
//
// If the rule already exists, the struct wiil still be populated, however the
// PreExisting flag will be set to true.
//
// Note that in the event of errors, NetworkACLRule will be in an inconsistent
// state and should not be used.
func CreateNetworkACLRule(conn *ec2.EC2, acl, cidr string, start, end int, egress bool) (NetworkACLRule, error) {
	rule := NetworkACLRule{
		CidrBlock:    cidr,
		Egress:       egress,
		NetworkAclID: acl,
		StartPort:    start,
		EndPort:      end,
	}

	// Check for pre-existing rules first
	n, err := FindPreExistingNetworkACLRule(conn, acl, cidr, start, end, egress)
	if err != nil {
		return rule, err
	}
	if n != -1 {
		rule.PreExisting = true
		rule.RuleNumber = n
		rule.Created = true
		return rule, nil
	}

	// No pre-existing rule, look for first vacant rule number.
	n, err = FindVacantNetworkACLRule(conn, acl)
	if err != nil {
		return rule, err
	}

	// Create the rule
	req := &ec2.CreateNetworkAclEntryInput{
		// The network range to allow or deny, in CIDR notation (for example 172.16.0.0/24).
		CidrBlock:    aws.String(cidr),
		Egress:       aws.Bool(egress),
		NetworkAclId: aws.String(acl),
		PortRange:    &ec2.PortRange{From: aws.Int64(int64(start)), To: aws.Int64(int64(end))},
		Protocol:     aws.String("TCP"),
		RuleAction:   aws.String("allow"),
		RuleNumber:   aws.Int64(int64(n)),
	}

	_, err = conn.CreateNetworkAclEntry(req)
	if err != nil {
		return rule, err
	}

	rule.RuleNumber = n
	rule.Created = true
	return rule, nil
}

// runNetworkACLRuleDelete runs most of the logic for DeleteNetworkACLRule,
// but does not set Created to false - that gets performed by
// RunNetworkACLRuleDelete, which wraps this function.
func runNetworkACLRuleDelete(conn *ec2.EC2, rule NetworkACLRule) error {
	// do nothing if the rule was pre-existing.
	if rule.PreExisting == true {
		return nil
	}

	req := &ec2.DeleteNetworkAclEntryInput{
		Egress:       aws.Bool(rule.Egress),
		NetworkAclId: aws.String(rule.NetworkAclID),
		RuleNumber:   aws.Int64(int64(rule.RuleNumber)),
	}

	_, err := conn.DeleteNetworkAclEntry(req)
	if err != nil {
		return err
	}

	rule.Created = false
	return nil
}

// DeleteNetworkACLRule deletes a newtork ACL rule, if it was not pre-existing.
func DeleteNetworkACLRule(conn *ec2.EC2, rule NetworkACLRule) (NetworkACLRule, error) {
	err := runNetworkACLRuleDelete(conn, rule)
	if err != nil {
		return rule, err
	}

	rule.Created = false
	return rule, nil
}
