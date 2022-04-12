package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Oloruntobi1/qgdc/internal/cache"
	mockcache "github.com/Oloruntobi1/qgdc/internal/cache/mock"
	mockdb "github.com/Oloruntobi1/qgdc/internal/database/mock"
	"github.com/Oloruntobi1/qgdc/internal/middleware"
	"github.com/Oloruntobi1/qgdc/internal/models"
	"github.com/Oloruntobi1/qgdc/internal/token"
	"github.com/Oloruntobi1/qgdc/util"

	"bou.ke/monkey"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_validateRequestAmount(t *testing.T) {
	type args struct {
		amount float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should return error if amount is less than zero",
			args: args{
				amount: -1,
			},
			wantErr: true,
		},
		{
			name: "should return error if amount is zero",
			args: args{
				amount: 0,
			},
			wantErr: true,
		},
		{
			name: "should not return error if amount is greater than zero",
			args: args{
				amount: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = validateRequestAmount(tt.args.amount); (err != nil) != tt.wantErr {
				t.Errorf("validateRequestAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				require.Equal(t, ErrInvalidAmount, err)
			}
		})
	}
}

func Test_isWalletBalanceGoingBelowZero(t *testing.T) {
	type args struct {
		walletBalance decimal.Decimal
		debitAmount   float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should return error if debit amount is greater than wallet balance",
			args: args{
				walletBalance: decimal.NewFromFloat(100),
				debitAmount:   200,
			},
			wantErr: true,
		},
		{
			name: "should not return error if debit amount is equal to wallet balance",
			args: args{
				walletBalance: decimal.NewFromFloat(100),
				debitAmount:   100,
			},
			wantErr: false,
		},
		{
			name: "should not return error if debit amount is lesser than wallet balance",
			args: args{
				walletBalance: decimal.NewFromFloat(100),
				debitAmount:   70,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = isWalletBalanceGoingBelowZero(tt.args.walletBalance, tt.args.debitAmount); (err != nil) != tt.wantErr {
				t.Errorf("isWalletBalanceGoingBelowZero() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				require.Equal(t, ErrInsufficientBalance, err)
			}
		})
	}
}

func Test_getWalletBalance(t *testing.T) {
	user := randomUser()
	wallet := randomWallet(user.ID)

	testCases := []struct {
		name          string
		walletID      int64
		setUpAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(mockRepo *mockdb.MockRepository, mockCache *mockcache.MockCacher)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "should return wallet balance",
			walletID: wallet.ID,
			setUpAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(mockRepo *mockdb.MockRepository, mockCache *mockcache.MockCacher) {
				mockRepo.EXPECT().
					GetUserByEmail(gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)

				mockRepo.EXPECT().
					GetWalletByUserID(gomock.Eq(user.ID)).
					Times(1).
					Return(wallet, nil)

				mockCache.EXPECT().
					Get(gomock.Any(),
						fmt.Sprintf("%d", wallet.ID)).
					Times(1).
					Return("", cache.ErrNil)

				mockCache.EXPECT().
					Set(gomock.Any(),
						fmt.Sprintf("%d", wallet.ID),
						wallet.Balance.String(), 100*time.Second).
					Return(nil)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				response := buildResponse(wallet)
				requireBodyMatchResponse(t, recorder.Body, response)
			},
		},
	}

	for i := range testCases {
		tt := testCases[i]
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockdb.NewMockRepository(ctrl)
			cache := mockcache.NewMockCacher(ctrl)

			tt.buildStubs(repo, cache)

			// start test server and send request
			server, err := NewServer(repo, cache, util.RandomString(32))
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/wallets/%d/balance", tt.walletID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tt.setUpAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tt.checkResponse(t, recorder)

		})
	}

}

func randomWallet(userID int64) *models.Wallet {
	return &models.Wallet{
		ID:        1,
		UUID:      uuid.New(),
		UserID:    userID,
		Balance:   decimal.NewFromFloat(100),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func randomUser() *models.User {
	return &models.User{
		ID:        1,
		UUID:      uuid.New(),
		Email:     util.RandomEmail(),
		FullName:  util.RandomUserName(),
		Password:  util.RandomString(6),
		CreatedAt: time.Now(),
	}
}

func requireBodyMatchResponse(t *testing.T, body *bytes.Buffer, wantResponse *util.ResponseEntity) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	var gotResponse *util.ResponseEntity
	err = json.Unmarshal(data, &gotResponse)
	require.NoError(t, err)
	fmt.Printf("Actual response: %v", wantResponse)
	fmt.Printf("Got response: %v", gotResponse)
	require.Equal(t, wantResponse, gotResponse)
}

func buildResponse(wallet *models.Wallet) *util.ResponseEntity {
	return &util.ResponseEntity{
		Success: true,
		Message: "",
		Data: map[string]interface{}{
			"balance": wallet.Balance.String(),
		},
	}
}

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	email string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(email, duration)
	require.NoError(t, err)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(middleware.AuthorizationHeaderKey, authorizationHeader)
}

func Test_creditWalletBalance(t *testing.T) {
	user := randomUser()
	wallet := randomWallet(user.ID)

	var amount float64 = 200

	testCases := []struct {
		name          string
		walletID      int64
		body          gin.H
		setUpAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(mockRepo *mockdb.MockRepository, mockCache *mockcache.MockCacher)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "should credit wallet balance",
			walletID: wallet.ID,
			body: gin.H{
				"amount": amount,
			},
			setUpAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(mockRepo *mockdb.MockRepository, mockCache *mockcache.MockCacher) {
				mockRepo.EXPECT().
					GetWallet(gomock.Eq(wallet.ID)).
					Times(1).
					Return(wallet, nil)

				monkey.Patch(time.Now, func() time.Time {
					return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
				})

				arg := &models.Wallet{
					ID:        wallet.ID,
					UUID:      wallet.UUID,
					UserID:    wallet.UserID,
					Balance:   wallet.Balance.Add(decimal.NewFromFloat(amount)),
					CreatedAt: wallet.CreatedAt,
					UpdatedAt: time.Now(),
				}

				mockRepo.EXPECT().
					UpdateWallet(gomock.Eq(arg)).
					Times(1).
					Return(wallet, nil)

				mockCache.EXPECT().
					Set(gomock.Any(),
						fmt.Sprintf("%d", wallet.ID),
						wallet.Balance.Add(decimal.NewFromFloat(amount)).String(), 100*time.Second).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				response := buildCreditResponse(wallet)
				requireBodyMatchResponse(t, recorder.Body, response)
			},
		},
	}

	for i := range testCases {
		tt := testCases[i]
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockdb.NewMockRepository(ctrl)
			cache := mockcache.NewMockCacher(ctrl)

			tt.buildStubs(repo, cache)

			// start test server and send request
			server, err := NewServer(repo, cache, util.RandomString(32))
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tt.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/wallets/%d/credit", tt.walletID)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tt.setUpAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tt.checkResponse(t, recorder)
		})
	}
}

func buildCreditResponse(wallet *models.Wallet) *util.ResponseEntity {
	return &util.ResponseEntity{
		Success: true,
		Message: util.WalletCreditSuccess,
		Data: map[string]interface{}{
			"balance": wallet.Balance.String(),
		},
	}
}

func Test_debitWalletBalance(t *testing.T) {
	user := randomUser()
	wallet := randomWallet(user.ID)

	var amount float64 = 200

	testCases := []struct {
		name          string
		walletID      int64
		addToBalance  float64
		body          gin.H
		setUpAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(mockRepo *mockdb.MockRepository, mockCache *mockcache.MockCacher)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:         "should debit wallet balance",
			walletID:     wallet.ID,
			addToBalance: 500,
			body: gin.H{
				"amount": amount,
			},
			setUpAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Email, time.Minute)
			},
			buildStubs: func(mockRepo *mockdb.MockRepository, mockCache *mockcache.MockCacher) {
				mockRepo.EXPECT().
					GetWallet(gomock.Eq(wallet.ID)).
					Times(1).
					Return(wallet, nil)

				monkey.Patch(time.Now, func() time.Time {
					return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
				})

				arg := &models.Wallet{
					ID:        wallet.ID,
					UUID:      wallet.UUID,
					UserID:    wallet.UserID,
					Balance:   wallet.Balance.Sub(decimal.NewFromFloat(amount)),
					CreatedAt: wallet.CreatedAt,
					UpdatedAt: time.Now(),
				}

				mockRepo.EXPECT().
					UpdateWallet(gomock.Eq(arg)).
					Times(1).
					Return(wallet, nil)

				mockCache.EXPECT().
					Set(gomock.Any(),
						fmt.Sprintf("%d", wallet.ID),
						wallet.Balance.Sub(decimal.NewFromFloat(amount)).String(), 100*time.Second).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				response := buildDebitResponse(wallet)
				requireBodyMatchResponse(t, recorder.Body, response)
			},
		},
	}

	for i := range testCases {
		tt := testCases[i]
		t.Run(tt.name, func(t *testing.T) {

			wallet.Balance = wallet.Balance.Add(decimal.NewFromFloat(tt.addToBalance))

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockdb.NewMockRepository(ctrl)
			cache := mockcache.NewMockCacher(ctrl)

			tt.buildStubs(repo, cache)

			// start test server and send request
			server, err := NewServer(repo, cache, util.RandomString(32))
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tt.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/wallets/%d/debit", tt.walletID)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tt.setUpAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tt.checkResponse(t, recorder)
		})
	}
}

func buildDebitResponse(wallet *models.Wallet) *util.ResponseEntity {
	return &util.ResponseEntity{
		Success: true,
		Message: util.WalletDebitSuccess,
		Data: map[string]interface{}{
			"balance": wallet.Balance.String(),
		},
	}
}
