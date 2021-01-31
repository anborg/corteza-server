package envoy

import (
	"context"
	"testing"
	"time"

	"github.com/cortezaproject/corteza-server/compose/types"
	"github.com/cortezaproject/corteza-server/pkg/envoy"
	"github.com/cortezaproject/corteza-server/pkg/envoy/json"
	"github.com/cortezaproject/corteza-server/pkg/envoy/resource"
	su "github.com/cortezaproject/corteza-server/pkg/envoy/store"
	"github.com/cortezaproject/corteza-server/store"
	"github.com/stretchr/testify/require"
)

// TestStoreJsonl_records takes data from s1, encodes it into jsonl files, decodes
// created jsonl files, encodes into s2 and compares the data from s2.
func TestStoreJsonl_records(t *testing.T) {
	type (
		tc struct {
			name string
			// Before the data gets processed
			pre func(ctx context.Context, s store.Storer) (error, *su.DecodeFilter)
			// After the data gets processed
			postStoreDecode func(req *require.Assertions, err error)
			postJsonlEncode func(req *require.Assertions, err error)
			postStoreEncode func(req *require.Assertions, err error)
			// Data assertions
			check func(ctx context.Context, s store.Storer, req *require.Assertions)
		}
	)

	ctx := context.Background()
	s := initStoreT(ctx, t)

	ni := uint64(0)
	su.NextID = func() uint64 {
		ni++
		return ni
	}

	cases := []*tc{
		{
			name: "base record",
			pre: func(ctx context.Context, s store.Storer) (error, *su.DecodeFilter) {
				truncateStore(ctx, s, t)
				ns := sTestComposeNamespace(ctx, t, s, "base")
				mod := sTestComposeModule(ctx, t, s, ns.ID, "base")
				usr := sTestUser(ctx, t, s, "base")
				sTestComposeRecord(ctx, t, s, ns.ID, mod.ID, usr.ID)

				df := su.NewDecodeFilter().
					ComposeRecord(&types.RecordFilter{
						NamespaceID: ns.ID,
						ModuleID:    mod.ID,
					})
				return nil, df
			},
			check: func(ctx context.Context, s store.Storer, req *require.Assertions) {
				ns, err := store.LookupComposeNamespaceBySlug(ctx, s, "base_namespace")
				req.NoError(err)
				mod, err := store.LookupComposeModuleByNamespaceIDHandle(ctx, s, ns.ID, "base_module")
				req.NoError(err)
				usr, err := store.LookupUserByHandle(ctx, s, "base_user")
				req.NoError(err)

				rr, _, err := store.SearchComposeRecords(ctx, s, mod, types.RecordFilter{
					ModuleID:    mod.ID,
					NamespaceID: ns.ID,
				})
				req.NoError(err)
				req.Len(rr, 1)
				rec := rr[0]

				req.Equal(ns.ID, rec.NamespaceID)
				req.Equal(mod.ID, rec.ModuleID)

				req.Equal(createdAt.Format(time.RFC3339), rec.CreatedAt.Format(time.RFC3339))
				req.Equal(updatedAt.Format(time.RFC3339), rec.UpdatedAt.Format(time.RFC3339))
				req.Equal(usr.ID, rec.OwnedBy)
				req.Equal(usr.ID, rec.CreatedBy)
				req.Equal(usr.ID, rec.UpdatedBy)

				req.Len(rec.Values, 2)
				vv := rec.Values.FilterByName("module_field_string")
				req.Len(vv, 1)
				req.Equal("string value", vv[0].Value)

				vv = rec.Values.FilterByName("module_field_number")
				req.Len(vv, 1)
				req.Equal("10", vv[0].Value)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := require.New(t)

			err, df := c.pre(ctx, s)
			if err != nil {
				t.Fatal(err.Error())
			}
			// Decode from store
			sd := su.Decoder()
			nn, err := sd.Decode(ctx, s, df)
			if c.postStoreDecode != nil {
				c.postStoreDecode(req, err)
			} else {
				req.NoError(err)
			}

			// Encode into jsonl
			je := json.NewBulkRecordEncoder(&json.EncoderConfig{})
			bld := envoy.NewBuilder(je)
			g, err := bld.Build(ctx, nn...)
			req.NoError(err)
			err = envoy.Encode(ctx, g, je)
			ss := je.Stream()
			if c.postJsonlEncode != nil {
				c.postJsonlEncode(req, err)
			} else {
				req.NoError(err)
			}

			// Cleanup the store
			truncateStoreRecords(ctx, s, t)

			// Encode back into store
			se := su.NewStoreEncoder(s, &su.EncoderConfig{})
			jd := json.Decoder()
			nn = make([]resource.Interface, 0, len(nn))
			for _, s := range ss {
				mm, err := jd.Decode(ctx, s.Source, &envoy.DecoderOpts{
					Name: "tmp.jsonl",
					Path: "/tmp.jsonl",
				})
				req.NoError(err)
				nn = append(nn, mm...)
			}

			tpl := resource.NewComposeRecordTemplate(
				"base_module",
				"base_namespace",
				"tmp.jsonl",
				resource.MappingTplSet{},
			)

			nn = append(nn, tpl)
			crs := resource.ComposeRecordShaper()
			nn, err = resource.Shape(nn, crs)
			req.NoError(err)
			bld = envoy.NewBuilder(se)
			g, err = bld.Build(ctx, nn...)
			req.NoError(err)

			err = envoy.Encode(ctx, g, se)
			if c.postStoreEncode != nil {
				c.postStoreEncode(req, err)
			} else {
				req.NoError(err)
			}

			// Assert
			c.check(ctx, s, req)

			// Cleanup the store
			truncateStoreRecords(ctx, s, t)
		})
		ni = 0
	}
}
