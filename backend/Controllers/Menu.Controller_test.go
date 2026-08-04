package controllers

import (
	models "backend/Models"
	"testing"
)

func TestBuildNavFiltersHidden(t *testing.T) {
	pages := []models.PageSummary{
		{PageName: "A", Visible: true},
		{PageName: "B", Visible: false},
		{PageName: "C", Visible: true},
	}
	nav := buildNav(pages, nil)
	if len(nav) != 2 {
		t.Fatalf("expected 2 visible entries, got %d", len(nav))
	}
	if nav[0].PageName != "A" || nav[1].PageName != "C" {
		t.Fatalf("expected A,C got %s,%s", nav[0].PageName, nav[1].PageName)
	}
}

func TestBuildNavDropsStubs(t *testing.T) {
	pages := []models.PageSummary{{PageName: "A", Visible: true}}
	menus := []*models.MenuModel{
		{PageName: "A", Caption: "Alpha"},
		{PageName: "", Caption: "Stub", Path: "/lists"}, // page-less seed stub
	}
	nav := buildNav(pages, menus)
	if len(nav) != 1 || nav[0].PageName != "A" {
		t.Fatalf("expected only page A, got %d entries", len(nav))
	}
}

func TestBuildNavCaptionFallback(t *testing.T) {
	pages := []models.PageSummary{{PageName: "Solo", Visible: true}}
	nav := buildNav(pages, nil)
	if len(nav) != 1 || nav[0].Caption != "Solo" {
		t.Fatalf("expected caption fallback 'Solo', got %+v", nav)
	}
}

func TestBuildNavKeepsMenuCaption(t *testing.T) {
	pages := []models.PageSummary{{PageName: "A", Path: "/a", Visible: true}}
	menus := []*models.MenuModel{{PageName: "A", Caption: "Alpha", Path: "/a"}}
	nav := buildNav(pages, menus)
	if nav[0].Caption != "Alpha" || nav[0].Path != "/a" {
		t.Fatalf("expected menu caption and page path, got %+v", nav[0])
	}
}

func TestBuildNavPathFromPage(t *testing.T) {
	pages := []models.PageSummary{{PageName: "A", Path: "/alpha", Visible: true}}
	menus := []*models.MenuModel{{PageName: "A", Caption: "Alpha", Path: "/stale"}}
	nav := buildNav(pages, menus)
	if nav[0].Path != "/alpha" {
		t.Fatalf("expected nav path from page '/alpha', got %q", nav[0].Path)
	}
	if nav[0].Caption != "Alpha" {
		t.Fatalf("expected caption 'Alpha', got %q", nav[0].Caption)
	}
}

func TestBuildNavPreservesOrder(t *testing.T) {
	pages := []models.PageSummary{
		{PageName: "First", Visible: true},
		{PageName: "Second", Visible: true},
		{PageName: "Third", Visible: true},
	}
	nav := buildNav(pages, nil)
	got := []string{nav[0].PageName, nav[1].PageName, nav[2].PageName}
	want := []string{"First", "Second", "Third"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %v want %v", i, got, want)
		}
	}
}
