package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cortezaproject/corteza-server/system/types"
)

type (
	base struct {
		rt string
		ii Identifiers
		rr RefSet

		ts    *Timestamps
		us    *Userstamps
		cfg   *EnvoyConfig
		urefs RefSet
	}

	EnvoyConfig struct {
		// SkipIf determines when the encoding should be skipped for this resource
		SkipIf     string
		OnExisting MergeAlg
	}

	Timestamp struct {
		S string
		T *time.Time
	}
	Timestamps struct {
		CreatedAt   *Timestamp
		UpdatedAt   *Timestamp
		DeletedAt   *Timestamp
		ArchivedAt  *Timestamp
		SuspendedAt *Timestamp
	}

	Userstamp struct {
		UserID uint64
		Ref    string
		U      *types.User

		// Formatted user representation; if not set one is generated
		formatted string
	}
	Userstamps struct {
		CreatedBy *Userstamp
		UpdatedBy *Userstamp
		DeletedBy *Userstamp
		OwnedBy   *Userstamp
	}
	UserstampIndex map[string]*Userstamp

	UserstampFormatter struct {
		tpl string
	}

	TimestampEncoder struct {
		tpl string
		tz  string
		tzL *time.Location
		ts  *Timestamps

		err error
	}

	MergeAlg int
)

const (
	// Default takes the operation defined default
	Default MergeAlg = iota
	// Skip skips the existing resource
	Skip
	// Replace replaces the existing resource
	Replace
	// MergeLeft updates the existing resource, giving priority to the existing data
	MergeLeft
	// MergeRight updates the existing resource, giving priority to the new data
	MergeRight
)

// State management methods

// AddIdentifier adds a set of identifiers to the current resource
func (t *base) AddIdentifier(ss ...string) {
	if t.ii == nil {
		t.ii = make(Identifiers)
	}

	t.ii.Add(ss...)
}

// AddRef adds a new reference to the current resource
func (t *base) AddRef(rt string, ii ...string) *Ref {
	if t.rr == nil {
		t.rr = make(RefSet, 0, 10)
	}

	iiC := make([]string, 0, len(ii))
	for _, i := range ii {
		if i != "" {
			iiC = append(iiC, i)
		}
	}

	ref := &Ref{ResourceType: rt, Identifiers: Identifiers{}.Add(iiC...)}
	t.rr = append(t.rr, ref)

	return ref
}

// SetResourceType sets the resource type of the current resource struct
func (t *base) SetResourceType(rt string) {
	t.rt = rt
}

func (t *base) SetTimestamps(ts *Timestamps) {
	t.ts = ts
}
func (t *base) Timestamps() *Timestamps {
	return t.ts
}

func (t *base) SetUserstamps(us *Userstamps) {
	t.us = us

	if us != nil {
		uu := []*Userstamp{us.CreatedBy, us.UpdatedBy, us.DeletedBy, us.OwnedBy}
		t.SetUserRefs(uu)
	}
}
func (t *base) Userstamps() *Userstamps {
	return t.us
}

func (t *base) SetConfig(cfg *EnvoyConfig) {
	t.cfg = cfg
}
func (t *base) Config() *EnvoyConfig {
	return t.cfg
}

func (t *base) SetUserRefs(uu []*Userstamp) {
	if t.urefs == nil {
		t.urefs = make(RefSet, 0, 4)
	}

	for _, u := range uu {
		if u == nil {
			continue
		}
		if u.UserID > 0 {
			t.urefs = append(t.urefs, t.AddRef(USER_RESOURCE_TYPE, strconv.FormatUint(u.UserID, 10)))
		} else if u.Ref != "" {
			t.urefs = append(t.urefs, t.AddRef(USER_RESOURCE_TYPE, u.Ref))
		}
	}
}
func (t *base) UserRefs() RefSet {
	return t.urefs
}

func (t *base) Identifiers() Identifiers {
	return t.ii
}
func (t *base) ResourceType() string {
	return t.rt
}
func (t *base) Refs() RefSet {
	return t.rr
}
func (t *base) HasRefs() bool {
	return t.rr == nil || len(t.rr) == 0
}

func (tt *Timestamps) Model(tpl string, tz string) (*Timestamps, error) {
	var err error
	if tt.CreatedAt != nil {
		tt.CreatedAt, err = tt.CreatedAt.Model(tpl, tz)
	}
	if tt.UpdatedAt != nil {
		tt.UpdatedAt, err = tt.UpdatedAt.Model(tpl, tz)
	}
	if tt.DeletedAt != nil {
		tt.DeletedAt, err = tt.DeletedAt.Model(tpl, tz)
	}
	if tt.ArchivedAt != nil {
		tt.ArchivedAt, err = tt.ArchivedAt.Model(tpl, tz)
	}
	if tt.SuspendedAt != nil {
		tt.SuspendedAt, err = tt.SuspendedAt.Model(tpl, tz)
	}

	return tt, err
}

func (ts *Timestamp) Model(tpl string, tz string) (*Timestamp, error) {
	if tz != "" {
		tzL, err := time.LoadLocation(tz)
		if err != nil {
			return nil, err
		}
		ts.T.In(tzL)
	}

	if tpl == "" {
		tpl = time.RFC3339
	}

	ts.S = ts.T.Format(tpl)
	return ts, nil
}

func NewTimestampEncoder() *TimestampEncoder {
	return &TimestampEncoder{
		ts: &Timestamps{},
	}
}

func (tse *TimestampEncoder) WithTimezone(tz string) *TimestampEncoder {
	if tz == "" {
		tz = "UTC"
	}

	var err error
	tse.tz = tz
	tse.tzL, err = time.LoadLocation(tz)
	if err != nil {
		tse.err = err
	}

	return tse
}

func (tse *TimestampEncoder) WithTemplate(tpl string) *TimestampEncoder {
	if tpl == "" {
		return tse
	}

	tse.tpl = tpl
	return tse
}

func (tse *TimestampEncoder) EncodeCUDA(c time.Time, u, d, a *time.Time) *TimestampEncoder {
	return tse.Encode("createdAt", &c).
		Encode("updatedAt", u).
		Encode("deletedAt", d).
		Encode("archivedAt", a)
}

func (tse *TimestampEncoder) EncodeTimestamps(ts *Timestamps) *TimestampEncoder {
	if ts.CreatedAt != nil {
		tse.Encode("createdAt", ts.CreatedAt.T)
	}
	if ts.UpdatedAt != nil {
		tse.Encode("updatedAt", ts.UpdatedAt.T)
	}
	if ts.DeletedAt != nil {
		tse.Encode("deletedAt", ts.DeletedAt.T)
	}
	if ts.ArchivedAt != nil {
		tse.Encode("archivedAt", ts.ArchivedAt.T)
	}
	if ts.SuspendedAt != nil {
		tse.Encode("suspendedAt", ts.SuspendedAt.T)
	}

	return tse
}

func (tse *TimestampEncoder) Encode(name string, t *time.Time) *TimestampEncoder {
	if tse.err != nil {
		return tse
	}

	if t == nil {
		return tse
	}

	tt := *t
	t = &tt

	if tse.tz != "" {
		loc, err := time.LoadLocation(tse.tz)
		if err != nil {
			tse.err = err
		}
		t.In(loc)
	}

	tpl := tse.tpl
	if tpl == "" {
		tpl = time.RFC3339
	}

	tf := t.Format(tpl)

	switch strings.ToLower(name) {
	case "createdat":
		tse.ts.CreatedAt = &Timestamp{
			S: tf,
			T: t,
		}
	case "updatedat":
		tse.ts.UpdatedAt = &Timestamp{
			S: tf,
			T: t,
		}
	case "deletedat":
		tse.ts.DeletedAt = &Timestamp{
			S: tf,
			T: t,
		}
	case "archivedat":
		tse.ts.ArchivedAt = &Timestamp{
			S: tf,
			T: t,
		}
	case "suspendedat":
		tse.ts.SuspendedAt = &Timestamp{
			S: tf,
			T: t,
		}
	default:
		tse.err = fmt.Errorf("unknown timestamp %v", name)
	}

	return tse
}

func (tse *TimestampEncoder) End() (*Timestamps, error) {
	if tse.err != nil {
		return nil, tse.err
	}

	return tse.ts, nil
}

func (ts *Timestamp) MarshalYAML() (interface{}, error) {
	return ts.S, nil
}

func (ts *Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(ts.S)
}

func (us *Userstamp) Stringify() (string, error) {
	if us == nil {
		return "", nil
	}

	if us.U != nil {
		if us.U.Handle != "" {
			return us.U.Handle, nil
		}
		if us.U.Username != "" {
			return us.U.Username, nil
		}
		if us.U.Email != "" {
			return us.U.Email, nil
		}
		if us.U.Name != "" {
			return us.U.Name, nil
		}
	}

	if us.Ref != "" {
		return us.Ref, nil
	}

	if us.UserID > 0 {
		return strconv.FormatUint(us.UserID, 10), nil
	}

	return "", errors.New("invalid userstamp")
}

func (us *Userstamp) MarshalYAML() (interface{}, error) {
	if us == nil {
		return nil, nil
	}

	if us.U != nil {
		if us.U.Handle != "" {
			return us.U.Handle, nil
		}
		if us.U.Username != "" {
			return us.U.Username, nil
		}
		if us.U.Email != "" {
			return us.U.Email, nil
		}
		if us.U.Name != "" {
			return us.U.Name, nil
		}
	}

	if us.Ref != "" {
		return us.Ref, nil
	}

	if us.UserID > 0 {
		return us.UserID, nil
	}

	return nil, errors.New("invalid userstamp")
}

func (us *Userstamp) MarshalJSON() ([]byte, error) {
	if us == nil {
		return nil, nil
	}

	l := ""

	if us.U != nil {
		if us.U.Handle != "" {
			l = us.U.Handle
		}
		if us.U.Username != "" {
			l = us.U.Username
		}
		if us.U.Email != "" {
			l = us.U.Email
		}
		if us.U.Name != "" {
			l = us.U.Name
		}
	} else {
		if us.Ref != "" {
			l = us.Ref
		}

		if us.UserID > 0 {
			l = strconv.FormatUint(us.UserID, 10)
		}
	}

	if l == "" {
		return nil, errors.New("invalid userstamp")
	}

	return json.Marshal(l)
}

func MakeTimestamp(ts string) *Timestamp {
	t := toTime(ts)
	if t == nil {
		return nil
	}

	return &Timestamp{
		S: ts,
		T: t,
	}
}

func MakeUserstamp(u *types.User) *Userstamp {
	sID := strconv.FormatUint(u.ID, 10)
	return &Userstamp{
		UserID: u.ID,
		U:      u,
		Ref:    FirstOkString(u.Handle, u.Email, u.Username, sID),
	}
}

func MakeUserstampFromRef(ref string) *Userstamp {
	id, err := strconv.ParseUint(ref, 10, 64)

	us := &Userstamp{}

	if err == nil && id != 0 {
		us.UserID = id
		us.U = &types.User{ID: id}
	}
	us.Ref = ref

	return us
}

func (ux UserstampIndex) Add(uu ...*types.User) {
	for _, u := range uu {
		sID := strconv.FormatUint(u.ID, 10)
		s := MakeUserstamp(u)

		ux[sID] = s
		ux[u.Email] = s
		if u.Handle != "" {
			ux[u.Handle] = s
		}
		if u.Username != "" {
			ux[u.Username] = s
		}
		if u.Name != "" {
			ux[u.Name+" "+u.Email] = s
		}
	}
}

func (ux UserstampIndex) GetByKey(kr interface{}) *Userstamp {
	if k, ok := kr.(string); ok {
		return ux[k]
	} else if k, ok := kr.(uint64); ok {
		return ux[strconv.FormatUint(k, 10)]
	}
	return nil
}

func (ux UserstampIndex) GetByStamp(s *Userstamp) *Userstamp {
	if s == nil {
		return nil
	}

	if s.Ref != "" {
		return ux.GetByKey(s.Ref)
	}
	if s.UserID > 0 {
		return ux.GetByKey(s.UserID)
	}
	if s.U != nil && s.U.ID > 0 {
		return ux.GetByKey(s.U.ID)
	}
	return s
}
