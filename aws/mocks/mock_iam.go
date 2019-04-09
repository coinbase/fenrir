package mocks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

// GetRoleResponse returns
type GetRoleResponse struct {
	Resp  *iam.GetRoleOutput
	Error error
}

// IAMClient returns
type IAMClient struct {
	aws.IAMAPI
	GetRoleResp map[string]*GetRoleResponse
}

func (m *IAMClient) init() {
	if m.GetRoleResp == nil {
		m.GetRoleResp = map[string]*GetRoleResponse{}
	}
}

// AddGetRole returns
func (m *IAMClient) AddGetRole(roleName, project, config, service string) {
	m.init()
	m.GetRoleResp[roleName] = &GetRoleResponse{
		Resp: &iam.GetRoleOutput{
			Role: &iam.Role{
				Arn: to.Strp(roleName),
				Tags: []*iam.Tag{
					&iam.Tag{Key: to.Strp("ProjectName"), Value: &project},
					&iam.Tag{Key: to.Strp("ConfigName"), Value: &config},
					&iam.Tag{Key: to.Strp("ServiceName"), Value: &service},
				},
			},
		},
	}
}

// GetRole returns
func (m *IAMClient) GetRole(in *iam.GetRoleInput) (*iam.GetRoleOutput, error) {
	m.init()
	resp := m.GetRoleResp[*in.RoleName]
	if resp == nil {
		return nil, fmt.Errorf("not found role err")
	}
	return resp.Resp, resp.Error
}
