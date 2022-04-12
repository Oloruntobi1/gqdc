package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Oloruntobi1/qgdc/internal/models"
	"github.com/Oloruntobi1/qgdc/internal/token"
	"github.com/Oloruntobi1/qgdc/util"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalance          = errors.New("wallet balance cannot be less than zero")
	ErrInvalidAmount                = errors.New("amount cannot be less than zero")
	ErrWalletNotBelongsToUser       = errors.New("unauthenticated user cannot access wallet") // only an admin can access wallet of other users
	ErrAuthorizationPayloadNotFound = errors.New("authorization payload not found")
	ErrAuthorizationPayloadInvalid  = errors.New("authorization payload invalid")
)

type walletIDUriBinding struct {
	WalletID int64 `uri:"wallet_id" binding:"required,min=1"`
}

type WalletBalanceResponse struct {
	Balance string `json:"balance"`
}

func (server *Server) getWalletBalance(ctx *gin.Context) {
	wallet, err := server.verifyWalletBelongsToUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	cacheErr := server.cache.Set(ctx, fmt.Sprintf("%d", wallet.ID), wallet.Balance.String(), 100*time.Second)
	if cacheErr != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(cacheErr))
		return
	}
	response := util.BuildResponseEntity(true, "", gin.H{
		"balance": wallet.Balance.String(),
	})
	ctx.JSON(http.StatusOK, response)
}

type updateWalletRequest struct {
	// Amount float64 `json:"amount" binding:"required,min=1"`
	// NOTE: I decided not to use Gin binding for this request because
	// I want to use my function to validate the amount sent in the request
	// so as to fulfill the reuquirement of the assignment.
	Amount float64 `json:"amount"`
}

func (server *Server) creditWalletBalance(ctx *gin.Context) {
	var param walletIDUriBinding
	if err := ctx.ShouldBindUri(&param); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var req updateWalletRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// check if amount sent in request is negative
	if err := validateRequestAmount(req.Amount); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// get wallet balance
	wallet, err := server.repo.GetWallet(param.WalletID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	wallet.Balance = wallet.Balance.Add(decimal.NewFromFloat(req.Amount))
	wallet.UpdatedAt = time.Now()
	// update wallet balance
	w, err := server.repo.UpdateWallet(wallet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	cacheErr := server.cache.Set(ctx, fmt.Sprintf("%d", wallet.ID), wallet.Balance.String(), 100*time.Second)
	if cacheErr != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(cacheErr))
		return
	}
	response := util.BuildResponseEntity(true, util.WalletCreditSuccess, gin.H{
		"balance": w.Balance.String(),
	})
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) debitWalletBalance(ctx *gin.Context) {
	var param walletIDUriBinding
	if err := ctx.ShouldBindUri(&param); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var req updateWalletRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// check if amount sent in request is negative
	if err := validateRequestAmount(req.Amount); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// get wallet balance
	wallet, err := server.repo.GetWallet(param.WalletID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// check if the debit operation will cause the balance to be negative
	err = isWalletBalanceGoingBelowZero(wallet.Balance, req.Amount)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// update wallet balance
	wallet.Balance = wallet.Balance.Sub(decimal.NewFromFloat(req.Amount))
	wallet.UpdatedAt = time.Now()

	w, err := server.repo.UpdateWallet(wallet)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	cacheErr := server.cache.Set(ctx, fmt.Sprintf("%d", wallet.ID), wallet.Balance.String(), 100*time.Second)
	if cacheErr != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(cacheErr))
		return
	}
	response := util.BuildResponseEntity(true, util.WalletDebitSuccess, gin.H{
		"balance": w.Balance.String(),
	})
	ctx.JSON(http.StatusOK, response)
}

// utility function to check if amount sent in request is negative
func validateRequestAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// utility function to check if the debit operation on any given
// wallet balance will cause the balance to be negative
func isWalletBalanceGoingBelowZero(walletBalance decimal.Decimal, debitAmount float64) error {
	if walletBalance.Sub(decimal.NewFromFloat(debitAmount)).Cmp(decimal.Zero) < 0 {
		return ErrInsufficientBalance
	}
	return nil
}

// verify the wallet id belongs to logged in user
func (server *Server) verifyWalletBelongsToUser(ctx *gin.Context) (*models.Wallet, error) {
	userID, err := server.getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// get user by email
	user, err := server.repo.GetUserByEmail(userID)
	if err != nil {
		return nil, err
	}
	// get wallet by user id
	wallet, err := server.repo.GetWalletByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	var walletID walletIDUriBinding
	if err := ctx.ShouldBindUri(&walletID); err != nil {
		return nil, err
	}
	// check if wallet id belongs to user
	if wallet.ID != walletID.WalletID {
		return nil, ErrWalletNotBelongsToUser
	}
	return wallet, nil
}

// get user id from context
func (server *Server) getUserIDFromContext(ctx *gin.Context) (string, error) {
	payload, ok := ctx.Get("authorization_payload")
	if !ok {
		return "", ErrAuthorizationPayloadNotFound
	}
	payloadData, ok := payload.(*token.Payload)
	if !ok {
		return "", ErrAuthorizationPayloadInvalid
	}
	return payloadData.Email, nil
}
