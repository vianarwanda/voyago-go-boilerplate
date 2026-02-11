package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
	"voyago/core-api/internal/pkg/apperror"
)

const (
	CategoryNotFound   = "CATEGORY_NOT_FOUND"
	CategoryIdRequired = "CATEGORY_ID_REQUIRED"
)

var (
	ErrCategoryNotFound = apperror.NewPersistance(
		CategoryNotFound,
		"category not found",
	)

	ErrCategoryIdRequired = apperror.NewPersistance(
		CategoryIdRequired,
		"id is required",
	)

	ErrUnsuppoertedLanguage = apperror.NewPersistance(
		"CATEGORY_UNSUPPORTED_LANGUAGE",
		"unsupported language",
	)

	ErrLocalizedCannotBeEmpty = apperror.NewPersistance(
		"CATEGORY_LOCALIZED_CANNOT_BE_EMPTY",
		"localized cannot be empty",
	)

	ErrLocalizedRequired = apperror.NewPersistance(
		"CATEGORY_LOCALIZED_CANNOT_BE_EMPTY",
		"localized value required",
	)
)

var allowedLang = map[string]struct{}{
	"en-US": {},
	"id-ID": {},
}

type Localized map[string]string

func (l *Localized) Scan(src any) error {
	if src == nil {
		*l = nil
		return nil
	}

	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for Localized")
	}

	var tmp map[string]string
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}

	*l = make(Localized)
	for k, v := range tmp {
		(*l)[k] = v
	}

	return nil
}

func (l Localized) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l Localized) Validate() error {
	if len(l) == 0 {
		return ErrLocalizedCannotBeEmpty
	}

	for k, v := range l {
		if _, ok := allowedLang[k]; !ok {
			return ErrUnsuppoertedLanguage
		}
		if v == "" {
			return ErrLocalizedRequired
		}
	}
	return nil
}

// Category entity
type Category struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey"`
	ParentID    *string    `gorm:"column:parent_id;type:uuid"`
	Name        Localized  `gorm:"column:name;type:jsonb;not null"`
	Slug        Localized  `gorm:"column:slug;type:jsonb;not null"`
	Description *Localized `gorm:"column:description;type:jsonb"`
	IconURL     *string    `gorm:"column:icon_url;type:text"`
	SortOrder   int        `gorm:"column:sort_order;type:int;default:0"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	NameDefault string     `gorm:"column:name_default;->"`
	SlugDefault string     `gorm:"column:slug_default;->"`
}

func (Category) TableName() string { return "categories" }

func (e *Category) Validate() error {
	if err := e.Name.Validate(); err != nil {
		return err
	}

	if err := e.Slug.Validate(); err != nil {
		return err
	}

	if e.Description != nil {
		if err := e.Description.Validate(); err != nil {
			return err
		}
	}

	return nil
}
