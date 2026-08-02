package services

import (
	"strings"
	"testing"
)

func TestToStorageReplacesNewlines(t *testing.T) {
	if got := ToStorage("line1\nline2"); got != "line1/nline2" {
		t.Fatalf("ToStorage: got %q, want %q", got, "line1/nline2")
	}
}

func TestFromStorageReplacesDelimiter(t *testing.T) {
	if got := FromStorage("line1/nline2"); got != "line1\nline2" {
		t.Fatalf("FromStorage: got %q, want %q", got, "line1\nline2")
	}
}

func TestRoundTripPreservesNewlines(t *testing.T) {
	original := "# Title\n\nParagraph one\nParagraph two"
	if got := FromStorage(ToStorage(original)); got != original {
		t.Fatalf("round-trip: got %q, want %q", got, original)
	}
}

func TestPreviewRendersMarkdownHeading(t *testing.T) {
	psi := &PageServiceImplementation{}
	html, err := psi.Preview("# Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<h1") || !strings.Contains(html, "Hello") {
		t.Fatalf("expected an <h1> containing Hello, got %q", html)
	}
}

func TestGetPageTextHandlesMarkdownTerminatedInput(t *testing.T) {
	psi := &PageServiceImplementation{}
	// Regression: input whose final line is Markdown (no trailing "<" line or
	// blank separator) previously panicked with index out of range.
	out, err := psi.GetPageText("# Title/n/nBody paragraph")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<h1") || !strings.Contains(out, "Body paragraph") {
		t.Fatalf("expected rendered heading + body, got %q", out)
	}
}

func TestPreviewPassesThroughRawHTMLLine(t *testing.T) {
	psi := &PageServiceImplementation{}
	raw := `<div class="x">raw</div>`
	html, err := psi.Preview(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, raw) {
		t.Fatalf("expected raw HTML passthrough, got %q", html)
	}
}
