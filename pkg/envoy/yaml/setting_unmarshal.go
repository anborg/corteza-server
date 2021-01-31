package yaml

import (
	"encoding/json"

	"github.com/cortezaproject/corteza-server/pkg/envoy"
	"github.com/cortezaproject/corteza-server/pkg/envoy/resource"
	"github.com/cortezaproject/corteza-server/system/types"
	sqlt "github.com/jmoiron/sqlx/types"
	"gopkg.in/yaml.v3"
)

func (wset *settingSet) UnmarshalYAML(n *yaml.Node) error {
	return each(n, func(k, v *yaml.Node) (err error) {
		var (
			wrap = &setting{}
		)

		if v == nil {
			return nodeErr(n, "malformed setting definition")
		}

		wrap.res = &types.SettingValue{}

		switch v.Kind {
		case yaml.MappingNode:
			if err = v.Decode(&wrap); err != nil {
				return
			}

		default:
			jj, err := json.Marshal(v.Value)
			if err != nil {
				return nodeErr(n, err.Error())
			}
			wrap.res.Value = jj

			if err = decodeScalar(k, "setting", &wrap.res.Name); err != nil {
				return err
			}
		}

		*wset = append(*wset, wrap)
		return
	})
}

func (wset settingSet) MarshalEnvoy() ([]resource.Interface, error) {
	nn := make([]resource.Interface, 0, len(wset))

	for _, res := range wset {
		if tmp, err := res.MarshalEnvoy(); err != nil {
			return nil, err
		} else {
			nn = append(nn, tmp...)
		}

	}

	return nn, nil
}

func (wrap *setting) UnmarshalYAML(n *yaml.Node) (err error) {
	if !isKind(n, yaml.MappingNode) {
		return nodeErr(n, "setting definition must be a map")
	}

	if wrap.res == nil {
		wrap.res = &types.SettingValue{}
	}

	// if err = n.Decode(&wrap.res); err != nil {
	// 	return
	// }

	err = eachMap(n, func(k, v *yaml.Node) (err error) {
		switch k.Value {
		case "name":
			return decodeScalar(v, "setting name", &wrap.res.Name)

		case "value":
			aux := ""
			err = decodeScalar(v, "setting value", &aux)
			if err != nil {
				return err
			}
			wrap.res.Value = sqlt.JSONText(aux)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if wrap.envoyConfig, err = decodeEnvoyConfig(n); err != nil {
		return
	}

	if wrap.ts, err = decodeTimestamps(n); err != nil {
		return
	}

	if wrap.us, err = decodeUserstamps(n); err != nil {
		return
	}

	return nil
}

func (wrap *setting) MarshalEnvoy() ([]resource.Interface, error) {
	rs := resource.NewSetting(wrap.res)
	rs.SetTimestamps(wrap.ts)
	rs.SetUserstamps(wrap.us)
	rs.SetConfig(wrap.envoyConfig)

	return envoy.CollectNodes(
		rs,
	)
}
