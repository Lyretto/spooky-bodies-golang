package controller

import (
	"net/http"

	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type levelAddParams struct {
	Content string `json:"content"`
}

func LevelsGetAll(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		levels := []model.Level{}

		db.Preload(clause.Associations).Limit(25).Find(&levels)

		context.JSON(http.StatusOK, gin.H{
			"results": levels,
		})
	}
}

func LevelsAdd(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var levelAddParams levelAddParams

		if err := context.BindJSON(&levelAddParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user model.User

		tx := db.Limit(1).Where("id = ?", "e97b3095-f92c-4d3e-a88b-25f2a4761c4a").Find(&user)

		if tx.Error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		level := model.Level{
			User:    &user,
			Content: levelAddParams.Content,
		}

		tx = db.Create(&level)

		if tx.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error})
			return
		}

		context.JSON(http.StatusOK, gin.H{"id": level.ID})
	}
}
