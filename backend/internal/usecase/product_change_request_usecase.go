// ===== FILE: backend/internal/usecase/product_change_request_usecase.go =====
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
	"gorm.io/datatypes"
)

type SubmitChangeInput struct {
	ChangeType string
	ProductID  *uuid.UUID
	Payload    json.RawMessage
	Reason     string
	UserID     uuid.UUID
}

type ProductChangeRequestUsecase interface {
	Submit(ctx context.Context, input SubmitChangeInput) (*domain.ProductChangeRequest, error)
	List(ctx context.Context, status string, limit, offset int) ([]domain.ProductChangeRequest, int64, error)
	Approve(ctx context.Context, id, reviewerID uuid.UUID, note, ip string) error
	Reject(ctx context.Context, id, reviewerID uuid.UUID, note string) error
}

type productChangeRequestUsecase struct {
	reqRepo        repository.ProductChangeRequestRepository
	productUsecase ProductUsecase
	activityRepo   repository.ActivityLogRepository
}

func NewProductChangeRequestUsecase(
	reqRepo repository.ProductChangeRequestRepository,
	productUsecase ProductUsecase,
	activityRepo repository.ActivityLogRepository,
) ProductChangeRequestUsecase {
	return &productChangeRequestUsecase{
		reqRepo:        reqRepo,
		productUsecase: productUsecase,
		activityRepo:   activityRepo,
	}
}

func validChangeType(t string) bool {
	return t == domain.ChangeTypeCreate || t == domain.ChangeTypeUpdate || t == domain.ChangeTypeDelete
}

func (u *productChangeRequestUsecase) Submit(ctx context.Context, input SubmitChangeInput) (*domain.ProductChangeRequest, error) {
	if !validChangeType(input.ChangeType) {
		return nil, fmt.Errorf("change_type harus create/update/delete")
	}
	if len(input.Payload) == 0 {
		return nil, fmt.Errorf("payload tidak boleh kosong")
	}
	req := &domain.ProductChangeRequest{
		ChangeType:  input.ChangeType,
		ProductID:   input.ProductID,
		Payload:     datatypes.JSON(input.Payload),
		Reason:      input.Reason,
		Status:      domain.ChangeStatusPending,
		RequestedBy: input.UserID,
	}
	if err := u.reqRepo.Create(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (u *productChangeRequestUsecase) List(ctx context.Context, status string, limit, offset int) ([]domain.ProductChangeRequest, int64, error) {
	return u.reqRepo.List(ctx, repository.ChangeRequestFilter{Status: status, Limit: limit, Offset: offset})
}

// productFromPayload menerima bentuk ringan: field produk yang boleh diubah via approval.
func productFromPayload(raw datatypes.JSON) (*domain.Product, error) {
	var p struct {
		ProductCode  string     `json:"product_code"`
		Name         string     `json:"name"`
		Brand        string     `json:"brand"`
		Color        string     `json:"color"`
		Size         string     `json:"size"`
		Unit         string     `json:"unit"`
		LocationID   *uuid.UUID `json:"location_id"`
		MinimumStock int        `json:"minimum_stock"`
		CostPrice    float64    `json:"cost_price"`
		SupplierID   *uuid.UUID `json:"supplier_id"`
		Status       string     `json:"status"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("payload produk tidak valid: %w", err)
	}
	if p.Name == "" {
		return nil, fmt.Errorf("nama produk wajib diisi")
	}
	if p.ProductCode == "" {
		return nil, fmt.Errorf("product_code wajib diisi")
	}
	if p.Unit == "" {
		p.Unit = "PCS"
	}
	if p.Status == "" {
		p.Status = "active"
	}
	return &domain.Product{
		ProductCode:  p.ProductCode,
		Name:         p.Name,
		Brand:        p.Brand,
		Color:        p.Color,
		Size:         p.Size,
		Unit:         p.Unit,
		LocationID:   p.LocationID,
		MinimumStock: p.MinimumStock,
		CostPrice:    p.CostPrice,
		SupplierID:   p.SupplierID,
		Status:       p.Status,
	}, nil
}

func (u *productChangeRequestUsecase) Approve(ctx context.Context, id, reviewerID uuid.UUID, note, ip string) error {
	req, err := u.reqRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req == nil {
		return domain.ErrChangeRequestNotFound
	}
	if req.Status != domain.ChangeStatusPending {
		return domain.ErrChangeRequestNotPending
	}

	var entityID uuid.UUID
	switch req.ChangeType {
	case domain.ChangeTypeCreate:
		p, err := productFromPayload(req.Payload)
		if err != nil {
			return err
		}
		if err := u.productUsecase.CreateProduct(ctx, p); err != nil {
			return err
		}
		entityID = p.ID
	case domain.ChangeTypeUpdate:
		if req.ProductID == nil {
			return fmt.Errorf("product_id wajib untuk update")
		}
		current, err := u.productUsecase.GetProductByID(ctx, *req.ProductID)
		if err != nil {
			return err
		}
		if current == nil {
			return fmt.Errorf("produk tidak ditemukan")
		}
		p, err := productFromPayload(req.Payload)
		if err != nil {
			return err
		}
		p.ID = current.ID
		p.CreatedAt = current.CreatedAt
		if err := u.productUsecase.UpdateProduct(ctx, p); err != nil {
			return err
		}
		entityID = current.ID
	case domain.ChangeTypeDelete:
		if req.ProductID == nil {
			return fmt.Errorf("product_id wajib untuk delete")
		}
		if err := u.productUsecase.DeleteProduct(ctx, *req.ProductID); err != nil {
			return err
		}
		entityID = *req.ProductID
	}

	now := time.Now()
	req.Status = domain.ChangeStatusApproved
	req.ReviewedBy = &reviewerID
	req.ReviewedAt = &now
	req.ReviewNote = note
	req.UpdatedAt = now
	if err := u.reqRepo.Update(ctx, req); err != nil {
		return err
	}

	if u.activityRepo != nil {
		_ = u.activityRepo.Create(ctx, &domain.ActivityLog{
			UserID:     &reviewerID,
			Action:     "approve_product_change",
			EntityType: "product_change_request",
			EntityID:   &req.ID,
			NewValue:   datatypes.JSON(fmt.Sprintf(`{"change_type":"%s","entity_id":"%s"}`, req.ChangeType, entityID)),
			IPAddress:  ip,
		})
	}
	return nil
}

func (u *productChangeRequestUsecase) Reject(ctx context.Context, id, reviewerID uuid.UUID, note string) error {
	req, err := u.reqRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req == nil {
		return domain.ErrChangeRequestNotFound
	}
	if req.Status != domain.ChangeStatusPending {
		return domain.ErrChangeRequestNotPending
	}
	now := time.Now()
	req.Status = domain.ChangeStatusRejected
	req.ReviewedBy = &reviewerID
	req.ReviewedAt = &now
	req.ReviewNote = note
	req.UpdatedAt = now
	return u.reqRepo.Update(ctx, req)
}
