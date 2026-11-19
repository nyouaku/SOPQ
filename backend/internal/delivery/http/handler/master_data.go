// ===== FILE: backend/internal/delivery/http/handler/master_data.go =====
package handler

import (
	"net/http"
	"strconv"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/nyouaku/SOPQ/backend/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CategoryHandler
type CategoryHandler struct {
	categoryUsecase usecase.CategoryUsecase
}

func NewCategoryHandler(categoryUsecase usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{categoryUsecase: categoryUsecase}
}

type CreateCategoryRequest struct {
	Name     string     `json:"name" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	category := &domain.Category{
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	if err := h.categoryUsecase.CreateCategory(c.Request.Context(), category); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	utils.RespondCreated(c, category)
}

func (h *CategoryHandler) ListCategories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	categories, total, err := h.categoryUsecase.ListCategories(c.Request.Context(), limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"categories": categories,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// SupplierHandler
type SupplierHandler struct {
	supplierUsecase usecase.SupplierUsecase
}

func NewSupplierHandler(supplierUsecase usecase.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{supplierUsecase: supplierUsecase}
}

type CreateSupplierRequest struct {
	Code          string `json:"code" validate:"required"`
	Name          string `json:"name" validate:"required"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
}

func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	supplier := &domain.Supplier{
		Code:          req.Code,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Phone:         req.Phone,
		Email:         req.Email,
		Address:       req.Address,
		IsActive:      true,
	}

	if err := h.supplierUsecase.CreateSupplier(c.Request.Context(), supplier); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	utils.RespondCreated(c, supplier)
}

func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	suppliers, total, err := h.supplierUsecase.ListSuppliers(c.Request.Context(), limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"suppliers": suppliers,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// WarehouseHandler
type WarehouseHandler struct {
	warehouseUsecase usecase.WarehouseUsecase
}

func NewWarehouseHandler(warehouseUsecase usecase.WarehouseUsecase) *WarehouseHandler {
	return &WarehouseHandler{warehouseUsecase: warehouseUsecase}
}

type CreateWarehouseRequest struct {
	Code    string `json:"code" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Address string `json:"address"`
}

type CreateLocationRequest struct {
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func (h *WarehouseHandler) CreateWarehouse(c *gin.Context) {
	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	warehouse := &domain.Warehouse{
		Code:     req.Code,
		Name:     req.Name,
		Address:  req.Address,
		IsActive: true,
	}

	if err := h.warehouseUsecase.CreateWarehouse(c.Request.Context(), warehouse); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	utils.RespondCreated(c, warehouse)
}

func (h *WarehouseHandler) UpdateWarehouse(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}

	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	warehouse := &domain.Warehouse{
		ID:       warehouseID,
		Code:     req.Code,
		Name:     req.Name,
		Address:  req.Address,
		IsActive: true,
	}

	if err := h.warehouseUsecase.UpdateWarehouse(c.Request.Context(), warehouse); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	utils.RespondOK(c, warehouse)
}

func (h *WarehouseHandler) DeleteWarehouse(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}

	if err := h.warehouseUsecase.DeleteWarehouse(c.Request.Context(), warehouseID); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	utils.RespondOK(c, gin.H{"message": "Warehouse deleted"})
}

func (h *WarehouseHandler) ListWarehouses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	warehouses, total, err := h.warehouseUsecase.ListWarehouses(c.Request.Context(), limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"warehouses": warehouses,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

func (h *WarehouseHandler) CreateLocation(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}

	var req CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	location := &domain.Location{
		WarehouseID: warehouseID,
		Code:        req.Code,
		Name:        req.Name,
	}

	if err := h.warehouseUsecase.CreateLocation(c.Request.Context(), location); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	utils.RespondCreated(c, location)
}

func (h *WarehouseHandler) UpdateLocation(c *gin.Context) {
	locationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid location ID")
		return
	}

	var req CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	location := &domain.Location{
		ID:   locationID,
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.warehouseUsecase.UpdateLocation(c.Request.Context(), location); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	utils.RespondOK(c, location)
}

func (h *WarehouseHandler) DeleteLocation(c *gin.Context) {
	locationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid location ID")
		return
	}

	if err := h.warehouseUsecase.DeleteLocation(c.Request.Context(), locationID); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	utils.RespondOK(c, gin.H{"message": "Location deleted"})
}

func (h *WarehouseHandler) GetLocations(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid warehouse ID")
		return
	}

	locations, err := h.warehouseUsecase.GetLocationsByWarehouse(c.Request.Context(), warehouseID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{"locations": locations})
}

// ===== END FILE =====
