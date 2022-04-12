package middleware

import (
	"errors"
	"net/http"

	"github.com/Oloruntobi1/qgdc/internal/cache"
	"github.com/Oloruntobi1/qgdc/util"

	"github.com/gin-gonic/gin"
)

// middleware to check redis cache for wallet balance
func CacheMiddleware(cacher cache.Cacher) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		walletID := ctx.Param("wallet_id")
		balance, err := cacher.Get(ctx, walletID)
		if errors.Is(err, cache.ErrNil) {
			ctx.Next()
		} else if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		} else {
			response := util.BuildResponseEntity(true, "", gin.H{
				"balance": balance,
			})
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}
}
