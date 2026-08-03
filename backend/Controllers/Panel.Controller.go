package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PanelController struct {
	PageService services.PageService
	MenuService services.MenuService
}

func InitPanelController(pageService services.PageService, menuService services.MenuService, tokenService services.TokenService, apiGroup *gin.RouterGroup) PanelController {
	pc := PanelController{PageService: pageService, MenuService: menuService}
	cp := apiGroup.Group("/auth/ControlPanel")
	cp.Use(tokenService.AuthenticateJWT())
	{
		cp.GET("/Pages", pc.ListPages)
		cp.GET("/Pages/:name", pc.GetPage)
		cp.POST("/Pages", pc.CreatePage)
		cp.PUT("/Pages/:name", pc.UpdatePage)
		cp.DELETE("/Pages/:name", pc.DeletePage)
		cp.PUT("/PageOrder", pc.ReorderPages)
		cp.POST("/Preview", pc.Preview)
	}
	return pc
}

func (pc *PanelController) ListPages(ctx *gin.Context) {
	pages, err := pc.PageService.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, pages)
}

func (pc *PanelController) GetPage(ctx *gin.Context) {
	name := ctx.Param("name")
	page, err := pc.PageService.GetRaw(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	detail := models.PageDetail{
		PageName: page.PageName,
		Source:   services.FromStorage(page.Page),
		ViewType: page.ViewType,
	}
	if menu, err := pc.MenuService.GetByPageName(name); err == nil {
		detail.Menu = &models.MenuBinding{Name: menu.Name, Caption: menu.Caption, Path: menu.Path}
	}
	ctx.JSON(http.StatusOK, detail)
}

func (pc *PanelController) CreatePage(ctx *gin.Context) {
	var req models.CreatePageRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.PageName == "" || req.Source == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "PageName and Source are required"})
		return
	}
	if err := pc.PageService.Create(req.PageName, req.Source, req.ViewType); err != nil {
		if errors.Is(err, services.ErrPageExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "a page with that name already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, Path: req.Menu.Path, PageName: req.PageName}
		if err := pc.MenuService.Upsert(menu); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func (pc *PanelController) UpdatePage(ctx *gin.Context) {
	name := ctx.Param("name")
	var req models.UpdatePageRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Source == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Source is required"})
		return
	}
	if err := pc.PageService.Update(name, req.Source, req.ViewType); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, Path: req.Menu.Path, PageName: name}
		if err := pc.MenuService.Upsert(menu); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (pc *PanelController) DeletePage(ctx *gin.Context) {
	name := ctx.Param("name")
	if err := pc.PageService.Delete(name); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := pc.MenuService.DeleteByPageName(name); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (pc *PanelController) ReorderPages(ctx *gin.Context) {
	var req models.ReorderRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := pc.PageService.SetOrder(req.PageNames); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

func (pc *PanelController) Preview(ctx *gin.Context) {
	var req models.PreviewRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	html, err := pc.PageService.Preview(req.Source)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"Html": html})
}
