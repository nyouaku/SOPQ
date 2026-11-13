package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nyouaku/SOPQ/backend/internal/usecase"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
)

const maxImportFileSize = 5 << 20 // 5 MB

type ImportHandler struct {
	importUsecase usecase.ImportUsecase
}

func NewImportHandler(importUsecase usecase.ImportUsecase) *ImportHandler {
	return &ImportHandler{importUsecase: importUsecase}
}

// DownloadProductTemplate godoc
// @Summary Download template import produk
// @Tags import
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Router /api/v1/import/products/template [get]
func (h *ImportHandler) DownloadProductTemplate(c *gin.Context) {
	file, filename, err := h.importUsecase.ProductImportTemplate()
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")

	if err := file.Write(c.Writer); err != nil {
		utils.RespondInternalError(c, err.Error())
	}
}

// ImportProducts godoc
// @Summary Import produk dari file Excel
// @Description Upload file .xlsx sesuai template. product_code yang sudah ada akan diperbarui.
// @Tags import
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File Excel (.xlsx)"
// @Router /api/v1/import/products [post]
func (h *ImportHandler) ImportProducts(c *gin.Context) {
	// Batasi ukuran body sebelum file disimpan ke memori.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImportFileSize)

	header, err := c.FormFile("file")
	if err != nil {
		utils.RespondBadRequest(c, "File tidak ditemukan pada field 'file'")
		return
	}
	if header.Size > maxImportFileSize {
		utils.RespondBadRequest(c, "Ukuran file maksimal 5 MB")
		return
	}

	src, err := header.Open()
	if err != nil {
		utils.RespondBadRequest(c, "File tidak bisa dibuka")
		return
	}
	defer src.Close()

	result, err := h.importUsecase.ImportProducts(c.Request.Context(), src)
	if err != nil {
		utils.RespondBadRequest(c, err.Error())
		return
	}

	utils.RespondOK(c, result)
}
