package controller

import (
	"net/http"
	"time"

	"github.com/Lyretto/spooky-bodies-golang/internal/auth"
	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type levelAddParams struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type levelUpdateParams struct {
	LevelID uuid.UUID `uri:"levelId" binding:"required,uuid"`
	Name    string    `json:"name"`
	Content string    `json:"content"`
}

type levelGetParams struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type levelDeleteParams struct {
	LevelID uuid.UUID `uri:"levelId" binding:"required,uuid"`
}

type levelValidateParams struct {
	LevelID          uuid.UUID `uri:"levelId" binding:"required,uuid"`
	Levelversion     int       `json:"version"`
	ValidationResult string    `json:"result"`
	AuthorScore      int       `json:"authorScore"`
	Thumbnail        []uint8   `json:"thumbnail"`
}

type levelReportParams struct {
	LevelID uuid.UUID `uri:"levelId" binding:"required,uuid"`
}

type levelVoteParams struct {
	LevelID  uuid.UUID      `uri:"levelId" binding:"required,uuid"`
	VoteType model.VoteType `json:"voteType"`
}

func levelsGetAll(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var getParams levelGetParams

		if err := context.BindJSON(&getParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		levels := []model.Level{}

		if err := db.Where("validationId != nil").Joins("JOIN validations ON validation.levelId = level.id AND validation.result = ?", model.ResultOk).Preload(clause.Associations).Offset(getParams.Offset).Limit(getParams.Limit).Find(&levels); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"results": levels,
		})
	}
}

func levelsLockValidation(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {

		//TODO: Lock level for validation (with timestamp 5 min), so multiple agents don't validate one level at the same time

		context.Status(http.StatusOK)
	}
}

func levelsGetAllSus(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		if !user.IsMod {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "no moderation authorization"})
			return
		}

		var getParams levelGetParams

		if err := context.BindJSON(&getParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		levels := []model.Level{}

		if err := db.Where("validationId = nil").Preload(clause.Associations).Offset(getParams.Offset).Limit(getParams.Limit).Find(&levels); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"results": levels,
		})
	}
}

func levelValidate(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		if !user.IsMod {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "no moderation authorization"})
			return
		}

		var validateParams levelValidateParams

		if err := context.BindUri(&validateParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := context.BindJSON(&validateParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var level model.Level
		tx := db.Where("id = ?", user.ID, validateParams.LevelID).First(&level)

		if tx.Error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": tx.Error})
			return
		}

		validation := model.Validation{
			LevelID:      validateParams.LevelID,
			LevelVersion: level.Version,
			Result:       validateParams.ValidationResult,
			ValidatorID:  user.ID,
		}

		tx = db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "version"}, {Name: "levelId"}},
			UpdateAll: true,
		}).Create(&validation)

		if tx.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error})
			return
		}

		level.ValidationId = validation.ID
		level.AuthorScore = validateParams.AuthorScore
		level.Thumbnail = validateParams.Thumbnail
		level.Published = time.Now()

		tx = db.Save(&level)

		if tx.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"validationId": validation.ID,
		})
	}
}

func levelsGetOwn(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var getParams levelGetParams

		if err := context.BindJSON(&getParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx := db.Where("userId = ?", user.ID)

		levels := []model.Level{}
		if err := tx.Preload(clause.Associations).Offset(getParams.Offset).Limit(getParams.Limit).Find(&levels); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"results": levels,
		})
	}
}

func levelsAdd(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var levelAddParams levelAddParams

		if err := context.BindJSON(&levelAddParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		level := model.Level{
			User:    user,
			Name:    levelAddParams.Name,
			Content: levelAddParams.Content,
		}

		tx := db.Create(&level)

		if tx.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error})
			return
		}

		context.JSON(http.StatusOK, gin.H{"id": level.ID})
	}
}

func levelsDelete(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var deleteParams levelDeleteParams

		if err := context.BindUri(&deleteParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		level := model.Level{
			ID: deleteParams.LevelID,
		}

		tx := db

		if !user.IsMod {
			tx = tx.Where("userId = ?", user.ID)
		}

		tx = tx.First(&level)

		if level.ID != uuid.Nil {
			db.Delete(&level)
		} else {
			context.JSON(http.StatusNonAuthoritativeInfo, gin.H{"error": "not authorized to delete this level"})
		}

		context.Status(http.StatusOK)
	}
}

func levelsUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var updateParams levelUpdateParams

		if err := context.BindUri(&updateParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := context.BindJSON(&updateParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		level := model.Level{
			ID:      updateParams.LevelID,
			UserID:  user.ID,
			Name:    updateParams.Name,
			Content: updateParams.Content,
		}

		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "level_id"}},
			UpdateAll: true,
		}).Create(&level).Error; err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		context.Status(http.StatusOK)
	}
}

func levelVote(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var voteParams levelVoteParams

		if err := context.BindJSON(&voteParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if voteParams.VoteType == "" {
			context.JSON(http.StatusBadRequest, gin.H{"error": "missing vote type"})
			return
		}

		var level model.Level
		tx := db.Limit(1).Where("userid!=? AND id = ?", user.ID, voteParams.LevelID).Find(&level)

		if tx.Error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "level not found"})
			return
		}

		vote := model.Vote{
			User:  user,
			Level: &level,
			Type:  voteParams.VoteType,
		}

		tx = db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "userId"}, {Name: "levelId"}},
			UpdateAll: true,
		}).Create(&vote)

		if tx.Error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "level not found"})
			return
		}

		context.JSON(http.StatusOK, gin.H{"voteID": vote.ID})
	}
}

func levelReport(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var reportParams levelReportParams

		if err := context.BindJSON(&reportParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var level model.Level
		tx := db.Where("userid != ? AND id = ?", user.ID, reportParams.LevelID).First(&level)

		if tx.Error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "level not found"})
			return
		}

		report := model.Report{
			User:  user,
			Level: &level,
		}

		tx = db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "userId"}, {Name: "levelId"}},
			UpdateAll: true,
		}).Create(&report)

		if tx.Error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "level not found"})
			return
		}

		// TODO: Update Validation of level when reports above threshold

		context.JSON(http.StatusOK, gin.H{"reportID": report.ID})
	}
}

func UseLevel(router gin.IRouter, db *gorm.DB) {
	levelRouter := router.Group("/media")

	levelRouter.GET("/levels", levelsGetAll(db))
	levelRouter.GET("/levels/sus", levelsGetAllSus(db))
	levelRouter.GET("/levels/own", levelsGetOwn(db))

	levelRouter.DELETE("/levels/{levelId}", levelsDelete(db))

	levelRouter.POST("/levels", levelsAdd(db))

	levelRouter.PUT("/levels/{levelId}", levelsUpdate(db))
	levelRouter.PUT("/levels/{levelId}/reports", levelReport(db))
	levelRouter.PUT("/levels/{levelId}/vote", levelVote(db))
	levelRouter.PUT("/levels/{levelId}/validate", levelValidate(db))
}
