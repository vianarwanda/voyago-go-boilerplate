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
	routeGroup = "/bookings"
)

func (r *RouteConfig) Setup() {
	bookings := r.Server.Group(routeGroup)
	bookings.Post("/", r.Handler.CreateBooking)
}
