package services

import (
	models "backend/Models"
	"context"
	"crypto/sha1"
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PageService interface {
	GetPage(string) (string, error)
	ConvertmdToHTML(md []byte) []byte
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

func (psi *PageServiceImplementation) GetPageText(pn string) (string, error) {
	var html []byte
	deneme := strings.Split(pn, "\n");
	for i := 0; i < len(deneme); i++{
		if strings.Contains(deneme[i], "<") {
			html = append(html, []byte(deneme[i])...);
		} else if deneme[i] == "" {
			continue
		} else {
			var newConvert []byte
			for {
				if !strings.Contains(deneme[i], "<") {
					newConvert = append(newConvert, []byte(deneme[i])...)
					newConvert = append(newConvert, []byte("\n")...)
					i++
				} else {
					break
				}
				if (i+1) == len(deneme) {
					break
				}
			}
			i--
			html = append(html, psi.ConvertmdToHTML(newConvert)...);
		}
	}
	return string(html), nil
}

func (psi *PageServiceImplementation) GetPage(pn string) (string, error) {
	var Page models.PageModel
	querry := bson.D{bson.E{Key: "PageName", Value: pn}}
	err := psi.collection.FindOne(psi.ctx, querry).Decode(&Page);
	if err != nil {
		return "", err
	}
	hasher := sha1.New()
	_, err = io.WriteString(hasher, Page.Page)
	if err != nil {
		return "", err
	}
	if Page.Text == "" || !testEq(Page.Hash, hasher.Sum(nil)) {
		Page.Text, err = psi.GetPageText(Page.Page)
		Page.Hash = hasher.Sum(nil)
		if err != nil {
			return "", err
		}
		_, err = psi.collection.UpdateOne(psi.ctx, querry, bson.D{
			bson.E{Key: "$set", Value: bson.D{
				bson.E{Key: "Page", Value: Page.Text},
				bson.E{Key: "Hash", Value: Page.Hash},
				bson.E{Key: "Text", Value: Page.Text},
			}},
		})
		if err != nil {
			return "", err
		}
	}
	response := strings.ReplaceAll(Page.Text, "\n", "")
	return response, nil
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
