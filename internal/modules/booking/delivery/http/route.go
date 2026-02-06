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

	bookings.Get("/:id", func(c *fiber.Ctx) error {
		// 1. Ambil parameter ID dari URL
		id := c.Params("id")

		// 2. Simulasi data (biasanya lo ambil dari DB di sini)
		// Kita buat response body-nya
		response := fiber.Map{
			"success": true,
			"message": "Booking details retrieved successfully",
			"data": fiber.Map{
				"booking_id": id,
				"status":     "confirmed",
				"customer":   "John Doe",
			},
		}

		// 3. Return JSON 200
		return c.Status(fiber.StatusOK).JSON(response)
	})
}
