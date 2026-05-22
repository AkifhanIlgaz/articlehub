package model

// HighlightedArticle; arama sonucunda eşleşen alanları HTML tag'leriyle işaretlenmiş halde taşır.
type HighlightedArticle struct {
	Article
	Highlight map[string][]string `json:"highlight"`
}

// FacetResult; SearchWithFacets'in döndürdüğü tam yanıt — hem hit'ler hem aggregation'lar.
type FacetResult struct {
	Total      int             `json:"total"`
	Articles   []Article       `json:"articles"`
	Categories []CategoryFacet `json:"categories"`
	Tags       []TagFacet      `json:"tags"`
	ByMonth    []MonthBucket   `json:"by_month"`
	ViewStats  ViewCountStats  `json:"view_stats"`
}

// CategoryFacet; tek bir kategorinin makale sayısını tutar (terms aggregation bucket).
type CategoryFacet struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// TagFacet; tek bir tag'in makale sayısını tutar.
type TagFacet struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// MonthBucket; date_histogram aggregation'ının aylık bucket'ı.
type MonthBucket struct {
	Month string `json:"month"` // "2024-01"
	Count int    `json:"count"`
}

// ViewCountStats; view_count üzerindeki stats aggregation sonucu.
type ViewCountStats struct {
	Min int     `json:"min"`
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
	Sum int     `json:"sum"`
}
