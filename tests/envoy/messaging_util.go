package envoy

import (
	"context"
	"testing"

	"github.com/cortezaproject/corteza-server/messaging/types"
	su "github.com/cortezaproject/corteza-server/pkg/envoy/store"
	"github.com/cortezaproject/corteza-server/store"
)

func sTestChannel(ctx context.Context, t *testing.T, s store.Storer, usrID uint64, pfx string) *types.Channel {
	ch := &types.Channel{
		ID:    su.NextID(),
		Name:  pfx + "_channel",
		Topic: "topic",
		Type:  types.ChannelTypeGroup,

		CreatorID: usrID,

		CreatedAt: createdAt,
		UpdatedAt: &updatedAt,
	}

	err := store.CreateMessagingChannel(ctx, s, ch)
	if err != nil {
		t.Fatal(err)
	}

	return ch
}
