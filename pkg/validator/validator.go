package validator

import (
	"context"
	"errors"
	"github.com/go-playground/validator"
	"log"
	"regexp"
)

var validate *validator.Validate

const (
	ErrInvalidFormat      = "Invalid format"
	ErrFieldRequired      = "Field is required"
	ErrFieldExceedsMaxLen = "Field exceeds maximum length"
	ErrFieldBelowMinLen   = "Field is below minimum length"
	ErrFieldExceedsMaxVal = "Field exceeds maximum value"
	ErrFieldBelowMinVal   = "Field is below minimum value"
	ErrNotEmail           = "Field is not a valid email"
	ErrUnknownValidation  = "Unknown validation error"
)

func init() {
	SetValidator(NewValidator())
}

func NewValidator() *validator.Validate {
	v := validator.New()
	err := v.RegisterValidation("tag", validateTag)
	if err != nil {
		log.Fatal("unable to register validator: ", err)
	}
	return v
}

func Validator() *validator.Validate {
	return validate
}

func SetValidator(v *validator.Validate) {
	validate = v
}

func validateTag(fl validator.FieldLevel) bool {
	re, _ := regexp.Compile(`^#[a-z0-9_\-]+$`)
	return re.MatchString(fl.Field().String())
}

func ValidateStruct(ctx context.Context, data interface{}) error {
	return parseValidationErrors(Validator().StructCtx(ctx, data))
}

func parseValidationErrors(err error) error {
	if err == nil {
		return nil
	}

	vErr, ok := err.(validator.ValidationErrors)
	if !ok || len(vErr) == 0 {
		return nil
	}

	validError := vErr[0]
	var validErrDesc string
	switch validError.Tag() {
	case "tag":
		validErrDesc = ErrInvalidFormat
	case "required":
		validErrDesc = ErrFieldRequired
	case "email":
		validErrDesc = ErrNotEmail
	case "max":
		validErrDesc = ErrFieldExceedsMaxLen
	case "min":
		validErrDesc = ErrFieldBelowMinLen
	case "lt", "lte":
		validErrDesc = ErrFieldExceedsMaxVal
	case "gt", "gte":
		validErrDesc = ErrFieldBelowMinVal
	default:
		validErrDesc = ErrUnknownValidation
	}
	return errors.New(validErrDesc + ": " + validError.Namespace())
}
