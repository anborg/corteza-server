package csv

import (
	"context"
	"encoding/csv"
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
	}
}

// Prepare prepares the composeRecord to be encoded
//
// Any validation, additional constraining should be performed here.
func (n *bulkComposeRecordEncoder) Prepare(ctx context.Context, state *envoy.ResourceState) (err error) {
	_, ok := state.Res.(*resource.ComposeRecord)
	if !ok {
		return encoderErrInvalidResource(resource.COMPOSE_RECORD_RESOURCE_TYPE, state.Res.ResourceType())
	}

	return nil
}

// Encode encodes the composeRecord to the document
//
// Encode is allowed to do some data manipulation, but no resource constraints
// should be changed.
func (n *bulkComposeRecordEncoder) Encode(ctx context.Context, w io.Writer, state *envoy.ResourceState) (err error) {
	enc := csv.NewWriter(w)

	// Generate header & cell index
	hh := make([]string, 0, 100)
	fx := make(map[string]int)
	// 1 sys fields
	hh = append(hh, "id", "ownedBy", "createdAt", "createdBy", "updatedAt", "updatedBy", "deletedAt", "deletedBy")
	offset := len(hh)

	// 2 modlue fields
	for i, f := range n.res.RelMod.Fields {
		hh = append(hh, f.Name)
		// Offset because of system values
		fx[f.Name] = offset + i
	}
	enc.Write(hh)

	err = n.res.Walker(func(r *resource.ComposeRecordRaw) error {

		row := make([]string, len(hh))

		var err error
		row[0] = r.ID
		if r.Us != nil {
			if r.Us.OwnedBy != nil {
				row[1], err = n.res.UserFlakes.GetByStamp(r.Us.OwnedBy).Stringify()
			}
			if r.Us.CreatedBy != nil {
				row[3], err = n.res.UserFlakes.GetByStamp(r.Us.CreatedBy).Stringify()
			}
			if r.Us.UpdatedBy != nil {
				row[5], err = n.res.UserFlakes.GetByStamp(r.Us.UpdatedBy).Stringify()
			}
			if r.Us.DeletedBy != nil {
				row[7], err = n.res.UserFlakes.GetByStamp(r.Us.DeletedBy).Stringify()
			}
		}

		if r.Ts != nil {
			r.Ts, err = r.Ts.Model(n.encoderConfig.TimeLayout, n.encoderConfig.Timezone)
			if err != nil {
				return err
			}
			if r.Ts.CreatedAt != nil {
				row[2] = r.Ts.CreatedAt.S
			}
			if r.Ts.UpdatedAt != nil {
				row[4] = r.Ts.UpdatedAt.S
			}
			if r.Ts.DeletedAt != nil {
				row[6] = r.Ts.DeletedAt.S
			}
		}
		if err != nil {
			return err
		}

		for k, v := range r.Values {
			cell, has := fx[k]
			if !has {
				return fmt.Errorf("unknown cell %s", k)
			}
			f := n.res.RelMod.Fields.FindByName(k)
			if f == nil {
				return fmt.Errorf("field %s not found", k)
			}

			if f.Kind == "User" {
				row[cell], err = n.res.UserFlakes.GetByKey(v).Stringify()
				if err != nil {
					return err
				}
			} else {
				row[cell] = v
			}
		}
		return enc.Write(row)
	})
	if err != nil {
		return err
	}

	enc.Flush()
	return nil
}
