package main

import (
	"fmt"

	"github.com/Lyretto/spooky-bodies-golang/internal/config"
	"github.com/Lyretto/spooky-bodies-golang/internal/controller"
	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	fmt.Println("hello from spooky bodies server!")

	config.Init()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Berlin",
		config.C.Database.Host,
		config.C.Database.Port,
		config.C.Database.User,
		config.C.Database.Password,
		config.C.Database.DatabaseName,
	)

	db, err := gorm.Open(postgres.Open(dsn))

	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&model.User{},
		&model.Level{},
		&model.Vote{},
		&model.Validation{},
		&model.Report{},
		&model.UserToken{},
	)

	// debug: ensure some user with uuid e97b3095-f92c-4d3e-a88b-25f2a4761c4a

	debugUser := model.User{
		ID:             uuid.MustParse("e97b3095-f92c-4d3e-a88b-25f2a4761c4a"),
		PlatformType:   model.PlatformSteam,
		PlatformUserID: "debug-platform-user-id",
	}

	tx := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&debugUser)

	if tx.Error != nil {
		panic(tx.Error)
	}

	router := gin.New()

	corsConfig := cors.DefaultConfig()

	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization")

	router.Use(cors.New(corsConfig))

	router.GET("/levels", controller.LevelsGetAll(db))
	router.POST("/levels", controller.LevelsAdd(db))
	router.PUT("/levels/{levelId}", controller.LevelsUpdate(db))
	router.DELETE("/levels/{levelId}", controller.LevelsDelete(db))
	router.PUT("/levels/{levelId}/reports", controller.LevelReport(db))
	router.PUT("/levels/{levelId}/votes", controller.LevelVote(db))

	router.Run("0.0.0.0:3000")
}
