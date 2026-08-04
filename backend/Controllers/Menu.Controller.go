package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"net/http"

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
	// The page is the source of truth for nav membership/order/visibility and now
	// its own Path; the menu docs only supply display captions. If the page list
	// is unavailable, fall back to returning the raw menus.
	pages, err := mc.PageService.List()
	if err != nil {
		ctx.JSON(http.StatusOK, menus)
		return
	}
	ctx.JSON(http.StatusOK, buildNav(pages, menus))
}

// buildNav produces the public navigation from the pages (the source of truth
// for membership, order, and visibility) joined to their menu documents for the
// display caption only. Each nav item's Path comes from the page itself. Pages
// arrive already Order-sorted from PageService.List(). A visible page with no
// menu caption falls back to its PageName; menu documents with no matching page
// (hand-seeded stubs) are dropped.
func buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel {
	capByPage := make(map[string]string, len(menus))
	for _, m := range menus {
		if m.PageName != "" {
			capByPage[m.PageName] = m.Caption
		}
	}
	nav := make([]*models.MenuModel, 0, len(pages))
	for _, p := range pages {
		if !p.Visible {
			continue
		}
		caption := p.PageName
		if c, ok := capByPage[p.PageName]; ok && c != "" {
			caption = c
		}
		nav = append(nav, &models.MenuModel{
			Name:     p.PageName,
			Caption:  caption,
			Path:     p.Path,
			PageName: p.PageName,
		})
	}
	return nav
}
