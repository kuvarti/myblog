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
	// The page is the source of truth for nav membership/order/visibility; the
	// menu docs only supply display captions/paths. If the page list is
	// unavailable, fall back to returning the raw menus.
	pages, err := mc.PageService.List()
	if err != nil {
		ctx.JSON(http.StatusOK, menus)
		return
	}
	ctx.JSON(http.StatusOK, buildNav(pages, menus))
}

// buildNav produces the public navigation from the pages (the source of truth
// for membership, order, and visibility) joined to their menu documents for
// display captions/paths. Pages arrive already Order-sorted from
// PageService.List(). A visible page with no menu document falls back to its
// PageName as the caption; menu documents with no matching page (hand-seeded
// stubs) are dropped.
func buildNav(pages []models.PageSummary, menus []*models.MenuModel) []*models.MenuModel {
	byPage := make(map[string]*models.MenuModel, len(menus))
	for _, m := range menus {
		if m.PageName != "" {
			byPage[m.PageName] = m
		}
	}
	nav := make([]*models.MenuModel, 0, len(pages))
	for _, p := range pages {
		if !p.Visible {
			continue
		}
		if m, ok := byPage[p.PageName]; ok {
			nav = append(nav, m)
		} else {
			nav = append(nav, &models.MenuModel{
				Name:     p.PageName,
				Caption:  p.PageName,
				Path:     "",
				PageName: p.PageName,
			})
		}
	}
	return nav
}
