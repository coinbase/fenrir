package iam

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/coinbase/fenrir/aws"
)

//////
// ROLE
//////

type Role struct {
	Arn  *string
	Tags map[string]*string
}

func (r *Role) ProjectName() *string {
	return r.Tags["ProjectName"]
}

func (r *Role) ConfigName() *string {
	return r.Tags["ConfigName"]
}

func (r *Role) ServiceName() *string {
	return r.Tags["ServiceName"]
}

// RoleExists returns whether profile exists
func GetRole(iamc aws.IAMAPI, roleName *string) (*Role, error) {
	out, err := iamc.GetRole(&iam.GetRoleInput{
		RoleName: roleName,
	})

	if err != nil {
		return nil, err
	}

	outRole := Role{
		Arn:  out.Role.Arn,
		Tags: map[string]*string{},
	}

	if out.Role.Tags != nil {
		for _, tag := range out.Role.Tags {
			if tag.Key == nil {
				continue
			}
			outRole.Tags[*tag.Key] = tag.Value
		}
	}

	return &outRole, nil
}
