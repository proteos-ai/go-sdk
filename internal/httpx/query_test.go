package httpx

import (
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func parseValues(t *testing.T, raw string) url.Values {
	t.Helper()
	v, err := url.ParseQuery(raw)
	require.NoError(t, err)
	return v
}

func TestToQueryParams_Primitives(t *testing.T) {
	type opts struct {
		Page     int    `query:"page"`
		Name     string `query:"name"`
		Verified bool   `query:"verified"`
	}
	got := parseValues(t, ToQueryParams(opts{Page: 2, Name: "alice", Verified: true}))
	require.Equal(t, "2", got.Get("page"))
	require.Equal(t, "alice", got.Get("name"))
	require.Equal(t, "true", got.Get("verified"))
}

func TestToQueryParams_FloatAndUint(t *testing.T) {
	type opts struct {
		Score float64 `query:"score"`
		ID    uint64  `query:"id"`
	}
	got := parseValues(t, ToQueryParams(opts{Score: 1.5, ID: 9999}))
	require.Equal(t, "1.5", got.Get("score"))
	require.Equal(t, "9999", got.Get("id"))
}

func TestToQueryParams_OmitemptySkipsZero(t *testing.T) {
	type opts struct {
		Page int    `query:"page"`
		Sort string `query:"sort,omitempty"`
	}
	got := parseValues(t, ToQueryParams(opts{Page: 0}))
	require.Equal(t, "0", got.Get("page"), "no omitempty -> always emitted")
	require.Empty(t, got.Get("sort"))

	got = parseValues(t, ToQueryParams(opts{Page: 0, Sort: "name"}))
	require.Equal(t, "name", got.Get("sort"))
}

func TestToQueryParams_BoolEncoding(t *testing.T) {
	type opts struct {
		A bool `query:"a"`
		B bool `query:"b"`
	}
	got := parseValues(t, ToQueryParams(opts{A: true, B: false}))
	require.Equal(t, "true", got.Get("a"))
	require.Equal(t, "false", got.Get("b"))
}

func TestToQueryParams_SliceRepeats(t *testing.T) {
	type opts struct {
		Tags []string `query:"tag"`
	}
	got := parseValues(t, ToQueryParams(opts{Tags: []string{"a", "b", "c"}}))
	values := got["tag"]
	sort.Strings(values)
	require.Equal(t, []string{"a", "b", "c"}, values)
}

func TestToQueryParams_SkipsNestedStructs(t *testing.T) {
	type nested struct {
		Inner string `query:"inner"`
	}
	type opts struct {
		Page int    `query:"page"`
		N    nested `query:"n"`
	}
	got := ToQueryParams(opts{Page: 1, N: nested{Inner: "x"}})
	require.NotContains(t, got, "n=")
	require.NotContains(t, got, "inner")
	require.Contains(t, got, "page=1")
}

func TestToQueryParams_EmbeddedStructFlattens(t *testing.T) {
	type Base struct {
		Page     int `query:"page"`
		PageSize int `query:"page_size"`
	}
	type opts struct {
		Base
		Name string `query:"name"`
	}
	got := parseValues(t, ToQueryParams(opts{Base: Base{Page: 1, PageSize: 50}, Name: "alice"}))
	require.Equal(t, "1", got.Get("page"))
	require.Equal(t, "50", got.Get("page_size"))
	require.Equal(t, "alice", got.Get("name"))
}

func TestToQueryParams_FlattenMap(t *testing.T) {
	type opts struct {
		Page    int            `query:"page"`
		Filters map[string]any `query:",flatten"`
	}
	got := parseValues(t, ToQueryParams(opts{
		Page: 0,
		Filters: map[string]any{
			"name[eq]": "alice",
			"age[gt]":  21,
			"tags":     []string{"x", "y"},
			"meta":     map[string]any{"deep": 1}, // nested -> skipped
			"verified": true,
		},
	}))
	require.Equal(t, "alice", got.Get("name[eq]"))
	require.Equal(t, "21", got.Get("age[gt]"))
	require.Equal(t, "true", got.Get("verified"))
	tags := got["tags"]
	sort.Strings(tags)
	require.Equal(t, []string{"x", "y"}, tags)
	require.Empty(t, got.Get("meta"))
}

func TestToQueryParams_DashTagSkips(t *testing.T) {
	type opts struct {
		Visible string `query:"visible"`
		Secret  string `query:"-"`
	}
	got := ToQueryParams(opts{Visible: "yes", Secret: "no"})
	require.Contains(t, got, "visible=yes")
	require.NotContains(t, got, "secret")
}

func TestToQueryParams_FallsBackToJSONTagThenLowerCamel(t *testing.T) {
	type opts struct {
		Page  int    `json:"page"`
		Email string `json:"email,omitempty"`
		Plain string // no tag -> lowerCamel of "Plain"
	}
	got := parseValues(t, ToQueryParams(opts{Page: 5, Plain: "x"}))
	require.Equal(t, "5", got.Get("page"))
	require.Equal(t, "x", got.Get("plain"))
	require.Empty(t, got.Get("email"), "json omitempty should be honored")
}

func TestToQueryParams_PointerStruct(t *testing.T) {
	type opts struct {
		Page int `query:"page"`
	}
	got := parseValues(t, ToQueryParams(&opts{Page: 7}))
	require.Equal(t, "7", got.Get("page"))
}

func TestToQueryParams_NilInputReturnsEmpty(t *testing.T) {
	require.Empty(t, ToQueryParams(nil))
	var p *struct{ X int }
	require.Empty(t, ToQueryParams(p))
}

func TestToQueryParams_DirectMap(t *testing.T) {
	got := parseValues(t, ToQueryParams(map[string]any{
		"a": 1,
		"b": "two",
	}))
	require.Equal(t, "1", got.Get("a"))
	require.Equal(t, "two", got.Get("b"))
}

func TestToQueryParams_UnexportedFieldsIgnored(t *testing.T) {
	type opts struct {
		Page   int `query:"page"`
		hidden int //nolint:unused
	}
	got := ToQueryParams(opts{Page: 1})
	require.Equal(t, "page=1", got)
}

func TestToQueryParams_PointerFieldNilSkipped(t *testing.T) {
	type opts struct {
		Page *int `query:"page,omitempty"`
	}
	require.Empty(t, ToQueryParams(opts{}))
	v := 5
	got := ToQueryParams(opts{Page: &v})
	require.True(t, strings.Contains(got, "page=5"))
}
