package controller

import (
	"net/http"

	"github.com/Lyretto/spooky-bodies-golang/internal/auth"
	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type signUpParams struct {
	PlatformType   model.PlatformType `gorm:"type:string" json:"platformType"`
	PlatformUserID string             `json:"platformUserId"`
}

func authSignUp(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var signUpParams signUpParams

		if err := context.BindJSON(&signUpParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//TODO platform type none only in develop environment
		if signUpParams.PlatformType == model.PlatformNone {
			signUpParams.PlatformUserID = uuid.NewString()
		}

		user := model.User{
			PlatformType:   signUpParams.PlatformType,
			PlatformUserID: signUpParams.PlatformUserID,
		}

		tx := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "platform_type"}, {Name: "platform_user_id"}},
			DoNothing: true,
		}).Create(&user)

		if tx == nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": tx.Error})
			return
		}

		context.JSON(http.StatusOK, gin.H{"platformUserId": user.PlatformUserID})
	}
}

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

func UseAuth(router gin.IRouter, db *gorm.DB) error {
	mw, err := auth.GetJWTMiddleware(db)

	if err != nil {
		return err
	}
	router.POST("/auth/signup", authSignUp(db))
	router.POST("/auth/login", mw.LoginHandler)

	router.Use(mw.MiddlewareFunc())
	router.Use(authCheckTokenActivityMiddlewareFunc(db))

	authRouter := router.Group("/auth")

	authRouter.POST("refresh_token", mw.RefreshHandler)
	authRouter.POST("logout", authLogout(db))

	return nil
}
