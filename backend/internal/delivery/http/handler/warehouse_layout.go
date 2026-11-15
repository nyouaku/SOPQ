package handler

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UpdateLayout godoc
// @Summary Update warehouse layout (JSON)
// @Description Save visual layout coordinates for a warehouse
// @Tags warehouses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse ID"
// @Param body body object true "Layout JSON"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/warehouses/{id}/layout [put]
func (h *WarehouseHandler) UpdateLayout(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		warehouseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			utils.RespondBadRequest(c, "Invalid warehouse ID")
			return
		}

		var req struct {
			LayoutJSON map[string]any `json:"layout_json"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondBadRequest(c, "Invalid layout JSON")
			return
		}

		jsonBytes, err := json.Marshal(req.LayoutJSON)
		if err != nil {
			utils.RespondBadRequest(c, "Failed to marshal layout")
			return
		}

		result := db.WithContext(c.Request.Context()).
			Exec(
				"UPDATE warehouses SET layout_json = ?, updated_at = NOW() WHERE id = ?",
				datatypes.JSON(jsonBytes), warehouseID,
			)
		if result.Error != nil {
			utils.RespondInternalError(c, result.Error.Error())
			return
		}

		utils.RespondOK(c, gin.H{"message": "Layout saved"})
	}
}

// GetLayout godoc
// @Summary Get warehouse layout (JSON)
// @Description Get visual layout coordinates for a warehouse
// @Tags warehouses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Warehouse ID"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/warehouses/{id}/layout [get]
func (h *WarehouseHandler) GetLayout(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		warehouseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			utils.RespondBadRequest(c, "Invalid warehouse ID")
			return
		}

		var result struct {
			LayoutJSON datatypes.JSON `gorm:"column:layout_json"`
		}
		if err := db.WithContext(c.Request.Context()).
			Table("warehouses").
			Select("layout_json").
			Where("id = ?", warehouseID).
			Scan(&result).Error; err != nil {
			utils.RespondInternalError(c, err.Error())
			return
		}

		var layout map[string]any
		if len(result.LayoutJSON) > 0 {
			if err := json.Unmarshal(result.LayoutJSON, &layout); err != nil {
				utils.RespondInternalError(c, err.Error())
				return
			}
		}

		utils.RespondOK(c, gin.H{"layout_json": layout})
	}
}
