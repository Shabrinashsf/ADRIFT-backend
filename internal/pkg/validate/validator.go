package validate

import (
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"

	"ADRIFT-backend/internal/dto"
	myerror "ADRIFT-backend/internal/pkg/error"
	"ADRIFT-backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

func (v *Validator) ParseAndValidate(ctx *gin.Context, req any) error {
	if err := ctx.ShouldBind(req); err != nil {
		return err
	}

	if err := v.bindFileFields(ctx, req); err != nil {
		return err
	}

	return v.validate.Struct(req)
}

func (v *Validator) bindFileFields(ctx *gin.Context, req any) error {
	rv := reflect.ValueOf(req).Elem()
	rt := rv.Type()

	for i := range rv.NumField() {
		field := rt.Field(i)
		if field.Type != reflect.TypeOf((*multipart.FileHeader)(nil)) {
			continue
		}

		file, err := ctx.FormFile(field.Tag.Get("form"))
		if err != nil {
			continue
		}

		rv.Field(i).Set(reflect.ValueOf(file))
	}

	return nil
}

func (v *Validator) Bind(ctx *gin.Context, req any) bool {
	if err := v.ParseAndValidate(ctx, req); err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, myerror.New(myerror.FormatValidationError(err), http.StatusBadRequest), nil).Send(ctx)
		return false
	}
	return true
}
