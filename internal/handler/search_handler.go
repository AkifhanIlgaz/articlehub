package handler

import (
	"context"
	"time"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
	"github.com/gofiber/fiber/v3"
)

// SearchStore; arama handler'ının bağımlı olduğu repository arayüzü.
type SearchStore interface {
	FullTextSearch(ctx context.Context, query string) ([]model.Article, error)
	SearchByCategoryAndDateRange(ctx context.Context, category string, from, to time.Time) ([]model.Article, error)
	FuzzySearch(ctx context.Context, q string) ([]model.Article, error)
	ListPaginated(ctx context.Context, from, size int, sortField, sortOrder string) ([]model.Article, int, error)
	SearchAfter(ctx context.Context, size int, sortField, sortOrder string, searchAfter []any) ([]model.Article, []any, error)
	SearchWithHighlight(ctx context.Context, q string) ([]model.HighlightedArticle, error)
	SearchWithFacets(ctx context.Context, q string) (*model.FacetResult, error)
	SuggestTitle(ctx context.Context, prefix string, size int) ([]string, error)
	SuggestCompletion(ctx context.Context, prefix string, size int) ([]string, error)
}

type SearchHandler struct {
	store SearchStore
}

func NewSearchHandler(store SearchStore) *SearchHandler {
	return &SearchHandler{store: store}
}

// Search; GET /search?q=&category=&from=&size=&sort=&order=&after=
// q + category filtreleri; from/size veya search_after sayfalama; sort/order sıralama.
// Yanıt: dto.SearchResponse
func (h *SearchHandler) Search(c fiber.Ctx) error {
	panic("not implemented")
}

// SearchFacets; GET /search/facets?q=
// Hits + kategori/tag/ay/istatistik aggregation'larını tek yanıtta döner.
// Yanıt: dto.FacetResponse
func (h *SearchHandler) SearchFacets(c fiber.Ctx) error {
	panic("not implemented")
}

// Suggest; GET /suggest?q=
// Önce completion suggester dener; fallback olarak edge n-gram (SuggestTitle) kullanır.
// Yanıt: dto.SuggestResponse
func (h *SearchHandler) Suggest(c fiber.Ctx) error {
	panic("not implemented")
}

// Health; GET /healthz — ES cluster health'ini döner, yellow/red ise 503.
func (h *SearchHandler) Health(c fiber.Ctx) error {
	panic("not implemented")
}

// AdminSeed; POST /admin/seed?count=500 — geliştirme ortamı için fake data üretir.
func (h *SearchHandler) AdminSeed(c fiber.Ctx) error {
	panic("not implemented")
}
