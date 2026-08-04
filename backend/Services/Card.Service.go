package services

import (
	models "backend/Models"
	"html"
	"regexp"
	"strings"
)

const cardSummaryMax = 160

var (
	cardShortcodeRe = regexp.MustCompile(`<card\s+path="([^"]*)"\s*/?>`)
	htmlImgRe       = regexp.MustCompile(`<img[^>]*\ssrc="([^"]*)"`)
	mdImgRe         = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)`)
	listMarkerRe    = regexp.MustCompile(`^\s*([-*+]\s+|\d+\.\s+|>\s+)`)
	emphasisStrip   = strings.NewReplacer("*", "", "_", "", "`", "")
)

// cardTitle prefers the page's menu caption, falling back to its PageName.
func cardTitle(caption, pageName string) string {
	if strings.TrimSpace(caption) != "" {
		return caption
	}
	return pageName
}

// extractSummary returns the first real paragraph of a page's (clean-newline)
// source as a single block: consecutive non-blank, non-raw-HTML lines starting
// at the first line that is not a heading or a Markdown image, with list markers
// and emphasis characters stripped, truncated to cardSummaryMax runes.
func extractSummary(source string) string {
	var block []string
	for _, raw := range strings.Split(source, "\n") {
		line := strings.TrimSpace(raw)
		if len(block) == 0 {
			// Seeking the block's first line: skip blanks, raw HTML, headings, images.
			if line == "" || strings.Contains(line, "<") ||
				strings.HasPrefix(line, "#") || strings.HasPrefix(line, "![") {
				continue
			}
			block = append(block, line)
			continue
		}
		// Inside the block: a blank or raw-HTML line ends it.
		if line == "" || strings.Contains(line, "<") {
			break
		}
		block = append(block, line)
	}
	if len(block) == 0 {
		return ""
	}
	text := listMarkerRe.ReplaceAllString(strings.Join(block, " "), "")
	text = strings.TrimSpace(emphasisStrip.Replace(text))
	if r := []rune(text); len(r) > cardSummaryMax {
		return strings.TrimSpace(string(r[:cardSummaryMax])) + "…"
	}
	return text
}

// extractImage returns the first image URL in the source: a raw <img src>
// first, else a Markdown ![alt](url), else "".
func extractImage(source string) string {
	if m := htmlImgRe.FindStringSubmatch(source); m != nil {
		return m[1]
	}
	if m := mdImgRe.FindStringSubmatch(source); m != nil {
		return m[1]
	}
	return ""
}

// buildCardHTML renders one linked card; the <img> is omitted when image == "".
func buildCardHTML(path, title, summary, image string) string {
	img := ""
	if image != "" {
		img = `<img class="card-img" src="` + html.EscapeString(image) + `" alt="">`
	}
	return `<a class="card" href="` + html.EscapeString(path) + `">` + img +
		`<div class="card-body">` +
		`<h3 class="card-title">` + html.EscapeString(title) + `</h3>` +
		`<p class="card-summary">` + html.EscapeString(summary) + `</p>` +
		`</div></a>`
}

// selectByTags is the authoritative List-membership rule: OR intersection with
// listTags, self excluded, incoming Order preserved.
func selectByTags(pages []models.PageModel, listTags []string, selfName string) []models.PageModel {
	if len(listTags) == 0 {
		return nil
	}
	want := make(map[string]bool, len(listTags))
	for _, t := range listTags {
		want[t] = true
	}
	var out []models.PageModel
	for _, p := range pages {
		if p.PageName == selfName {
			continue
		}
		for _, t := range p.Tags {
			if want[t] {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

// expandShortcodes replaces every <card path="…"> with resolve(path); an
// unresolved path is dropped so no raw shortcode leaks to the browser.
func expandShortcodes(htmlStr string, resolve func(path string) (string, bool)) string {
	return cardShortcodeRe.ReplaceAllStringFunc(htmlStr, func(match string) string {
		sub := cardShortcodeRe.FindStringSubmatch(match)
		if card, ok := resolve(sub[1]); ok {
			return card
		}
		return ""
	})
}

type CardService interface {
	ExpandShortcodes(htmlStr string) (string, error)
	ExpandCards(htmlStr string, page models.PageModel) (string, error)
}

type CardServiceImplementation struct {
	pages PageService
	menus MenuService
}

func NewCardService(pages PageService, menus MenuService) CardService {
	return &CardServiceImplementation{pages: pages, menus: menus}
}

// cardFor renders a fetched page as a card, applying overrides then auto-extract.
func (c *CardServiceImplementation) cardFor(page models.PageModel) string {
	caption := ""
	if m, err := c.menus.GetByPageName(page.PageName); err == nil {
		caption = m.Caption
	}
	source := FromStorage(page.Page)
	summary := page.Summary
	if summary == "" {
		summary = extractSummary(source)
	}
	image := page.Image
	if image == "" {
		image = extractImage(source)
	}
	return buildCardHTML(page.Path, cardTitle(caption, page.PageName), summary, image)
}

func (c *CardServiceImplementation) resolveCard(path string) (string, bool) {
	page, err := c.pages.GetRawByPath(path)
	if err != nil {
		return "", false
	}
	return c.cardFor(page), true
}

func (c *CardServiceImplementation) ExpandShortcodes(htmlStr string) (string, error) {
	return expandShortcodes(htmlStr, c.resolveCard), nil
}

func (c *CardServiceImplementation) ExpandCards(htmlStr string, page models.PageModel) (string, error) {
	out, err := c.ExpandShortcodes(htmlStr)
	if err != nil {
		return htmlStr, err
	}
	if page.ViewType != "List" {
		return out, nil
	}
	candidates, err := c.pages.FindByTags(page.ListTags)
	if err != nil {
		// Best-effort: keep the shortcode-expanded body; skip the grid on error.
		return out, nil
	}
	selected := selectByTags(candidates, page.ListTags, page.PageName)
	if len(selected) == 0 {
		return out, nil
	}
	var grid strings.Builder
	grid.WriteString(`<div class="card-grid">`)
	for _, p := range selected {
		grid.WriteString(c.cardFor(p))
	}
	grid.WriteString(`</div>`)
	return out + grid.String(), nil
}
