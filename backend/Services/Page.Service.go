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
	ConvertmdToHTML(md []byte) []byte
	Preview(sourceClean string) (string, error)
	List() ([]models.PageSummary, error)
	GetRaw(name string) (models.PageModel, error)
	Create(name, sourceClean, viewType string) error
	Update(name, sourceClean, viewType string) error
	Delete(name string) error
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
		{Key: "ViewType", Value: 1},
	})
	cursor, err := psi.collection.Find(psi.ctx, bson.D{{}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(psi.ctx)
	var summaries []models.PageSummary
	if err := cursor.All(psi.ctx, &summaries); err != nil {
		return nil, err
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

func (psi *PageServiceImplementation) Create(name, sourceClean, viewType string) error {
	count, err := psi.collection.CountDocuments(psi.ctx, bson.D{{Key: "PageName", Value: name}})
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPageExists
	}
	_, err = psi.collection.InsertOne(psi.ctx, bson.D{
		{Key: "PageName", Value: name},
		{Key: "Page", Value: ToStorage(sourceClean)},
		{Key: "Hash", Value: []byte{}},
		{Key: "Text", Value: ""},
		{Key: "ViewType", Value: viewType},
	})
	return err
}

func (psi *PageServiceImplementation) Update(name, sourceClean, viewType string) error {
	res, err := psi.collection.UpdateOne(psi.ctx,
		bson.D{{Key: "PageName", Value: name}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "Page", Value: ToStorage(sourceClean)},
			{Key: "ViewType", Value: viewType},
			{Key: "Hash", Value: []byte{}},
			{Key: "Text", Value: ""},
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

func (psi *PageServiceImplementation) GetPage(pn string) (models.PageModel, error) {
	var Page models.PageModel
	querry := bson.D{bson.E{Key: "PageName", Value: pn}}
	err := psi.collection.FindOne(psi.ctx, querry).Decode(&Page);
	if err != nil {
		return models.PageModel{}, err
	}
	hasher := sha1.New()
	_, err = io.WriteString(hasher, Page.Page)
	if err != nil {
		return models.PageModel{}, err
	}
	if Page.Text == "" || !testEq(Page.Hash, hasher.Sum(nil)) {
		Page.Text, err = psi.GetPageText(Page.Page)
		Page.Hash = hasher.Sum(nil)
		if err != nil {
			return models.PageModel{}, err
		}
		_, err = psi.collection.UpdateOne(psi.ctx, querry, bson.D{
			bson.E{Key: "$set", Value: bson.D{
				bson.E{Key: "Hash", Value: Page.Hash},
				bson.E{Key: "Text", Value: Page.Text},
			}},
		})
		if err != nil {
			return models.PageModel{}, err
		}
	}
	Page.Text = strings.ReplaceAll(Page.Text, "\n", "")
	return Page, nil
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
