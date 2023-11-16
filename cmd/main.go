package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Lyretto/spooky-bodies-golang/internal/config"
	"github.com/Lyretto/spooky-bodies-golang/internal/controller"
	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("hello from spooky bodies server!")

	config.Init()

	os.Setenv("TOKEN_HOUR_LIFESPAN", strconv.Itoa(config.C.TokenLifeSpan))

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

	router := gin.New()

	corsConfig := cors.DefaultConfig()

	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization")

	router.Use(cors.New(corsConfig))

	controller.UseAuth(router, db)
	controller.UseLevel(router, db)

	router.Run("0.0.0.0:3000")
}
