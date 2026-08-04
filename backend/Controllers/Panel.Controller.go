package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type PanelController struct {
	PageService services.PageService
	MenuService services.MenuService
	CardService services.CardService
}

func InitPanelController(pageService services.PageService, menuService services.MenuService, cardService services.CardService, tokenService services.TokenService, apiGroup *gin.RouterGroup) PanelController {
	pc := PanelController{PageService: pageService, MenuService: menuService, CardService: cardService}
	cp := apiGroup.Group("/auth/ControlPanel")
	cp.Use(tokenService.AuthenticateJWT())
	{
		cp.GET("/Pages", pc.ListPages)
		cp.GET("/Pages/:name", pc.GetPage)
		cp.POST("/Pages", pc.CreatePage)
		cp.PUT("/Pages/:name", pc.UpdatePage)
		cp.DELETE("/Pages/:name", pc.DeletePage)
		cp.PUT("/PageOrder", pc.ReorderPages)
		cp.PUT("/PageVisibility", pc.SetPageVisibility)
		cp.POST("/Preview", pc.Preview)
	}
	return pc
}

// pathValidity classifies a page path so CreatePage/UpdatePage can map each
// failure to the right HTTP status. Uniqueness needs the DB and is checked
// separately; this covers only format and reserved-route rules.
type pathValidity int

const (
	pathOK        pathValidity = iota
	pathBadFormat              // empty or missing leading slash
	pathReserved               // collides with a client-only route
)

func validatePagePathFormat(path string) pathValidity {
	if path == "" || !strings.HasPrefix(path, "/") {
		return pathBadFormat
	}
	if path == "/panel" || path == "/lists" {
		return pathReserved
	}
	return pathOK
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
		Path:     page.Path,
		Source:   services.FromStorage(page.Page),
		ViewType: page.ViewType,
		Tags:     page.Tags,
		Summary:  page.Summary,
		Image:    page.Image,
		ListTags: page.ListTags,
	}
	if menu, err := pc.MenuService.GetByPageName(name); err == nil {
		detail.Menu = &models.MenuBinding{Name: menu.Name, Caption: menu.Caption}
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
	switch validatePagePathFormat(req.Path) {
	case pathBadFormat:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path must start with /"})
		return
	case pathReserved:
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path is reserved"})
		return
	}
	taken, err := pc.PageService.PathTaken(req.Path, req.PageName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if taken {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path already used by another page"})
		return
	}
	if err := pc.PageService.Create(models.PageWrite{
		PageName: req.PageName, Path: req.Path, Source: req.Source, ViewType: req.ViewType,
		Tags: req.Tags, Summary: req.Summary, Image: req.Image, ListTags: req.ListTags,
	}); err != nil {
		if errors.Is(err, services.ErrPageExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "a page with that name already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, PageName: req.PageName}
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
	switch validatePagePathFormat(req.Path) {
	case pathBadFormat:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path must start with /"})
		return
	case pathReserved:
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path is reserved"})
		return
	}
	taken, err := pc.PageService.PathTaken(req.Path, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if taken {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "path already used by another page"})
		return
	}
	if err := pc.PageService.Update(name, models.PageWrite{
		Path: req.Path, Source: req.Source, ViewType: req.ViewType,
		Tags: req.Tags, Summary: req.Summary, Image: req.Image, ListTags: req.ListTags,
	}); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.Menu != nil {
		menu := models.MenuModel{Name: req.Menu.Name, Caption: req.Menu.Caption, PageName: name}
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

func (pc *PanelController) SetPageVisibility(ctx *gin.Context) {
	var req models.VisibilityRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := pc.PageService.SetVisibility(req.PageName, req.Visible); err != nil {
		if errors.Is(err, services.ErrPageNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
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
	if expanded, exErr := pc.CardService.ExpandShortcodes(html); exErr == nil {
		html = expanded
	}
	ctx.JSON(http.StatusOK, gin.H{"Html": html})
}
