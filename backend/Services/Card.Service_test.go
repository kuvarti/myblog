package services

import (
	models "backend/Models"
	"strings"
	"testing"
)

func TestCardTitleFallsBackToPageName(t *testing.T) {
	if got := cardTitle("Alpha", "A"); got != "Alpha" {
		t.Fatalf("want caption, got %q", got)
	}
	if got := cardTitle("", "A"); got != "A" {
		t.Fatalf("want PageName fallback, got %q", got)
	}
}

func TestExtractSummarySkipsHeadingsAndHTML(t *testing.T) {
	src := "# Title\n<img src=\"x.jpg\">\nThe first real paragraph."
	if got := extractSummary(src); got != "The first real paragraph." {
		t.Fatalf("got %q", got)
	}
}

func TestExtractSummaryStripsMarkersAndEmphasis(t *testing.T) {
	if got := extractSummary("- **Bold** item"); got != "Bold item" {
		t.Fatalf("got %q", got)
	}
	// A leading number that is part of the text must survive.
	if got := extractSummary("2024 was a good year"); got != "2024 was a good year" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractSummaryTruncates(t *testing.T) {
	long := strings.Repeat("a", 200)
	got := extractSummary(long)
	if len([]rune(got)) != 161 || !strings.HasSuffix(got, "…") {
		t.Fatalf("want 160 runes + ellipsis, got %d runes", len([]rune(got)))
	}
}

func TestExtractImagePrefersHTMLThenMarkdown(t *testing.T) {
	if got := extractImage("intro\n<img alt=\"\" src=\"/a.jpg\">"); got != "/a.jpg" {
		t.Fatalf("html img: got %q", got)
	}
	if got := extractImage("intro\n![alt](/b.png \"t\")"); got != "/b.png" {
		t.Fatalf("md img: got %q", got)
	}
	if got := extractImage("no images here"); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestBuildCardHTML(t *testing.T) {
	withImg := buildCardHTML("/p", "Title", "Sum", "/i.jpg")
	if !strings.Contains(withImg, `href="/p"`) ||
		!strings.Contains(withImg, `src="/i.jpg"`) ||
		!strings.Contains(withImg, "Title") || !strings.Contains(withImg, "Sum") {
		t.Fatalf("missing parts: %q", withImg)
	}
	noImg := buildCardHTML("/p", "T", "S", "")
	if strings.Contains(noImg, "<img") {
		t.Fatalf("expected no <img>, got %q", noImg)
	}
}

func TestSelectByTagsOrMatchAndSelfExclude(t *testing.T) {
	pages := []models.PageModel{
		{PageName: "self", Tags: []string{"blog"}},
		{PageName: "A", Tags: []string{"blog", "go"}},
		{PageName: "B", Tags: []string{"go"}},
		{PageName: "C", Tags: []string{"news"}},
	}
	got := selectByTags(pages, []string{"blog", "go"}, "self")
	if len(got) != 2 || got[0].PageName != "A" || got[1].PageName != "B" {
		t.Fatalf("want A,B got %+v", got)
	}
	if len(selectByTags(pages, nil, "self")) != 0 {
		t.Fatalf("empty listTags must match nothing")
	}
}

func TestExpandShortcodesReplacesAndDrops(t *testing.T) {
	resolve := func(path string) (string, bool) {
		if path == "/ok" {
			return "<CARD>", true
		}
		return "", false
	}
	in := `intro <card path="/ok"> mid <card path="/missing"> end`
	got := expandShortcodes(in, resolve)
	if !strings.Contains(got, "<CARD>") || strings.Contains(got, "/missing") ||
		strings.Contains(got, "<card") {
		t.Fatalf("got %q", got)
	}
}

func TestExtractSummaryEmptySource(t *testing.T) {
	if got := extractSummary(""); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestExtractSummaryJoinsBlockLines(t *testing.T) {
	if got := extractSummary("A short lead.\n…continued detail…"); got != "A short lead. …continued detail…" {
		t.Fatalf("got %q", got)
	}
}

func TestExpandShortcodesLeavesPlainHTMLUnchanged(t *testing.T) {
	in := "<p>no shortcodes here</p>"
	got := expandShortcodes(in, func(string) (string, bool) { return "X", true })
	if got != in {
		t.Fatalf("got %q", got)
	}
}

func TestSelectByTagsPreservesIncomingOrder(t *testing.T) {
	// Out-of-order input (not sorted by any field) must be returned in the same order.
	pages := []models.PageModel{
		{PageName: "B", Tags: []string{"go"}},
		{PageName: "A", Tags: []string{"go"}},
	}
	got := selectByTags(pages, []string{"go"}, "")
	if len(got) != 2 || got[0].PageName != "B" || got[1].PageName != "A" {
		t.Fatalf("want B,A (incoming order), got %+v", got)
	}
}
