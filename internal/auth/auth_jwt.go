package auth

import (
	"net/http"
	"time"

	"github.com/Lyretto/spooky-bodies-golang/internal/config"
	"github.com/Lyretto/spooky-bodies-golang/pkg/model"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type login struct {
	PlatformType   model.PlatformType `json:"platformType"`
	PlatformUserID string             `json:"username"`
	PlatformToken  string             `json:"platfromToken"`
}

var jwtIdentityKey = "userId"

func GetJWTUser(ctx *gin.Context) *model.User {
	u, exists := ctx.Get(jwtIdentityKey)

	if !exists {
		return nil
	}

	user, ok := u.(*model.User)

	if !ok {
		return nil
	}

	return user
}

func CheckTokenActivity(context *gin.Context, db *gorm.DB) (bool, *model.UserToken, error) {
	token := jwt.GetToken(context)

	if token == "" {
		context.AbortWithStatus(http.StatusUnauthorized)
		return false, nil, nil
	}

	var userToken model.UserToken
	tx := db.Where(&model.UserToken{Token: token}).Find(&userToken)

	if tx.Error != nil {
		context.AbortWithStatus(http.StatusUnauthorized)
		return false, nil, tx.Error
	}

	return true, &userToken, nil
}

func GetJWTMiddleware(db *gorm.DB) (*jwt.GinJWTMiddleware, error) {
	// this lib doesn't provide a proper refresh token. this is alright for now but it should be replaced
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:            "spooky-bodies",
		SigningAlgorithm: "HS512",
		Key:              []byte(config.C.JWTKey),
		Timeout:          time.Hour,
		MaxRefresh:       time.Hour,
		IdentityKey:      jwtIdentityKey,
		TokenLookup:      "header: Authorization",
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginParams login

			if err := c.Bind(&loginParams); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			var user model.User

			tx := db.Where(&model.User{PlatformType: loginParams.PlatformType, PlatformUserID: loginParams.PlatformUserID}).Find(&user)

			if tx.Error != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			switch user.PlatformType {
			case model.PlatformSteam:
				return nil, jwt.ErrFailedAuthentication
			case model.PlatformNintendo:
				return nil, jwt.ErrFailedAuthentication
			case model.PlatformNone:
				c.Set("user", user)
			default:
				return nil, jwt.ErrFailedAuthentication
			}

			return user, nil
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return jwt.MapClaims{
					jwtIdentityKey: v.ID,
				}
			}

			return jwt.MapClaims{}
		},
		RefreshResponse: func(c *gin.Context, code int, jwtToken string, validUntil time.Time) {
			_, userToken, _ := CheckTokenActivity(c, db)

			if userToken == nil {
				return
			}

			userToken.Token = jwtToken
			userToken.ValidUntil = validUntil

			if err := db.Save(userToken).Error; err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.JSON(code, gin.H{
				"token": jwtToken,
			})
		},
		LoginResponse: func(c *gin.Context, code int, jwtToken string, validUntil time.Time) {
			user, userExists := c.Get("user")

			if !userExists {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			u, ok := user.(*model.User)

			if !ok {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			userToken := model.UserToken{
				User:       u,
				Token:      jwtToken,
				ValidUntil: validUntil,
			}

			if err := db.Save(&userToken).Error; err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.JSON(code, gin.H{
				"token": jwtToken,
			})
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)

			var user model.User

			tx := db.Where(&model.User{ID: uuid.MustParse(claims[jwtIdentityKey].(string))}).Find(&user)

			if tx.Error != nil {
				return nil
			}

			return user
		},
	})
}
