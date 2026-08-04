package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PageController struct {
	PageService services.PageService
	CardService services.CardService
}

func InitPageController(PageService services.PageService, CardService services.CardService, server *gin.RouterGroup) PageController {
	pc := PageController{PageService: PageService, CardService: CardService}
	server.GET("/Page", pc.GetPage)
	return pc
}

// GetPage resolves a page by its Path (the per-page route key, preferred) or by
// PageName (legacy). A blank/missing query is a 400 rather than a panic.
func (pc *PageController) GetPage(ctx *gin.Context) {
	q := ctx.Request.URL.Query()
	var (
		respons models.PageModel
		err     error
	)
	if path := q.Get("Path"); path != "" {
		respons, err = pc.PageService.GetPageByPath(path)
	} else if name := q.Get("PageName"); name != "" {
		respons, err = pc.PageService.GetPage(name)
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Path or PageName query is required"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	final := respons.Text
	if expanded, exErr := pc.CardService.ExpandCards(respons.Text, respons); exErr == nil {
		final = expanded
	}
	ctx.JSON(http.StatusOK, gin.H{
		"ViewType": respons.ViewType,
		"Page":     final,
	})
}
