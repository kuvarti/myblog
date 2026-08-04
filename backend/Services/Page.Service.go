package services

import (
	models "backend/Models"
	"context"
	"crypto/sha1"
	"errors"
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPageExists   = errors.New("page already exists")
	ErrPageNotFound = errors.New("page not found")
)

type PageService interface {
	GetPage(string) (models.PageModel, error)
	GetPageByPath(path string) (models.PageModel, error)
	ConvertmdToHTML(md []byte) []byte
	Preview(sourceClean string) (string, error)
	List() ([]models.PageSummary, error)
	GetRaw(name string) (models.PageModel, error)
	Create(w models.PageWrite) error
	Update(name string, w models.PageWrite) error
	GetRawByPath(path string) (models.PageModel, error)
	FindByTags(tags []string) ([]models.PageModel, error)
	Delete(name string) error
	SetOrder(names []string) error
	SetVisibility(name string, visible bool) error
	PathTaken(path, excludePageName string) (bool, error)
}

// ToStorage encodes clean newlines into the bespoke "/n" line delimiter used in
// the stored Page source. FromStorage reverses it. The render pipeline is left
// untouched — these only translate the delimiter at the API boundary.
func ToStorage(clean string) string {
	return strings.ReplaceAll(clean, "\n", "/n")
}

func FromStorage(stored string) string {
	return strings.ReplaceAll(stored, "/n", "\n")
}

type PageServiceImplementation struct {
	collection *mongo.Collection
	ctx context.Context
}

func NewPageService(ctx context.Context, col *mongo.Collection) PageService {
	return &PageServiceImplementation{
		collection: col,
		ctx: ctx,
	}
}

func (psi *PageServiceImplementation) ConvertmdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

// Preview renders raw clean-newline source to HTML without persisting, mirroring
// GetPage's render path exactly (storage-encode, GetPageText, strip newlines).
func (psi *PageServiceImplementation) Preview(sourceClean string) (string, error) {
	text, err := psi.GetPageText(ToStorage(sourceClean))
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(text, "\n", ""), nil
}

func (psi *PageServiceImplementation) List() ([]models.PageSummary, error) {
	opts := options.Find().SetProjection(bson.D{
		{Key: "PageName", Value: 1},
		{Key: "Path", Value: 1},
		{Key: "ViewType", Value: 1},
		{Key: "Order", Value: 1},
		{Key: "Hidden", Value: 1},
	}).SetSort(bson.D{{Key: "Order", Value: 1}})
	cursor, err := psi.collection.Find(psi.ctx, bson.D{{}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(psi.ctx)
	var summaries []models.PageSummary
	if err := cursor.All(psi.ctx, &summaries); err != nil {
		return nil, err
	}
	for i := range summaries {
		summaries[i].Visible = !summaries[i].Hidden
	}
	return summaries, nil
}

func (psi *PageServiceImplementation) GetRaw(name string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "PageName", Value: name}}).Decode(&page)
	if err != nil {
		return models.PageModel{}, err
	}
	return page, nil
}

func (psi *PageServiceImplementation) Create(w models.PageWrite) error {
	count, err := psi.collection.CountDocuments(psi.ctx, bson.D{{Key: "PageName", Value: w.PageName}})
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPageExists
	}
	total, err := psi.collection.CountDocuments(psi.ctx, bson.D{{}})
	if err != nil {
		return err
	}
	_, err = psi.collection.InsertOne(psi.ctx, bson.D{
		{Key: "PageName", Value: w.PageName},
		{Key: "Path", Value: w.Path},
		{Key: "Page", Value: ToStorage(w.Source)},
		{Key: "Hash", Value: []byte{}},
		{Key: "Text", Value: ""},
		{Key: "ViewType", Value: w.ViewType},
		{Key: "Order", Value: total},
		{Key: "Hidden", Value: false},
		{Key: "Tags", Value: w.Tags},
		{Key: "Summary", Value: w.Summary},
		{Key: "Image", Value: w.Image},
		{Key: "ListTags", Value: w.ListTags},
	})
	return err
}

// GetRawByPath fetches a page by Path without rendering/caching — used to read
// card metadata off a referenced page.
func (psi *PageServiceImplementation) GetRawByPath(path string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "Path", Value: path}}).Decode(&page)
	return page, err
}

// FindByTags returns candidate pages whose Tags intersect tags (Mongo $in),
// Order-sorted, raw (source included). Authoritative membership is selectByTags.
func (psi *PageServiceImplementation) FindByTags(tags []string) ([]models.PageModel, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	opts := options.Find().SetSort(bson.D{{Key: "Order", Value: 1}})
	cursor, err := psi.collection.Find(psi.ctx,
		bson.D{{Key: "Tags", Value: bson.D{{Key: "$in", Value: tags}}}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(psi.ctx)
	var pages []models.PageModel
	if err := cursor.All(psi.ctx, &pages); err != nil {
		return nil, err
	}
	return pages, nil
}

// PathTaken reports whether some other page already owns this path. The current
// page is excluded by PageName so re-saving an unchanged path is allowed.
func (psi *PageServiceImplementation) PathTaken(path, excludePageName string) (bool, error) {
	count, err := psi.collection.CountDocuments(psi.ctx, bson.D{
		{Key: "Path", Value: path},
		{Key: "PageName", Value: bson.D{{Key: "$ne", Value: excludePageName}}},
	})
	return count > 0, err
}

// SetOrder rewrites each named page's Order to its position in the slice, so the
// list (and the Order-sorted public menu) follow the given sequence. Names not
// found are skipped.
func (psi *PageServiceImplementation) SetOrder(names []string) error {
	for i, name := range names {
		_, err := psi.collection.UpdateOne(psi.ctx,
			bson.D{{Key: "PageName", Value: name}},
			bson.D{{Key: "$set", Value: bson.D{{Key: "Order", Value: i}}}},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetVisibility flips a page's Hidden flag. Storage is inverted (Hidden), so a
// visible page is Hidden:false; a missing field also reads as visible.
func (psi *PageServiceImplementation) SetVisibility(name string, visible bool) error {
	res, err := psi.collection.UpdateOne(psi.ctx,
		bson.D{{Key: "PageName", Value: name}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "Hidden", Value: !visible}}}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrPageNotFound
	}
	return nil
}

func (psi *PageServiceImplementation) Update(name string, w models.PageWrite) error {
	res, err := psi.collection.UpdateOne(psi.ctx,
		bson.D{{Key: "PageName", Value: name}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "Path", Value: w.Path},
			{Key: "Page", Value: ToStorage(w.Source)},
			{Key: "ViewType", Value: w.ViewType},
			{Key: "Hash", Value: []byte{}},
			{Key: "Text", Value: ""},
			{Key: "Tags", Value: w.Tags},
			{Key: "Summary", Value: w.Summary},
			{Key: "Image", Value: w.Image},
			{Key: "ListTags", Value: w.ListTags},
		}}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrPageNotFound
	}
	return nil
}

func (psi *PageServiceImplementation) Delete(name string) error {
	res, err := psi.collection.DeleteOne(psi.ctx, bson.D{{Key: "PageName", Value: name}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrPageNotFound
	}
	return nil
}

func (psi *PageServiceImplementation) GetPageText(pn string) (string, error) {
	var html []byte
	lines := strings.Split(strings.ReplaceAll(pn, "/n", "\n"), "\n");
	for i := 0; i < len(lines); i++{
		if strings.Contains(lines[i], "<") {
			html = append(html, []byte(lines[i])...);
		} else if lines[i] == "" {
			continue
		} else {
			var newConvert []byte
			// Accumulate the run of consecutive non-HTML lines into one Markdown
			// block. The bound (i < len(lines)) is required: without it the loop
			// reads past the slice and panics when the source ends on a Markdown
			// line (no trailing "<" line or blank separator).
			for i < len(lines) && !strings.Contains(lines[i], "<") {
				newConvert = append(newConvert, []byte(lines[i])...)
				newConvert = append(newConvert, []byte("\n")...)
				i++
			}
			i--
			html = append(html, psi.ConvertmdToHTML(newConvert)...);
		}
	}
	return string(html), nil
}

// renderAndCache ensures an already-fetched page has an up-to-date rendered
// Text cache, persisting it when stale, and returns the page with newlines
// stripped for transport. The cache write is keyed on PageName so it works
// whether the page was found by name or by path.
func (psi *PageServiceImplementation) renderAndCache(page models.PageModel) (models.PageModel, error) {
	hasher := sha1.New()
	if _, err := io.WriteString(hasher, page.Page); err != nil {
		return models.PageModel{}, err
	}
	if page.Text == "" || !testEq(page.Hash, hasher.Sum(nil)) {
		text, err := psi.GetPageText(page.Page)
		if err != nil {
			return models.PageModel{}, err
		}
		page.Text = text
		page.Hash = hasher.Sum(nil)
		_, err = psi.collection.UpdateOne(psi.ctx,
			bson.D{{Key: "PageName", Value: page.PageName}},
			bson.D{{Key: "$set", Value: bson.D{
				{Key: "Hash", Value: page.Hash},
				{Key: "Text", Value: page.Text},
			}}},
		)
		if err != nil {
			return models.PageModel{}, err
		}
	}
	page.Text = strings.ReplaceAll(page.Text, "\n", "")
	return page, nil
}

func (psi *PageServiceImplementation) GetPage(pn string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "PageName", Value: pn}}).Decode(&page)
	if err != nil {
		return models.PageModel{}, err
	}
	return psi.renderAndCache(page)
}

func (psi *PageServiceImplementation) GetPageByPath(path string) (models.PageModel, error) {
	var page models.PageModel
	err := psi.collection.FindOne(psi.ctx, bson.D{{Key: "Path", Value: path}}).Decode(&page)
	if err != nil {
		return models.PageModel{}, err
	}
	return psi.renderAndCache(page)
}

func testEq(a, b []byte) bool {
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
