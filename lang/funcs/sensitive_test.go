package funcs

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestSensitive(t *testing.T) {
	tests := []struct {
		Input   cty.Value
		WantErr string
	}{
		{
			cty.NumberIntVal(1),
			``,
		},
		{
			// Unknown values stay unknown while becoming sensitive
			cty.UnknownVal(cty.String),
			``,
		},
		{
			// Null values stay unknown while becoming sensitive
			cty.NullVal(cty.String),
			``,
		},
		{
			// DynamicVal can be marked as sensitive
			cty.DynamicVal,
			``,
		},
		{
			// The marking is shallow only
			cty.ListVal([]cty.Value{cty.NumberIntVal(1)}),
			``,
		},
		{
			// A value already marked is allowed and stays marked
			cty.NumberIntVal(1).Mark("sensitive"),
			``,
		},
		{
			// A value deep already marked is allowed and stays marked,
			// _and_ we'll also mark the outer collection as sensitive.
			cty.ListVal([]cty.Value{cty.NumberIntVal(1).Mark("sensitive")}),
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sensitive(%#v)", test.Input), func(t *testing.T) {
			got, err := Sensitive(test.Input)

			if test.WantErr != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.WantErr; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.HasMark("sensitive") {
				t.Errorf("result is not marked sensitive")
			}
			gotRaw, _ := got.Unmark()
			if !gotRaw.RawEquals(test.Input) {
				t.Errorf("wrong unmarked result\ngot:  %#v\nwant: %#v", got, test.Input)
			}
		})
	}
}

func TestNonsensitive(t *testing.T) {
	tests := []struct {
		Input   cty.Value
		WantErr string
	}{
		{
			cty.NumberIntVal(1).Mark("sensitive"),
			``,
		},
		{
			cty.DynamicVal.Mark("sensitive"),
			``,
		},
		{
			cty.UnknownVal(cty.String).Mark("sensitive"),
			``,
		},
		{
			cty.NullVal(cty.EmptyObject).Mark("sensitive"),
			``,
		},
		{
			// The inner sensitive remains afterwards
			cty.ListVal([]cty.Value{cty.NumberIntVal(1).Mark("sensitive")}).Mark("sensitive"),
			``,
		},

		// Passing a value that is already non-sensitive is an error,
		// because this function should always be used with specific
		// intention, not just as a "make everything visible" hammer.
		{
			cty.NumberIntVal(1),
			`the given value is not sensitive, so this call is redundant`,
		},
		{
			cty.DynamicVal,
			`the given value is not sensitive, so this call is redundant`,
		},
		{
			cty.NullVal(cty.String),
			`the given value is not sensitive, so this call is redundant`,
		},
		{
			cty.UnknownVal(cty.String),
			`the given value is not sensitive, so this call is redundant`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("nonsensitive(%#v)", test.Input), func(t *testing.T) {
			got, err := Nonsensitive(test.Input)

			if test.WantErr != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.WantErr; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if got.HasMark("sensitive") {
				t.Errorf("result is still marked sensitive")
			}
			wantRaw, _ := test.Input.Unmark()
			if !got.RawEquals(wantRaw) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Input)
			}
		})
	}
}
