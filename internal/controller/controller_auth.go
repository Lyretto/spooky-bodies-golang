package controller

import (
	"net/http"

	"github.com/Lyretto/spooky-bodies-golang/internal/auth"
	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func authLogout(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		token := jwt.GetToken(context)

		if token == "" {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		user := auth.GetJWTUser(context)

		if user == nil {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		db.Where(&model.UserToken{
			User:  user,
			Token: token,
		}).Delete(&model.UserToken{})

		context.Status(http.StatusNoContent)
	}
}

func authCheckTokenActivityMiddlewareFunc(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		_, _, _ = auth.CheckTokenActivity(context, db)
	}
}

func useAuth(router gin.IRouter, db *gorm.DB) error {
	mw, err := auth.GetJWTMiddleware(db)

	if err != nil {
		return err
	}

	// login should bypass jwt middleware
	router.POST("/auth/login", mw.LoginHandler)

	router.Use(mw.MiddlewareFunc())
	router.Use(authCheckTokenActivityMiddlewareFunc(db))

	authRouter := router.Group("/auth")

	authRouter.POST("refresh_token", mw.RefreshHandler)
	authRouter.POST("logout", authLogout(db))

	return nil
}
