package server

import (
	"github.com/Oloruntobi1/qgdc/internal/cache"
	"github.com/Oloruntobi1/qgdc/internal/database"
	"github.com/Oloruntobi1/qgdc/internal/middleware"
	"github.com/Oloruntobi1/qgdc/internal/token"

	"github.com/gin-gonic/gin"
)

type Server struct {
	repo       database.Repository
	tokenMaker token.Maker
	router     *gin.Engine
	cache      cache.Cacher
}

func NewServer(repo database.Repository, cache cache.Cacher, secret string) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(secret)
	if err != nil {
		panic(err)
	}
	server := &Server{
		repo:       repo,
		tokenMaker: tokenMaker,
		cache:      cache,
	}
	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.Use(middleware.LoggerToFile())

	v1Routes := router.Group("/api/v1/")

	userRoutes := v1Routes.Group("users")

	userRoutes.POST("", server.createUser)
	userRoutes.POST("/login", server.loginUser)
	userRoutes.GET("", server.getUsers)

	authRoutes := v1Routes.Group("wallets/").Use(middleware.AuthMiddleware(server.tokenMaker))
	authRoutes.GET(":wallet_id/balance", middleware.CacheMiddleware(server.cache), server.getWalletBalance)
	authRoutes.POST(":wallet_id/credit", server.creditWalletBalance)
	authRoutes.POST(":wallet_id/debit", server.debitWalletBalance)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"success": false,
		"message": "",
		"error":   err.Error(),
	}
}
