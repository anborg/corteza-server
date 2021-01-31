package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/cortezaproject/corteza-server/pkg/envoy/resource"
	"github.com/cortezaproject/corteza-server/pkg/id"
)

var (
	// wrapper around NextID that will aid service testing
	NextID = func() uint64 {
		return id.Next()
	}

	// wrapper around time.Now() that will aid testing
	now = func() *time.Time {
		c := time.Now().Round(time.Second)
		return &c
	}
)

const (
	rbacSep = ":"
)

func RbacResToRef(rr string) (*resource.Ref, error) {
	if rr == "" {
		return nil, nil
	}

	ref := &resource.Ref{}

	rr = strings.TrimSpace(rr)
	rr = strings.TrimRight(rr, rbacSep)

	parts := strings.Split(rr, rbacSep)

	// When len is 1; only top-level defined (system, compose, messaging)
	if len(parts) == 1 {
		ref.ResourceType = rr
		return ref, nil
	}

	// When len is 2; top-level and sub level defined (compose:namespace, system:user, ...)
	if len(parts) == 2 {
		ref.ResourceType = rr + rbacSep
		return ref, nil
	}

	//When len is 3; both levels defined; resource ref also provided
	if len(parts) == 3 {
		ref.ResourceType = strings.Join(parts[0:2], rbacSep) + rbacSep
		if parts[2] != "*" {
			ref.Identifiers = resource.MakeIdentifiers(parts[2])
		}
		return ref, nil
	}

	return nil, fmt.Errorf("invalid resource provided: %s", rr)
}
