package server

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/Oloruntobi1/qgdc/internal/models"

	"github.com/Oloruntobi1/qgdc/util"

	"github.com/gin-gonic/gin"
)

var (
	ErrUserAlreadyExists = errors.New("user already with this email exists")
	// email or password incorrect
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type createUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// check if user already exists
	user, err := server.repo.GetUserByEmail(req.Email)
	if err != util.ErrUserNotFound {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if user != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrUserAlreadyExists))
		return
	}
	// hash password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// build user
	arg := &models.User{
		FullName:       req.FullName,
		Email:          req.Email,
		Password:       req.Password,
		HashedPassword: hashedPassword,
	}

	// create user
	err = server.repo.CreateUser(arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// create wallet for user in the background
	//go server.repo.CreateWallet(arg.ID)

	response := util.BuildResponseEntity(true, "User created successfully", nil)
	ctx.JSON(http.StatusCreated, response)
}

type loginUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// get user by email
	user, err := server.repo.GetUserByEmail(req.Email)
	if err != nil {
		if err == util.ErrUserNotFound {
			ctx.JSON(http.StatusBadRequest, errorResponse(ErrInvalidCredentials))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// verify password
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrInvalidCredentials))
		return
	}
	// get token duration from env
	duration := os.Getenv("ACCESS_TOKEN_DURATION")
	// convert string to time duration
	d, err := time.ParseDuration(duration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// create token
	accessToken, err := server.tokenMaker.CreateToken(
		user.Email,
		d,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := util.BuildResponseEntity(true, "", gin.H{
		"access_token": accessToken,
	})
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) getUsers(ctx *gin.Context) {
	users, err := server.repo.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := util.BuildResponseEntity(true, "", gin.H{
		"users": users,
	})
	ctx.JSON(http.StatusOK, response)
}
