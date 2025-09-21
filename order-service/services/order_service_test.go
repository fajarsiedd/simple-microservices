package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"order-service/entities"
	"order-service/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrderService_FindByProductID(t *testing.T) {
	productID := uint(123)
	cacheKey := fmt.Sprintf("orders:productid:%d", productID)
	cacheTTL := 5 * time.Minute

	expectedOrders := []entities.Order{
		{ID: 1, ProductID: productID, Qty: 2, Status: "completed"},
		{ID: 2, ProductID: productID, Qty: 1, Status: "completed"},
	}
	jsonOrders, _ := json.Marshal(expectedOrders)

	t.Run("should return orders from cache on cache hit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockOrderRepository(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
		mockMessaging := mocks.NewMockMessagingService(ctrl)
		mockCache := mocks.NewMockCacheService(ctrl)
		s := NewOrderService(mockRepo, mockHTTPClient, mockMessaging, mockCache)

		// Expect get cache success
		mockCache.EXPECT().Get(cacheKey).Return(string(jsonOrders), nil)

		// Expect repo should have not been called
		mockRepo.EXPECT().FindByProductID(gomock.Any()).Times(0)

		orders, err := s.FindByProductID(productID)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, orders)
	})

	t.Run("should get from repo and set cache on cache miss", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockOrderRepository(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
		mockMessaging := mocks.NewMockMessagingService(ctrl)
		mockCache := mocks.NewMockCacheService(ctrl)
		s := NewOrderService(mockRepo, mockHTTPClient, mockMessaging, mockCache)

		// Expect get cache failed or empty
		mockCache.EXPECT().Get(cacheKey).Return("", errors.New("cache miss"))

		// Expect call repo succeeded
		mockRepo.EXPECT().FindByProductID(productID).Return(expectedOrders, nil)

		// Expect save data to redis cache
		mockCache.EXPECT().SetWithTTL(cacheKey, string(jsonOrders), cacheTTL).Return(nil)

		orders, err := s.FindByProductID(productID)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, orders)
	})

	t.Run("should return error on repo failure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockOrderRepository(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
		mockMessaging := mocks.NewMockMessagingService(ctrl)
		mockCache := mocks.NewMockCacheService(ctrl)
		s := NewOrderService(mockRepo, mockHTTPClient, mockMessaging, mockCache)

		expectedErr := errors.New("db connection error")

		// Expect get cache failed
		mockCache.EXPECT().Get(cacheKey).Return("", errors.New("cache miss"))

		// Expect call repo failed
		mockRepo.EXPECT().FindByProductID(productID).Return([]entities.Order{}, expectedErr)

		// Expect set cache should not have been called
		mockCache.EXPECT().SetWithTTL(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		orders, err := s.FindByProductID(productID)

		assert.Nil(t, orders)
		assert.EqualError(t, err, expectedErr.Error())
	})
}
