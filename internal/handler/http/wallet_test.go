package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AntonioForYou/wallet-api/internal/domain"
	"github.com/AntonioForYou/wallet-api/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockWalletRepo struct {
	mockBalance int64
	mockErr     error
}

func (m *mockWalletRepo) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	return m.mockBalance, m.mockErr
}

func (m *mockWalletRepo) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) (int64, error) {
	return m.mockBalance, m.mockErr
}

type mockDispatcher struct {
	dispatchFunc func(job domain.Job)
}

func (m *mockDispatcher) Dispatch(job domain.Job) {
	if m.dispatchFunc != nil {
		m.dispatchFunc(job)
	}
}

func TestHandleGetBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validUUID := uuid.New()

	tests := []struct {
		name           string
		walletID       string
		mockBalance    int64
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Успешное получение баланса",
			walletID:       validUUID.String(),
			mockBalance:    1500,
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Неверный формат UUID",
			walletID:       "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Кошелек не найден",
			walletID:       validUUID.String(),
			mockErr:        postgres.ErrWalletNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(nil, &mockWalletRepo{mockBalance: tt.mockBalance, mockErr: tt.mockErr})
			r := gin.Default()
			handler.RegisterRoutes(r)

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/wallets/"+tt.walletID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleDepositWithdraw_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		reqBody        interface{}
		expectedStatus int
	}{
		{
			name:           "Ошибка: нулевая сумма",
			reqBody:        domain.WalletRequest{WalletID: uuid.New(), OperationType: domain.Deposit, Amount: 0},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: неизвестная операция",
			reqBody:        map[string]interface{}{"walletId": uuid.New(), "operationType": "INVALID", "amount": 100},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(nil, &mockWalletRepo{})
			r := gin.Default()
			handler.RegisterRoutes(r)

			jsonBytes, _ := json.Marshal(tt.reqBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleDepositWithdraw_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDispatcher := &mockDispatcher{
		dispatchFunc: func(job domain.Job) {
			job.ResultChan <- domain.Result{NewBalance: 1500, Err: nil}
		},
	}

	handler := NewHandler(mockDispatcher, &mockWalletRepo{})
	r := gin.Default()
	handler.RegisterRoutes(r)

	reqBody := domain.WalletRequest{
		WalletID:      uuid.New(),
		OperationType: domain.Deposit,
		Amount:        1000,
	}
	jsonBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"balance":1500`)
}
