package controllers

import (
	models "backend/Models"
	services "backend/Services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	TokenService services.TokenService
	UserService services.UserService
}

func InitUserController(UserService services.UserService, TokenService services.TokenService, server *gin.RouterGroup) UserController {
	uc := UserController {
		UserService: UserService,
		TokenService: TokenService,
	}
	authGroup := server.Group("/auth")
	controlpanel := authGroup.Group("ControlPanel");
	authGroup.POST("/login", uc.Login)
	authGroup.Use(uc.TokenService.AuthenticateJWT())
	{
		authGroup.GET("/AmIAuth", uc.Authc)
	}
	controlpanel.Use(uc.TokenService.AuthenticateJWT())
	{
	}
	return uc
}

func (uc UserController) Login(ctx *gin.Context) {
	var loginData models.LoginModel
	if err := ctx.BindJSON(&loginData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := uc.UserService.Login(loginData.Username, loginData.Password); err != nil {
		ctx.JSON(http.StatusOK, err.Error())
		return
	}
	if token, err := uc.TokenService.GenerateToken(loginData.Username); err != nil {
		ctx.JSON(http.StatusContinue, err.Error())
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"Message": "Login Complated!",
			"token": token,
		})
	}
}

func (uc UserController) Authc(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{ "Yes": "YES"})
}
