package mocks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

// DescribeSubnetsResponse returns
type DescribeSubnetsResponse struct {
	Resp  *ec2.DescribeSubnetsOutput
	Error error
}

// DescribeSecurityGroupsResponse returns
type DescribeSecurityGroupsResponse struct {
	Resp  *ec2.DescribeSecurityGroupsOutput
	Error error
}

// EC2Client returns
type EC2Client struct {
	aws.EC2API
	DescribeSecurityGroupsResp map[string]*DescribeSecurityGroupsResponse
	DescribeSubnetsResp        map[string]*DescribeSubnetsResponse
}

func (m *EC2Client) init() {
	if m.DescribeSecurityGroupsResp == nil {
		m.DescribeSecurityGroupsResp = map[string]*DescribeSecurityGroupsResponse{}
	}
	if m.DescribeSubnetsResp == nil {
		m.DescribeSubnetsResp = map[string]*DescribeSubnetsResponse{}
	}
}

// AddSecurityGroup returns
func (m *EC2Client) AddSecurityGroup(name string, projectName string, configName string, serviceName string, err error) {
	m.init()
	m.DescribeSecurityGroupsResp[name] = &DescribeSecurityGroupsResponse{
		Resp: &ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []*ec2.SecurityGroup{
				MakeMockSecurityGroup(name, projectName, configName, serviceName),
			},
		},
		Error: err,
	}
}

// AddSubnet returns
func (m *EC2Client) AddSubnet(nameTag string, id string, tag bool) {
	m.init()
	tags := []*ec2.Tag{
		&ec2.Tag{Key: to.Strp("Name"), Value: to.Strp(nameTag)},
	}

	if tag {
		tags = []*ec2.Tag{
			&ec2.Tag{Key: to.Strp("Name"), Value: to.Strp(nameTag)},
			&ec2.Tag{Key: to.Strp("DeployWithFenrir"), Value: to.Strp("true")},
		}
	}

	m.DescribeSubnetsResp[nameTag] = &DescribeSubnetsResponse{
		Resp: &ec2.DescribeSubnetsOutput{
			Subnets: []*ec2.Subnet{
				&ec2.Subnet{
					SubnetId: to.Strp(id),
					Tags:     tags,
				},
			},
		},
	}
}

// DescribeSecurityGroups returns
func (m *EC2Client) DescribeSecurityGroups(in *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	m.init()
	sgName := in.Filters[0].Values[0]
	resp := m.DescribeSecurityGroupsResp[*sgName]
	if resp == nil {
		return &ec2.DescribeSecurityGroupsOutput{SecurityGroups: []*ec2.SecurityGroup{}}, nil
	}
	return resp.Resp, resp.Error
}

// MakeMockSecurityGroup returns
func MakeMockSecurityGroup(name string, projectName string, configName string, serviceName string) *ec2.SecurityGroup {
	return &ec2.SecurityGroup{
		GroupId: to.Strp("group-id"),
		Tags: []*ec2.Tag{
			&ec2.Tag{Key: to.Strp("Name"), Value: to.Strp(name)},
			&ec2.Tag{Key: to.Strp("ProjectName"), Value: to.Strp(projectName)},
			&ec2.Tag{Key: to.Strp("ConfigName"), Value: to.Strp(configName)},
			&ec2.Tag{Key: to.Strp("ServiceName"), Value: to.Strp(serviceName)},
		},
	}
}

// DescribeSubnets returns
func (m *EC2Client) DescribeSubnets(in *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	if m.DescribeSubnetsResp == nil {
		return nil, fmt.Errorf("Add Subnets")
	}

	if len(in.Filters) != 1 || len(in.Filters[0].Values) != 1 {
		return nil, nil
	}

	subnetName := in.Filters[0].Values[0]
	resp := m.DescribeSubnetsResp[*subnetName]

	if resp == nil {
		return &ec2.DescribeSubnetsOutput{Subnets: []*ec2.Subnet{}}, nil
	}

	return resp.Resp, resp.Error
}
