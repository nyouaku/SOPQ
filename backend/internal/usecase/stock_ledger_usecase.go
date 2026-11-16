// ===== FILE: backend/internal/usecase/stock_ledger_usecase.go =====
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nyouaku/SOPQ/backend/internal/domain"
	"github.com/nyouaku/SOPQ/backend/internal/repository"
)

type StockLedgerUsecase interface {
	// GetProductLedger mengembalikan kartu stok satu produk di satu gudang.
	GetProductLedger(ctx context.Context, productID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error)
	// GetWarehouseLedger mengembalikan riwayat ledger satu gudang.
	GetWarehouseLedger(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error)
	// GetLedgerByDateRange mengembalikan riwayat ledger antar tanggal.
	GetLedgerByDateRange(ctx context.Context, warehouseID *uuid.UUID, from, to time.Time, limit, offset int) ([]domain.StockLedger, int64, error)
	// RecordEntry mencatat satu entri ledger baru (untuk integrasi dari modul lain).
	RecordEntry(ctx context.Context, input RecordLedgerInput) (*domain.StockLedger, error)
}

type stockLedgerUsecase struct {
	ledgerRepo repository.StockLedgerRepository
	stockRepo  repository.StockRepository
}

func NewStockLedgerUsecase(
	ledgerRepo repository.StockLedgerRepository,
	stockRepo repository.StockRepository,
) StockLedgerUsecase {
	return &stockLedgerUsecase{
		ledgerRepo: ledgerRepo,
		stockRepo:  stockRepo,
	}
}

type RecordLedgerInput struct {
	ProductID       uuid.UUID
	WarehouseID     uuid.UUID
	TransactionType string
	ReferenceType   string
	ReferenceID     *uuid.UUID
	ReferenceCode   string
	QuantityIn      int
	QuantityOut     int
	UnitCost        float64
	Note            string
	CreatedBy       *uuid.UUID
}

func (u *stockLedgerUsecase) RecordEntry(ctx context.Context, input RecordLedgerInput) (*domain.StockLedger, error) {
	// Dapatkan saldo terakhir
	balance, err := u.ledgerRepo.GetBalance(ctx, input.ProductID, input.WarehouseID)
	if err != nil {
		return nil, err
	}
	balanceAfter := balance + input.QuantityIn - input.QuantityOut

	totalCost := float64(input.QuantityIn+input.QuantityOut) * input.UnitCost

	entry := &domain.StockLedger{
		ProductID:       input.ProductID,
		WarehouseID:     input.WarehouseID,
		TransactionType: input.TransactionType,
		ReferenceType:   input.ReferenceType,
		ReferenceID:     input.ReferenceID,
		ReferenceCode:   input.ReferenceCode,
		QuantityIn:      input.QuantityIn,
		QuantityOut:     input.QuantityOut,
		BalanceAfter:    balanceAfter,
		UnitCost:        input.UnitCost,
		TotalCost:       totalCost,
		Note:            input.Note,
		CreatedBy:       input.CreatedBy,
	}

	if err := u.ledgerRepo.Create(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (u *stockLedgerUsecase) GetProductLedger(ctx context.Context, productID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error) {
	return u.ledgerRepo.ListByProduct(ctx, productID, warehouseID, limit, offset)
}

func (u *stockLedgerUsecase) GetWarehouseLedger(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]domain.StockLedger, int64, error) {
	return u.ledgerRepo.ListByWarehouse(ctx, warehouseID, limit, offset)
}

func (u *stockLedgerUsecase) GetLedgerByDateRange(ctx context.Context, warehouseID *uuid.UUID, from, to time.Time, limit, offset int) ([]domain.StockLedger, int64, error) {
	return u.ledgerRepo.ListByDateRange(ctx, warehouseID, from, to, limit, offset)
}
