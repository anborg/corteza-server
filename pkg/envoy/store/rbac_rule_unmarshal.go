package store

import (
	"strconv"

	"github.com/cortezaproject/corteza-server/pkg/envoy"
	"github.com/cortezaproject/corteza-server/pkg/envoy/resource"
	"github.com/cortezaproject/corteza-server/pkg/envoy/util"
	"github.com/cortezaproject/corteza-server/pkg/rbac"
)

func newRbacRule(rl *rbac.Rule) *rbacRule {
	return &rbacRule{
		rule: rl,
	}
}

// MarshalEnvoy converts the rbac rule struct to a resource
func (rl *rbacRule) MarshalEnvoy() ([]resource.Interface, error) {
	refRole := strconv.FormatUint(rl.rule.RoleID, 10)

	refRes, err := util.RbacResToRef(rl.rule.Resource.String())
	if err != nil {
		return nil, err
	}

	// Remove the identifier once we're finished with it
	rl.rule.Resource = rl.rule.Resource.TrimID()

	return envoy.CollectNodes(
		resource.NewRbacRule(rl.rule, refRole, refRes),
	)
}
