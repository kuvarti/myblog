package controllers

import "testing"

func TestValidatePagePathFormat(t *testing.T) {
	cases := []struct {
		path string
		want pathValidity
	}{
		{"", pathBadFormat},
		{"about", pathBadFormat},
		{"/", pathOK},
		{"/about", pathOK},
		{"/panel", pathReserved},
		{"/lists", pathReserved},
	}
	for _, c := range cases {
		if got := validatePagePathFormat(c.path); got != c.want {
			t.Errorf("validatePagePathFormat(%q) = %d, want %d", c.path, got, c.want)
		}
	}
}
