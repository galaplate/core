package supports

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type (
	XValidator struct{}

	GlobalErrorHandlerResp struct {
		Status  int               `json:"status"`
		Message string            `json:"message"`
		Errors  map[string]string `json:"errors"`
	}
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	err := validate.RegisterValidation("confirmation", fieldConfirmation)
	if err != nil {
		log.Panic(err)
	}
}

func fieldConfirmation(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().String()
	parent := fl.Top().Elem()

	param := fl.Param()
	confirmationField := parent.FieldByName(param).String()

	return fieldValue == confirmationField
}

func getJSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	name := strings.Split(tag, ",")[0]
	if name == "-" {
		return ""
	}

	return name
}

func getFieldJSONName(structType reflect.Type, fieldName string) string {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Name == fieldName {
			return getJSONFieldName(field)
		}
	}

	return ""
}

func (v XValidator) Validate(data interface{}) error {
	errorMessages := map[string]string{}
	var errorMessage string

	errs := validate.Struct(data)
	if errs != nil {
		for index, err := range errs.(validator.ValidationErrors) {
			jsonFieldName := getFieldJSONName(reflect.TypeOf(data), err.Field())
			errorMessages[jsonFieldName] = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", jsonFieldName, err.Tag())
			if index == 0 {
				errorMessage = errorMessages[jsonFieldName]
			}
		}
	}

	if len(errorMessages) > 0 {
		errorResponse := GlobalErrorHandlerResp{
			Status:  422,
			Errors:  errorMessages,
			Message: errorMessage,
		}
		errorJSON, err := json.Marshal(errorResponse)
		if err != nil {
			return fmt.Errorf("could not marshal validation errors: %v", err)
		}
		return fmt.Errorf("%s", string(errorJSON))
	}

	return nil
}

func (v XValidator) WithMessage(data GlobalErrorHandlerResp) error {
	errorJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("could not marshal validation errors: %v", err)
	}

	return &fiber.Error{
		Code:    fiber.StatusUnprocessableEntity,
		Message: string(errorJSON),
	}
}
