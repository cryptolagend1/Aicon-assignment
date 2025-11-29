package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"Aicon-assignment/internal/domain/entity"
	domainErrors "Aicon-assignment/internal/domain/errors"
	"Aicon-assignment/internal/usecase"
)

// MockItemUsecase is a mock implementation of ItemUsecase
type MockItemUsecase struct {
	mock.Mock
}

func (m *MockItemUsecase) GetAllItems(ctx context.Context) ([]*entity.Item, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) GetItemByID(ctx context.Context, id int64) (*entity.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) CreateItem(ctx context.Context, input usecase.CreateItemInput) (*entity.Item, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) UpdateItem(ctx context.Context, id int64, input usecase.UpdateItemInput) (*entity.Item, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) DeleteItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemUsecase) GetCategorySummary(ctx context.Context) (*usecase.CategorySummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.CategorySummary), args.Error(1)
}

func TestItemHandler_UpdateItem(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		requestBody    interface{}
		setupMock      func(*MockItemUsecase)
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: nameのみ更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "更新された名前",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("更新された名前", "時計", "初期ブランド", 100000, "2023-01-01")
				updatedItem.ID = 1
				updatedItem.CreatedAt = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				updatedItem.UpdatedAt = time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.MatchedBy(func(input usecase.UpdateItemInput) bool {
					return input.Name != nil && *input.Name == "更新された名前" &&
						input.Brand == nil && input.PurchasePrice == nil
				})).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var item entity.Item
				err := json.Unmarshal(rec.Body.Bytes(), &item)
				require.NoError(t, err)
				assert.Equal(t, "更新された名前", item.Name)
				assert.Equal(t, int64(1), item.ID)
			},
		},
		{
			name: "正常系: brandのみ更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"brand": "更新されたブランド",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("初期アイテム", "時計", "更新されたブランド", 100000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.MatchedBy(func(input usecase.UpdateItemInput) bool {
					return input.Brand != nil && *input.Brand == "更新されたブランド" &&
						input.Name == nil && input.PurchasePrice == nil
				})).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var item entity.Item
				err := json.Unmarshal(rec.Body.Bytes(), &item)
				require.NoError(t, err)
				assert.Equal(t, "更新されたブランド", item.Brand)
			},
		},
		{
			name: "正常系: purchase_priceのみ更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"purchase_price": 200000,
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("初期アイテム", "時計", "初期ブランド", 200000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.MatchedBy(func(input usecase.UpdateItemInput) bool {
					return input.PurchasePrice != nil && *input.PurchasePrice == 200000 &&
						input.Name == nil && input.Brand == nil
				})).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var item entity.Item
				err := json.Unmarshal(rec.Body.Bytes(), &item)
				require.NoError(t, err)
				assert.Equal(t, 200000, item.PurchasePrice)
			},
		},
		{
			name: "正常系: 複数フィールド更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"name":           "新しい名前",
				"brand":          "新しいブランド",
				"purchase_price": 300000,
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("新しい名前", "時計", "新しいブランド", 300000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.MatchedBy(func(input usecase.UpdateItemInput) bool {
					return input.Name != nil && *input.Name == "新しい名前" &&
						input.Brand != nil && *input.Brand == "新しいブランド" &&
						input.PurchasePrice != nil && *input.PurchasePrice == 300000
				})).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var item entity.Item
				err := json.Unmarshal(rec.Body.Bytes(), &item)
				require.NoError(t, err)
				assert.Equal(t, "新しい名前", item.Name)
				assert.Equal(t, "新しいブランド", item.Brand)
				assert.Equal(t, 300000, item.PurchasePrice)
			},
		},
		{
			name: "異常系: 無効なID形式",
			id:   "invalid",
			requestBody: map[string]interface{}{
				"name": "更新された名前",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid item ID",
		},
		{
			name: "異常系: 無効なJSON形式",
			id:   "1",
			requestBody: "invalid json",
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request format",
		},
		{
			name: "異常系: フィールドが全てnil",
			id:   "1",
			requestBody: map[string]interface{}{},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name: "異常系: nameが空文字",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name: "異常系: nameが100文字超過",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "ロレックス デイトナ 16520 18K イエローゴールド ブラック文字盤 自動巻き クロノグラフ メンズ 腕時計 1988年製 ヴィンテージ 希少 コレクション アイテム",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name: "異常系: purchase_priceが負の値",
			id:   "1",
			requestBody: map[string]interface{}{
				"purchase_price": -1,
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name: "異常系: アイテムが見つからない",
			id:   "999",
			requestBody: map[string]interface{}{
				"name": "更新された名前",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				mockUsecase.On("UpdateItem", mock.Anything, int64(999), mock.Anything).Return((*entity.Item)(nil), domainErrors.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "item not found",
		},
		{
			name: "異常系: バリデーションエラー（use case層）",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "更新された名前",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.Anything).Return((*entity.Item)(nil), domainErrors.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name: "異常系: サーバーエラー",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "更新された名前",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.Anything).Return((*entity.Item)(nil), domainErrors.ErrDatabaseError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			mockUsecase := new(MockItemUsecase)
			tt.setupMock(mockUsecase)

			handler := NewItemHandler(mockUsecase)

			// Create request body
			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPatch, "/items/"+tt.id, bytes.NewBuffer(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/items/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			// Execute handler
			err = handler.UpdateItem(c)

			// Assertions
			// Echo's c.JSON() writes the response and returns nil on success
			// For error cases, the handler still returns nil (c.JSON writes the error response)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus >= 400 {
				// Check error response body
				if rec.Body.Len() > 0 {
					var errorResp ErrorResponse
					err := json.Unmarshal(rec.Body.Bytes(), &errorResp)
					if err == nil {
						assert.Contains(t, errorResp.Error, tt.expectedError)
					}
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResponse != nil {
					tt.checkResponse(t, rec)
				}
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}

