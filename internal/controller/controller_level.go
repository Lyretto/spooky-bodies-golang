package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/Lyretto/spooky-bodies-golang/internal/auth"
	"github.com/Lyretto/spooky-bodies-golang/internal/config"
	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type levelParams struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type levelGetParams struct {
	Offset  int `form:"offset"`
	Limit   int `form:"limit"`
	OnlySus int `form:"only_sus"`
}

type levelValidateParams struct {
	Levelversion     int     `json:"version"`
	Content          string  `json:"content"`
	ValidationResult string  `json:"result"`
	AuthorScore      int     `json:"authorScore"`
	Thumbnail        []uint8 `json:"thumbnail"`
}

type levelVoteParams struct {
	VoteType model.VoteType `json:"voteType"`
}

func levelsGetAll(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var getParams levelGetParams

		if err := context.BindQuery(&getParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		levels := []model.Level{}

		var levelCount int64
		var err error

		user := auth.GetJWTUser(context)

		if config.C.Environment != config.EnvironmentProduction || user.Role == model.UserRoleMod || user.Role == model.UserRoleAgent {
			tx := db.Model(&model.Level{}).Preload(clause.Associations)

			if getParams.OnlySus == 1 {
				tx = tx.Where("validation_id is null")
			}

			if user.Role == model.UserRoleAgent {
				tx = tx.Where("validation_lock is null OR validation_lock < ?", time.Now().Add(time.Minute*time.Duration(config.C.TokenLifeSpan)))
			}

			tx.Count(&levelCount)

			retrieveTx := tx.Offset(getParams.Offset).
				Limit(getParams.Limit).
				Find(&levels)

			err = retrieveTx.Error
		} else {
			tx := db.
				Model(&model.Level{}).
				Preload(clause.Associations).
				Where("validation_id is not null").
				Joins("JOIN "+((&model.Validation{}).TableName())+" v ON v.level_id = "+((&model.Level{}).TableName())+".id AND v.result = ?", model.ResultOk)

			tx.Count(&levelCount)

			retrieveTx := tx.Offset(getParams.Offset).
				Limit(getParams.Limit).
				Find(&levels)

			err = retrieveTx.Error
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				context.Status(http.StatusNotFound)
				return
			}

			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"levels": levels,
			"total":  levelCount,
		})
	}
}

func levelValidate(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		if user.Role != model.UserRoleMod && user.Role != model.UserRoleAgent {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "no moderation authorization"})
			return
		}

		levelID, err := uuid.Parse(context.Param("levelId"))

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

		tx := db.Where("id = ?", user.ID, levelID).First(&level)

		if tx.Error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": tx.Error})
			return
		}

		validation := model.Validation{
			LevelID:      levelID,
			LevelVersion: level.Version,
			Result:       validateParams.ValidationResult,
			ValidatorID:  user.ID,
		}

		tx = db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "version"}, {Name: "level_id"}},
			UpdateAll: true,
		}).Create(&validation)

		if tx.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error})
			return
		}

		if level.Validation != nil {
			tx = db.Delete(&level.Validation)

			if tx.Error != nil {
				/*
					TODO: Old validation could not be deleted, would should happen ?

					context.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error})
					return
				*/
			}
		}

		level.Content = validateParams.Content
		level.ValidationId = &validation.ID
		level.AuthorScore = validateParams.AuthorScore
		level.Thumbnail = validateParams.Thumbnail
		level.Published = time.Now()
		level.Version += 1

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

		if err := context.BindQuery(&getParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		levels := []model.Level{}

		var levelCount int64

		tx := db.
			Model(&model.Level{}).
			Preload(clause.Associations).
			Where("user_id = ?", user.ID)

		tx.Count(&levelCount)

		retrieveTx := tx.Offset(getParams.Offset).
			Limit(getParams.Limit).
			Find(&levels)

		err := retrieveTx.Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				context.Status(http.StatusNotFound)
				return
			}

			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"levels": levels,
			"total":  levelCount,
		})
	}
}

func levelsAdd(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var levelAddParams levelParams

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

		levelID, err := uuid.Parse(context.Param("levelId"))

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		level := model.Level{
			ID: levelID,
		}

		tx := db

		if user.Role != model.UserRoleMod && user.Role != model.UserRoleAgent {
			tx = tx.Where("user_id = ?", user.ID)
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

func lockLevelValidation(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		if user.Role == model.UserRoleAgent {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Not authorized to lock validation for level"})
			return
		}

		levelID, err := uuid.Parse(context.Param("levelId"))

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var level model.Level

		tx := db.Where("id = ?", levelID).First(&level)

		if tx.Error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if level.Validation != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "level is already validated"})
			return
		}

		if level.ValidationLock.Before(time.Now().Add(time.Minute*time.Duration(config.C.TokenLifeSpan))) && user.ID != *level.ValidationAgentID {
			context.JSON(http.StatusBadRequest, gin.H{"error": "is in lock by another agent"})
			return
		}

		level.ValidationLock = time.Now()
		level.ValidationAgentID = &user.ID

		tx = db.Save(level)

		if tx.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		context.Status(http.StatusOK)
	}
}

func levelsUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		var updateParams levelParams

		levelID, err := uuid.Parse(context.Param("levelId"))

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := context.BindJSON(&updateParams); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var level model.Level

		tx := db.Where("user_id = ? AND id = ?", user.ID, levelID).First(&level)

		if tx.Error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		level.Name = updateParams.Name
		level.Content = updateParams.Content

		tx = db.Save(&level)

		if tx.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		context.Status(http.StatusOK)
	}
}

func levelVote(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		user := auth.GetJWTUser(context)

		levelID, err := uuid.Parse(context.Param("levelId"))

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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
		tx := db.Limit(1).Where("user_id!=? AND id = ?", user.ID, levelID).Find(&level)

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
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "level_id"}},
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

		levelID, err := uuid.Parse(context.Param("levelId"))

		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var level model.Level
		tx := db.Where("user_id != ? AND id = ?", user.ID, levelID).First(&level)

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

		context.JSON(http.StatusOK, gin.H{"reportId": report.ID})
	}
}

func UseLevel(router gin.IRouter, db *gorm.DB) {
	levelRouter := router.Group("/levels")

	levelRouter.GET("", levelsGetAll(db))
	//levelRouter.GET("/sus", levelsGetAllSus(db))
	router.GET("me/levels", levelsGetOwn(db))

	levelRouter.DELETE("/:levelId", levelsDelete(db))

	levelRouter.POST("", levelsAdd(db))

	levelRouter.PUT("/:levelId", levelsUpdate(db))
	levelRouter.PUT("/:levelId/reports", levelReport(db))
	levelRouter.PUT("/:levelId/vote", levelVote(db))
	levelRouter.PUT("/:levelId/validate", levelValidate(db))
}
