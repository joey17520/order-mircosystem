package controller

import (
	"github.com/gin-gonic/gin"
	"order-microsystem/api-service/internal/proxy"
)

type InventoryController struct {
	inventoryProxy *proxy.InventoryProxy
}

func NewInventoryController(inventoryController *proxy.InventoryProxy) *InventoryController {
	return &InventoryController{
		inventoryProxy: inventoryController,
	}
}

func (c *InventoryController) GetAllInventory(ctx *gin.Context) {
	var req struct {
		Offset int32 `form:"offset"`
		Limit  int32 `form:"limit"`
	}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.inventoryProxy.GetAllInventory(ctx.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{
		"inventory": resp,
	})
}
