package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cortezaproject/corteza-server/compose/types"
	"github.com/cortezaproject/corteza-server/pkg/envoy"
	"github.com/cortezaproject/corteza-server/pkg/envoy/resource"
)

type (
	bulkComposeRecordEncoder struct {
		encoderConfig *EncoderConfig

		res *resource.ComposeRecord

		refNs string

		refMod string
		relMod *types.Module

		tse *resource.TimestampEncoder
	}
)

func bulkComposeRecordEncoderFromResource(rec *resource.ComposeRecord, cfg *EncoderConfig) *bulkComposeRecordEncoder {
	return &bulkComposeRecordEncoder{
		encoderConfig: cfg,

		res: rec,

		tse: resource.NewTimestampEncoder().WithTimezone(cfg.Timezone).WithTemplate(cfg.TimeLayout),
	}
}

func (n *bulkComposeRecordEncoder) Prepare(ctx context.Context, state *envoy.ResourceState) (err error) {
	_, ok := state.Res.(*resource.ComposeRecord)
	if !ok {
		return encoderErrInvalidResource(resource.COMPOSE_RECORD_RESOURCE_TYPE, state.Res.ResourceType())
	}

	return nil
}

func (n *bulkComposeRecordEncoder) Encode(ctx context.Context, w io.Writer, state *envoy.ResourceState) (err error) {
	enc := json.NewEncoder(w)

	err = n.res.Walker(func(r *resource.ComposeRecordRaw) error {
		m, err := makeMap(
			"id", r.ID,
		)

		ts, err := n.tse.EncodeTimestamps(r.Ts).End()
		if err != nil {
			return err
		}
		m, err = mapTimestamps(m, ts)

		m, err = addMap(m,
			"createdBy", n.res.UserFlakes.GetByStamp(r.Us.CreatedBy),
			"updatedBy", n.res.UserFlakes.GetByStamp(r.Us.UpdatedBy),
			"deletedBy", n.res.UserFlakes.GetByStamp(r.Us.DeletedBy),
			"ownedBy", n.res.UserFlakes.GetByStamp(r.Us.OwnedBy),
		)
		if err != nil {
			return err
		}

		for k, v := range r.Values {
			f := n.res.RelMod.Fields.FindByName(k)
			if f == nil {
				return fmt.Errorf("field %s not found", k)
			}
			if f.Kind == "User" {
				m, err = addMap(m,
					k, n.res.UserFlakes.GetByKey(v),
				)
			} else {
				m, err = addMap(m,
					k, v,
				)
			}
			if err != nil {
				return err
			}
		}

		return enc.Encode(m)
	})

	if err != nil {
		return err
	}
	return nil
}
