package entity

type BookingDetail struct {
	ID           string  `gorm:"column:id;type:uuid;primaryKey"`
	BookingID    string  `gorm:"column:booking_id;type:uuid;not null"`
	ProductID    string  `gorm:"column:product_id;type:uuid;not null"`
	ProductName  *string `gorm:"column:product_name;type:varchar(100)"`
	Qty          int32   `gorm:"column:qty;type:int;not null;default:1"`
	PricePerUnit float64 `gorm:"column:price_per_unit;type:decimal(15,2);not null"`
	SubTotal     float64 `gorm:"column:sub_total;type:decimal(15,2);not null"`
	CreatedAt    int64   `gorm:"column:created_at;type:bigint;not null;autoCreateTime:milli"`
	UpdatedAt    *int64  `gorm:"column:updated_at;type:bigint;autoUpdateTime:false"`
}

func (BookingDetail) TableName() string {
	return "booking_details"
}

// [ENTITY STANDARD: DOMAIN VALIDATION]
func (e *BookingDetail) Validate() error {
	return nil
}
