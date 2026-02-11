package http

import (
	"voyago/core-api/internal/infrastructure/config"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	Config  *config.Config
	Server  *fiber.App
	Handler *Handler
}

const (
	routeGroup = "/products"
)

func (r *RouteConfig) Setup() {
	products := r.Server.Group(routeGroup)
	//products.Get("/");
	//products.Post("/");
	products.Get("/categories", r.Handler.ReadCategory)
	products.Post("/categories", r.Handler.CreateCategory)
	products.Get("/categories/:id", r.Handler.ReadCategoryDetail)
}
