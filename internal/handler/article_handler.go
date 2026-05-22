package handler

import (
	"context"

	"github.com/AkifhanIlgaz/articlehub/internal/model"
	"github.com/gofiber/fiber/v3"
)

// ArticleStore; handler'ın bağımlı olduğu repository arayüzü.
// Concrete tipi yerine interface'e bağımlı olmak test edilebilirliği artırır.
type ArticleStore interface {
	Index(ctx context.Context, article model.Article) error
	GetArticle(ctx context.Context, id string) (*model.Article, error)
	ReplaceArticle(ctx context.Context, id string, article model.Article) error
	PatchArticle(ctx context.Context, id string, fields map[string]any) error
	DeleteArticle(ctx context.Context, id string) error
}

type ArticleHandler struct {
	store ArticleStore
}

func NewArticleHandler(store ArticleStore) *ArticleHandler {
	return &ArticleHandler{store: store}
}

// Create; POST /articles
// Body: dto.CreateArticleRequest → model.Article → ES index → 201 + dto.ArticleResponse
func (h *ArticleHandler) Create(c fiber.Ctx) error {
	panic("not implemented")
}

// GetByID; GET /articles/:id
// ES'den get, bulunamazsa 404.
func (h *ArticleHandler) GetByID(c fiber.Ctx) error {
	panic("not implemented")
}

// Update; PUT /articles/:id
// Body: dto.UpdateArticleRequest — tam güncelleme.
func (h *ArticleHandler) Update(c fiber.Ctx) error {
	panic("not implemented")
}

// PartialUpdate; PATCH /articles/:id
// Body: dto.PartialUpdateArticleRequest — sadece gönderilen alanlar güncellenir.
func (h *ArticleHandler) PartialUpdate(c fiber.Ctx) error {
	panic("not implemented")
}

// Delete; DELETE /articles/:id — 204 döner.
func (h *ArticleHandler) Delete(c fiber.Ctx) error {
	panic("not implemented")
}
