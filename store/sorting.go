package store

import (
	"fmt"
	"github.com/cortezaproject/corteza-server/pkg/slice"
	"regexp"
	"strings"
)

type (
	// Sort is a helper struct that should be embedded in filter types
	// to help with the sorting
	Sorting struct {
		Sort SortExprSet `json:"sort,omitempty"`
	}

	SortExpr struct {
		Column     string
		Descending bool
		// NullsFirst bool
	}

	SortExprSet []*SortExpr
)

func NewSorting(sort string) (s Sorting, err error) {
	s = Sorting{}

	if s.Sort, err = parseSort(sort); err != nil {
		return
	}

	return
}

// parses sort string
//
// We allow a simplified version of what SQL supports, so:
//   "<name>( <direction>), ..."
//
// Unlike before, we do not use pkg/ql for parsing this as we do not allow
// any complex sorting expressions
func parseSort(in string) (set SortExprSet, err error) {
	exprMatcher := regexp.MustCompile(`([0-9a-zA-Z_]+)(\s+(asc|ASC|desc|DESC))?`)

	set = SortExprSet{}

	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	for _, expr := range strings.Split(in, ",") {
		mm := exprMatcher.FindStringSubmatch(strings.TrimSpace(expr))

		o := &SortExpr{}
		switch {
		case len(mm) == 0:
			return nil, fmt.Errorf("invalid sort expression")
		case len(mm) >= 2:
			o.Column = mm[1]
			fallthrough
		case len(mm) >= 4:
			o.Descending = strings.ToUpper(mm[3]) == "DESC"
		}

		set = append(set, o)
	}

	return set, nil
}

// UnmarshalJSON parses stringified sort expression when passed inside JSON
func (set *SortExprSet) UnmarshalJSON(in []byte) error {
	tmp, err := parseSort(string(in))
	*set = tmp
	return err
}

// UnmarshalJSON parses stringified sort expression when passed inside JSON
func (set *SortExprSet) Set(in string) error {
	tmp, err := parseSort(in)
	*set = tmp
	return err
}

// Validate returns error if any of the SortExpr columns is missing from the given list
func (set SortExprSet) Validate(cc ...string) error {
	var valid = slice.ToStringBoolMap(cc)
	for _, c := range set {
		if !valid[c.Column] {
			return fmt.Errorf("invalid sort %q column userd", c.Column)
		}
	}

	return nil
}

// Clone returns cloned sort expression set
func (set SortExprSet) Clone() (out SortExprSet) {
	out = make([]*SortExpr, len(set))
	for i := range set {
		out[i] = &SortExpr{}
		*(out[i]) = *(set[i])
	}

	return out
}

// Reverse reverses direction on each expression
func (set SortExprSet) Reverse() {
	for i := range set {
		set[i].Descending = !set[i].Descending
	}
}

// Reverse reverses direction on each expression
func (set SortExprSet) Columns() []string {
	out := make([]string, len(set))
	for i := range set {
		out[i] = set[i].Column
	}

	return out
}

func (set SortExprSet) String() string {
	out := make([]string, len(set))
	for i := range set {
		out[i] = set[i].Column

		if set[i].Descending {
			out[i] += " DESC"
		}

	}

	return strings.Join(out, ", ")
}