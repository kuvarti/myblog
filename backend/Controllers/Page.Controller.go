package controllers

import (
	services "backend/Services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PageController struct {
	PageService services.PageService
}

func InitPageController(PageService services.PageService, server *gin.Engine) PageController {
	pc := PageController{
		PageService: PageService,
	}
	server.GET("/Page", pc.GetPage)
	return pc
}

func (pc *PageController) GetPage(ctx *gin.Context) {
	name := ctx.Request.URL.Query()
	respons, err := pc.PageService.GetPage(name["PageName"][0])
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"ViewType" : respons.ViewType,
		"Page" : respons.Text,
	})
}
