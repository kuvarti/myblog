package controllers

import (
	models "backend/Models"
	"testing"
)

func names(menus []*models.MenuModel) []string {
	out := make([]string, len(menus))
	for i, m := range menus {
		out[i] = m.PageName
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSortMenusByOrder(t *testing.T) {
	menus := []*models.MenuModel{
		{PageName: "a"}, {PageName: "b"}, {PageName: "c"},
	}
	got := names(sortMenusByOrder(menus, map[string]int{"a": 2, "b": 0, "c": 1}))
	if want := []string{"b", "c", "a"}; !equal(got, want) {
		t.Fatalf("ordered: got %v want %v", got, want)
	}
}

func TestSortMenusMissingSortLast(t *testing.T) {
	// "x" and "z" have no page order; they must sort after ordered entries and
	// keep their relative order.
	menus := []*models.MenuModel{
		{PageName: "x"}, {PageName: "a"}, {PageName: "z"}, {PageName: "b"},
	}
	got := names(sortMenusByOrder(menus, map[string]int{"a": 0, "b": 1}))
	if want := []string{"a", "b", "x", "z"}; !equal(got, want) {
		t.Fatalf("missing-last: got %v want %v", got, want)
	}
}

func TestSortMenusStableTies(t *testing.T) {
	// Equal orders preserve input order.
	menus := []*models.MenuModel{
		{PageName: "a"}, {PageName: "b"}, {PageName: "c"},
	}
	got := names(sortMenusByOrder(menus, map[string]int{"a": 0, "b": 0, "c": 0}))
	if want := []string{"a", "b", "c"}; !equal(got, want) {
		t.Fatalf("stable-ties: got %v want %v", got, want)
	}
}
