package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type MenuController struct {
	MenuService services.MenuService
	PageService services.PageService
}

func InitMenuController(MenuService services.MenuService, PageService services.PageService, server *gin.RouterGroup) MenuController {
	group := server.Group("/MenuList")
	mc := MenuController{
		MenuService: MenuService,
		PageService: PageService,
	}
	group.GET("/Menu", mc.GetMenu)
	return mc
}

func (mc *MenuController) GetMenu(ctx *gin.Context) {
	menus, err := mc.MenuService.GetMenu()
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}
	// Order lives on the page (single source of truth); sort the nav by it.
	orderMap := map[string]int{}
	if summaries, err := mc.PageService.List(); err == nil {
		for _, s := range summaries {
			orderMap[s.PageName] = s.Order
		}
	}
	ctx.JSON(http.StatusOK, sortMenusByOrder(menus, orderMap))
}

// sortMenusByOrder stably sorts menu entries by their page's Order. Entries whose
// page is absent from the map sort last, keeping their relative order.
func sortMenusByOrder(menus []*models.MenuModel, orderMap map[string]int) []*models.MenuModel {
	const last = int(^uint(0) >> 1) // max int
	order := func(m *models.MenuModel) int {
		if o, ok := orderMap[m.PageName]; ok {
			return o
		}
		return last
	}
	sort.SliceStable(menus, func(i, j int) bool {
		return order(menus[i]) < order(menus[j])
	})
	return menus
}
