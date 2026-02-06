package usecase

import (
	"context"
)

// -------- DTOs --------
type CreateBookingRequest struct {
	// BookingID   string                       `json:"booking_id" validate:"required,uuid" label:"Booking ID"`
	BookingCode string                       `json:"code" validate:"required,min=3,max=50" label:"Booking code"`
	UserID      string                       `json:"user_id" validate:"required,uuid" label:"User ID"`
	TotalAmount float64                      `json:"total_amount" validate:"gte=0" label:"Total amount"`
	Details     []CreateBookingDetailRequest `json:"details" validate:"required,min=1,dive" label:"Details"`
}

type CreateBookingDetailRequest struct {
	ProductID    string  `json:"product_id" validate:"required,uuid_rfc4122" label:"Product ID"`
	ProductName  *string `json:"product_name" validate:"omitempty,max=100" label:"Product name"`
	Qty          int32   `json:"qty" validate:"required,gt=0" label:"Quantity"`
	PricePerUnit float64 `json:"price_per_unit" validate:"required,gt=0" label:"Price per unit"`
	SubTotal     float64 `json:"sub_total" validate:"required,gt=0" label:"Sub total"`
}

type CreateBookingResponse struct {
	BookingID   string                        `json:"id"`
	BookingCode string                        `json:"code"`
	UserID      string                        `json:"user_id"`
	TotalAmount float64                       `json:"total_amount"`
	Details     []CreateBookingDetailResponse `json:"details"`
}

type CreateBookingDetailResponse struct {
	ProductID    string  `json:"product_id"`
	ProductName  *string `json:"product_name"`
	Qty          int32   `json:"qty"`
	PricePerUnit float64 `json:"price_per_unit"`
	SubTotal     float64 `json:"sub_total"`
}

// -------- Usecase Interfaces --------
// [CONTRACT DEFINITION]
// CreateBookingUseCase defines the business contract for booking creation.
// High-level orchestration is hidden behind this interface.
type CreateBookingUseCase interface {
	// Execute processes the booking request.
	// It returns a CreateBookingResponse on success or an apperror.AppError on failure.
	Execute(ctx context.Context, req *CreateBookingRequest) (*CreateBookingResponse, error)
}
