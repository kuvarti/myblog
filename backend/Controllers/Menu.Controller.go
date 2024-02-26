package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MenuController struct {
	MenuService services.MenuService
}

func InitMenuController(MenuService services.MenuService, server *gin.Engine) MenuController {
	group := server.Group("/MenuList")
	mc := MenuController{
		MenuService: MenuService,
	}
	group.GET("/Menu", mc.GetMenu)
	return mc
}

func (mc *MenuController) GetMenu(ctx *gin.Context) {
	var returnMenu []*models.MenuModel
	returnMenu, err := mc.MenuService.GetMenu()
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, returnMenu)
}
