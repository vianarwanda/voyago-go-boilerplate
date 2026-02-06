package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type playgroundValidator struct {
	driver *validator.Validate
}

var _ Validator = (*playgroundValidator)(nil)

func NewPlaygroundValidator() Validator {
	driver := validator.New()
	driver.RegisterTagNameFunc(func(fld reflect.StructField) string {
		jsonName := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if jsonName == "-" || jsonName == "" {
			jsonName = fld.Name
		}

		labelName := fld.Tag.Get("label")
		if labelName == "" {
			labelName = jsonName
		}
		return fmt.Sprintf("%s|%s", jsonName, labelName)
	})
	// driver.RegisterTagNameFunc(func(fld reflect.StructField) string {
	// 	name := fld.Tag.Get("label")
	// 	if name != "" {
	// 		return name
	// 	}

	// 	jsonName := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	// 	if jsonName != "-" && jsonName != "" {
	// 		return jsonName
	// 	}

	// 	return fld.Name
	// })
	return &playgroundValidator{
		driver: driver,
	}
}

func (v *playgroundValidator) Validate(i any) error {
	return v.driver.Struct(i)
}

func (v *playgroundValidator) ToCustomError(err error) []ValidationError {
	var result []ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			message := v.translateTag(fe)
			code := fe.Tag()
			if code == "uuid_rfc4122" {
				code = "uuid"
			}
			result = append(result, ValidationError{
				Field:   v.getJsonLabel(fe),
				Message: message,
				Code:    code,
			})
		}
	}
	return result
}

func (v *playgroundValidator) ToMap(err error) map[string]any {
	res := make(map[string]any)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			message := v.translateTag(fe)
			code := fe.Tag()
			if code == "uuid_rfc4122" {
				code = "uuid"
			}
			res[v.getJsonLabel(fe)] = map[string]any{
				"message": message,
				"code":    code,
				"param":   fe.Param(),
			}
		}
	}
	return res
}

func (v *playgroundValidator) ToDetails(err error) []map[string]any {
	var res []map[string]any

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return res
	}

	for _, fe := range ve {
		message := v.translateTag(fe)
		code := fe.Tag()
		if code == "uuid_rfc4122" {
			code = "uuid"
		}
		entry := map[string]any{
			"field":   v.getJsonLabel(fe),
			"message": message,
			"code":    code,
			"param":   fe.Param(),
		}
		res = append(res, entry)
	}

	return res
}

func (v *playgroundValidator) translateTag(fe validator.FieldError) string {
	displayLabel := v.getLabel(fe)
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", displayLabel)

	case "min":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", displayLabel, param)
		}
		return fmt.Sprintf("%s must be at least %s", displayLabel, param)

	case "max":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must not be greater than %s characters", displayLabel, param)
		}
		return fmt.Sprintf("%s must not be greater than %s", displayLabel, param)

	case "email":
		return fmt.Sprintf("%s is an invalid email address", displayLabel)

	case "uuid", "uuid_rfc4122":
		return fmt.Sprintf("%s must be a valid UUID", displayLabel)

	case "gt":
		return fmt.Sprintf("%s must be greater than %s", displayLabel, param)

	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", displayLabel, param)

	case "lt":
		return fmt.Sprintf("%s must be less than %s", displayLabel, param)

	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", displayLabel, param)

	case "eq":
		return fmt.Sprintf("%s must be equal to %s", displayLabel, param)

	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", displayLabel, param)

	default:
		return fmt.Sprintf("%s is invalid", displayLabel)
	}
}

func (v *playgroundValidator) getLabel(fe validator.FieldError) string {
	parts := strings.Split(fe.Field(), "|")
	displayLabel := parts[0]

	if len(parts) > 1 {
		displayLabel = parts[1]
	}
	return displayLabel
}

func (v *playgroundValidator) getJsonLabel(fe validator.FieldError) string {
	parts := strings.Split(fe.Field(), "|")
	displayJson := parts[0]

	if len(parts) > 1 {
		displayJson = parts[0]
	}
	return displayJson
}
