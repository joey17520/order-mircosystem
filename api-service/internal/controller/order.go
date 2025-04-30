package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"order-microsystem/api-service/internal/domain/model"
	"order-microsystem/api-service/internal/proxy"
)

type OrderController struct {
	orderProxy *proxy.OrderProxy
}

func NewOrderController(orderProxy *proxy.OrderProxy) *OrderController {
	return &OrderController{
		orderProxy: orderProxy,
	}
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req model.CreateOrderReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stdCtx := ctx.Request.Context()
	resp, err := c.orderProxy.CreateOrder(&stdCtx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"order": resp})
}
