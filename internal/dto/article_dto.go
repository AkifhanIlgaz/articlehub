package dto

// CreateArticleRequest; POST /articles body'si.
type CreateArticleRequest struct {
	Title       string   `json:"title"        validate:"required,min=3,max=200"`
	Slug        string   `json:"slug"         validate:"required"`
	Content     string   `json:"content"      validate:"required"`
	Author      string   `json:"author"       validate:"required"`
	Tags        []string `json:"tags"`
	ReadingTime int      `json:"reading_time"`
}

// UpdateArticleRequest; PUT /articles/:id body'si (tam güncelleme).
type UpdateArticleRequest struct {
	Title       string   `json:"title"        validate:"required,min=3,max=200"`
	Slug        string   `json:"slug"         validate:"required"`
	Content     string   `json:"content"      validate:"required"`
	Author      string   `json:"author"       validate:"required"`
	Tags        []string `json:"tags"`
	ReadingTime int      `json:"reading_time"`
}

// PartialUpdateArticleRequest; PATCH /articles/:id body'si.
// Pointer alanlar sayesinde sıfır değer ile "gönderilmedi" ayrımı yapılabilir.
type PartialUpdateArticleRequest struct {
	Title       *string  `json:"title,omitempty"`
	Content     *string  `json:"content,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ReadingTime *int     `json:"reading_time,omitempty"`
}

// SearchRequest; GET /search query parametreleri.
type SearchRequest struct {
	Q        string `query:"q"`
	Category string `query:"category"`
	From     int    `query:"from"`
	Size     int    `query:"size"`
	Sort     string `query:"sort"`  // "published_at" | "view_count" | "relevance"
	Order    string `query:"order"` // "asc" | "desc"
	After    string `query:"after"` // search_after cursor (JSON-encoded)
}

// ArticleResponse; dışarıya açılan makale temsili — ES iç alanları (_id, _score vb.) gizlenir.
type ArticleResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Content     string   `json:"content"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	ReadingTime int      `json:"reading_time"`
	ViewCount   int      `json:"view_count"`
	LikeCount   int      `json:"like_count"`
	PublishedAt string   `json:"published_at"`
}

// SearchResponse; GET /search yanıtı.
type SearchResponse struct {
	Total    int               `json:"total"`
	Articles []ArticleResponse `json:"articles"`
	After    []any             `json:"after,omitempty"` // search_after cursor
}

// HighlightResponse; highlight içeren arama yanıtı.
type HighlightResponse struct {
	Article   ArticleResponse     `json:"article"`
	Highlight map[string][]string `json:"highlight"`
}

// FacetResponse; GET /search/facets yanıtı.
type FacetResponse struct {
	Total      int               `json:"total"`
	Articles   []ArticleResponse `json:"articles"`
	Categories []CategoryFacet   `json:"categories"`
	Tags       []TagFacet        `json:"tags"`
	ByMonth    []MonthBucket     `json:"by_month"`
	ViewStats  ViewCountStats    `json:"view_stats"`
}

type CategoryFacet struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type TagFacet struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type MonthBucket struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type ViewCountStats struct {
	Min int     `json:"min"`
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
	Sum int     `json:"sum"`
}

// SuggestResponse; GET /suggest yanıtı.
type SuggestResponse struct {
	Suggestions []string `json:"suggestions"`
}
