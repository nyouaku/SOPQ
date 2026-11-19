// ===== FILE: backend/internal/delivery/http/handler/product.go =====
package handler

import (
	"net/http"
	"strconv"

	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/middleware"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"github.com/nyouaku/SOPQ/backend/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productUsecase usecase.ProductUsecase
	// stockRepo & ledgerRepo dipakai untuk membuat stok awal + baris ledger 'initial'
	// saat menambah produk baru dengan stok awal.
	stockRepo  repository.StockRepository
	ledgerRepo repository.StockLedgerRepository
}

func NewProductHandler(productUsecase usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
	}
}

// SetStockRepos menyuntikkan stock + ledger repo (opsional). Dipanggil dari wiring
// routes.go agar handler bisa membuat stok awal saat menambah produk baru.
func (h *ProductHandler) SetStockRepos(stockRepo repository.StockRepository, ledgerRepo repository.StockLedgerRepository) {
	h.stockRepo = stockRepo
	h.ledgerRepo = ledgerRepo
}

// DTOs
type CreateProductRequest struct {
	ProductCode  string     `json:"product_code" validate:"required"`
	Name         string     `json:"name" validate:"required"`
	Brand        string     `json:"brand"`
	Color        string     `json:"color"`
	Size         string     `json:"size"`
	Unit         string     `json:"unit" validate:"required"`
	LocationID   *uuid.UUID `json:"location_id"`
	MinimumStock int        `json:"minimum_stock"`
	SupplierID   *uuid.UUID `json:"supplier_id"`
	Status       string     `json:"status"`
	// Stok awal (opsional): bila diisi, buat baris stock di gudang tsb.
	InitialWarehouseID string `json:"initial_warehouse_id"`
	InitialStock       int    `json:"initial_stock"`
}

type UpdateProductRequest struct {
	Name         string     `json:"name" validate:"required"`
	Brand        string     `json:"brand"`
	Color        string     `json:"color"`
	Size         string     `json:"size"`
	Unit         string     `json:"unit" validate:"required"`
	LocationID   *uuid.UUID `json:"location_id"`
	MinimumStock int        `json:"minimum_stock"`
	SupplierID   *uuid.UUID `json:"supplier_id"`
	Status       string     `json:"status"`
}

type ProductListResponse struct {
	Products []domain.Product `json:"products"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	Limit    int              `json:"limit"`
}

// CreateProduct godoc
// @Summary Create product
// @Description Create new product (admin only)
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProductRequest true "Product data"
// @Success 201 {object} utils.SuccessResponse{data=domain.Product}
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	product := &domain.Product{
		ProductCode:  req.ProductCode,
		Name:         req.Name,
		Brand:        req.Brand,
		Color:        req.Color,
		Size:         req.Size,
		Unit:         req.Unit,
		LocationID:   req.LocationID,
		MinimumStock: req.MinimumStock,
		SupplierID:   req.SupplierID,
		Status:       req.Status,
	}

	if product.Status == "" {
		product.Status = "active"
	}

	if err := h.productUsecase.CreateProduct(c.Request.Context(), product); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	// Ambil userID (opsional) utk pencipta baris ledger.
	userID, _ := middleware.GetUserIDFromContext(c)

	// Stok awal (opsional): bila initial_warehouse_id + initial_stock diisi, buat baris
	// stock di gudang tsb + catat ke ledger sebagai 'initial'. Gagal ledger tak menggagalkan
	// pembuatan produk (produk + stok sudah tersimpan).
	if req.InitialWarehouseID != "" && req.InitialStock > 0 && h.stockRepo != nil && h.ledgerRepo != nil {
		whID, werr := uuid.Parse(req.InitialWarehouseID)
		if werr != nil {
			utils.RespondBadRequest(c, "Invalid initial_warehouse_id")
			return
		}
		newStock := &domain.Stock{
			ProductID:      product.ID,
			WarehouseID:    whID,
			QuantitySystem: req.InitialStock,
			Version:        0,
		}
		if err := h.stockRepo.CreateOrUpdate(c.Request.Context(), newStock); err != nil {
			utils.RespondError(c, http.StatusBadRequest, "INITIAL_STOCK_FAILED", err.Error())
			return
		}
		_ = h.ledgerRepo.Create(c.Request.Context(), &domain.StockLedger{
			ProductID:       product.ID,
			WarehouseID:     whID,
			TransactionType: domain.LedgerTypeInitial,
			ReferenceType:   "inventory",
			QuantityIn:      req.InitialStock,
			BalanceAfter:    req.InitialStock,
			Note:            "Stok awal saat tambah produk",
			CreatedBy:       &userID,
		})
	}

	utils.RespondCreated(c, product)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get product details
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} utils.SuccessResponse{data=domain.Product}
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}

	product, err := h.productUsecase.GetProductByID(c.Request.Context(), id)
	if err != nil {
		utils.RespondNotFound(c, err.Error())
		return
	}

	utils.RespondOK(c, product)
}

// GetProductByCode godoc
// @Summary Get product by code
// @Description Get product by product_code (isi QR, kritikal untuk scan, < 500ms)
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param code path string true "Product code (isi QR)"
// @Success 200 {object} utils.SuccessResponse{data=domain.Product}
// @Router /api/v1/products/code/{code} [get]
func (h *ProductHandler) GetProductByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.RespondBadRequest(c, "Code is required")
		return
	}

	product, err := h.productUsecase.GetProductByCode(c.Request.Context(), code)
	if err != nil {
		utils.RespondNotFound(c, "product not found")
		return
	}

	utils.RespondOK(c, product)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update product (admin only)
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body UpdateProductRequest true "Product data"
// @Success 200 {object} utils.SuccessResponse{data=domain.Product}
// @Router /api/v1/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	// Muat produk lama: Save() menulis SEMUA kolom, jadi field yang tidak dikirim
	// (product_code, photo_url, dll) jangan sampai ter-overwrite jadi kosong.
	product, err := h.productUsecase.GetProductByID(c.Request.Context(), id)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	if product == nil {
		utils.RespondNotFound(c, "Product not found")
		return
	}

	product.Name = req.Name
	product.Brand = req.Brand
	product.Color = req.Color
	product.Size = req.Size
	product.Unit = req.Unit
	product.LocationID = req.LocationID
	product.MinimumStock = req.MinimumStock
	product.SupplierID = req.SupplierID
	if req.Status != "" {
		product.Status = req.Status
	}

	if err := h.productUsecase.UpdateProduct(c.Request.Context(), product); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	// Reload with associations
	updated, err := h.productUsecase.GetProductByID(c.Request.Context(), id)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	utils.RespondOK(c, updated)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Soft delete product (admin only)
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondBadRequest(c, "Invalid product ID")
		return
	}

	if err := h.productUsecase.DeleteProduct(c.Request.Context(), id); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	utils.RespondOK(c, gin.H{"message": "Product deleted successfully"})
}

// ListProducts godoc
// @Summary List products
// @Description List products with pagination and filters
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by name or product code"
// @Param supplier_id query string false "Filter by supplier ID"
// @Param status query string false "Filter by status"
// @Success 200 {object} utils.SuccessResponse{data=ProductListResponse}
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	filter := repository.ProductFilter{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Limit:  limit,
		Offset: offset,
	}

	if supplierID := c.Query("supplier_id"); supplierID != "" {
		id, err := uuid.Parse(supplierID)
		if err == nil {
			filter.SupplierID = &id
		}
	}

	products, total, err := h.productUsecase.ListProducts(c.Request.Context(), filter)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	response := ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	utils.RespondOK(c, response)
}

// TransferProducts godoc
// @Summary Transfer products to a different location/area
// @Description Move multiple products to a new storage location in one operation
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "Transfer request"
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/products/transfer [post]
func (h *ProductHandler) TransferProducts(c *gin.Context) {
	var req struct {
		ProductIDs   []string `json:"product_ids" validate:"required,min=1"`
		ToLocationID string   `json:"to_location_id" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "Invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		utils.RespondValidationError(c, validator.FormatValidationError(err))
		return
	}

	// Ambil user ID dari context (JWT). Wajib lewat helper middleware — middleware
	// menyimpan claims di key "user" (bukan "user_id"), jadi c.Get("user_id") selalu
	// kosong dan endpoint ini sempat selalu 401 walau token valid.
	userUUID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	toLocationID, err := uuid.Parse(req.ToLocationID)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid target location ID")
		return
	}

	productIDs := make([]uuid.UUID, 0, len(req.ProductIDs))
	for _, raw := range req.ProductIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			utils.RespondBadRequest(c, "Invalid product ID: "+raw)
			return
		}
		productIDs = append(productIDs, id)
	}

	affected, err := h.productUsecase.TransferProducts(c.Request.Context(), productIDs, toLocationID, userUUID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	utils.RespondOK(c, gin.H{
		"transferred": affected,
	})
}
// ===== END FILE =====
