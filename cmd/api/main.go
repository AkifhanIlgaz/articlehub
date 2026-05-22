package main

import (
	"log"

	"github.com/AkifhanIlgaz/articlehub/internal/config"
	"github.com/AkifhanIlgaz/articlehub/internal/handler"
	"github.com/AkifhanIlgaz/articlehub/internal/repository"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	es, err := elasticsearch.NewTyped(elasticsearch.WithAddresses(cfg.ESAddress))
	if err != nil {
		log.Fatalf("elasticsearch: %v", err)
	}

	repo, err := repository.New(es)
	if err != nil {
		log.Fatalf("repository: %v", err)
	}

	articleH := handler.NewArticleHandler(repo)
	searchH := handler.NewSearchHandler(repo)

	app := fiber.New(fiber.Config{
		// TODO: global error handler ekle — ES hatalarını kullanıcıya direkt sızdırma
	})

	registerRoutes(app, articleH, searchH)

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func registerRoutes(app *fiber.App, articleH *handler.ArticleHandler, searchH *handler.SearchHandler) {
	app.Post("/articles", articleH.Create)
	app.Get("/articles/:id", articleH.GetByID)
	app.Put("/articles/:id", articleH.Update)
	app.Patch("/articles/:id", articleH.PartialUpdate)
	app.Delete("/articles/:id", articleH.Delete)

	app.Get("/search", searchH.Search)
	app.Get("/search/facets", searchH.SearchFacets)
	app.Get("/suggest", searchH.Suggest)

	app.Post("/admin/seed", searchH.AdminSeed)
	app.Get("/healthz", searchH.Health)
}
