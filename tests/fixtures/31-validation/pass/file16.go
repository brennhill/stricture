// service/product_service.go — Service with excessive LOC (1200 total).
package service

import (
	"context"
	"errors"
	"super-lint-test/internal/domain"
	"super-lint-test/pkg/logging"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrDuplicateSKU        = errors.New("duplicate SKU")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrVariantNotFound     = errors.New("variant not found")
	ErrInvalidDiscount     = errors.New("invalid discount")
	ErrPromotionExpired    = errors.New("promotion expired")
)

// ProductRepository interface (20 methods).
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters ProductFilters) ([]*domain.Product, error)
	// ... 14 more repository methods ...
}

// ProductService implements product business logic.
type ProductService struct {
	repo               ProductRepository
	inventoryService   *InventoryService
	pricingService     *PricingService
	categoryService    *CategoryService
	variantService     *VariantService
	promotionService   *PromotionService
	logger             logging.Logger
}

// NewProductService creates product service.
func NewProductService(
	repo ProductRepository,
	inventorySvc *InventoryService,
	pricingSvc *PricingService,
	categorySvc *CategoryService,
	variantSvc *VariantService,
	promotionSvc *PromotionService,
	logger logging.Logger,
) *ProductService {
	return &ProductService{
		repo:             repo,
		inventoryService: inventorySvc,
		pricingService:   pricingSvc,
		categoryService:  categorySvc,
		variantService:   variantSvc,
		promotionService: promotionSvc,
		logger:           logger,
	}
}

// The file continues with 100+ methods (each 10-15 lines):
// CreateProduct, GetProduct, UpdateProduct, DeleteProduct, ListProducts, SearchProducts,
// AddCategory, RemoveCategory, UpdateCategory, ListCategories,
// AddVariant, RemoveVariant, UpdateVariant, ListVariants, GetVariantBySKU,
// AddPromotion, RemovePromotion, UpdatePromotion, ListPromotions, ApplyPromotion,
// UpdateInventory, CheckStock, ReserveStock, ReleaseStock, GetStockLevel,
// UpdatePrice, GetPrice, CalculateDiscount, ApplyBulkDiscount,
// GenerateSKU, ValidateSKU, ImportProducts, ExportProducts,
// GetProductAnalytics, GetTopSellingProducts, GetLowStockProducts,
// ... (80+ more methods) ...
// Total: 1200 lines

func (s *ProductService) CreateProduct(ctx context.Context, name, sku string, price int64) (*domain.Product, error) {
	// Validation and business logic (15 lines)
	return nil, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	// Retrieval logic (12 lines)
	return nil, nil
}

// ... 98+ more methods (1150 lines) ...
