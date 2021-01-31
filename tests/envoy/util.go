package envoy

import (
	"context"
	"testing"
	"time"

	"github.com/cortezaproject/corteza-server/store"
)

var (
	createdAt, _   = time.Parse(time.RFC3339, "2021-01-01T11:10:09Z")
	updatedAt, _   = time.Parse(time.RFC3339, "2021-01-02T11:10:09Z")
	suspendedAt, _ = time.Parse(time.RFC3339, "2021-01-03T11:10:09Z")
)

func truncateStore(ctx context.Context, s store.Storer, t *testing.T) {
	err := ce(
		s.TruncateComposeNamespaces(ctx),
		s.TruncateComposeModules(ctx),
		s.TruncateComposeModuleFields(ctx),
		s.TruncateComposeRecords(ctx, nil),
		s.TruncateComposePages(ctx),
		s.TruncateComposeCharts(ctx),

		s.TruncateMessagingChannels(ctx),

		s.TruncateRoles(ctx),
		s.TruncateUsers(ctx),
		s.TruncateApplications(ctx),
		s.TruncateSettings(ctx),
		s.TruncateRbacRules(ctx),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func truncateStoreRecords(ctx context.Context, s store.Storer, t *testing.T) {
	err := ce(
		s.TruncateComposeRecords(ctx, nil),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func parseTime(t *testing.T, ts string) *time.Time {
	tt, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err.Error())
	}
	return &tt
}
