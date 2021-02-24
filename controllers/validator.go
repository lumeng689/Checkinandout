package controllers

import (
	"log"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// CCValidator - global setup for input form validators
type CCValidator struct {
	v     *validator.Validate
	trans *ut.Translator
}

var regexpPhoneNum = regexp.MustCompile(`\d{3}-\d{3}-\d{4}`)
var regexpState = regexp.MustCompile(`[A-Z]{2}`)
var regexpZipCode = regexp.MustCompile(`\d{5}`)

// InitValidator - as is
func (s *CCServer) InitValidator() {

	translator := en.New()
	uni := ut.New(translator, translator)

	// this is usually known or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	trans, found := uni.GetTranslator("en")
	if !found {
		log.Fatal("translator not found")
	}

	v := validator.New()

	// Rule for PhoneNum
	_ = v.RegisterValidation("phone_num", func(fl validator.FieldLevel) bool {
		return regexpPhoneNum.Match([]byte(fl.Field().String()))
	})

	// Rule for state abbrev.
	_ = v.RegisterValidation("state", func(fl validator.FieldLevel) bool {
		return regexpState.Match([]byte(fl.Field().String()))
	})

	// Rule for ZipCode
	_ = v.RegisterValidation("zip_code", func(fl validator.FieldLevel) bool {
		return regexpZipCode.Match([]byte(fl.Field().String()))
	})

	// Report for required
	_ = v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is a required field", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	// Report for minimum length
	_ = v.RegisterTranslation("min", trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} is too short", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("min", fe.Field())
		return t
	})

	// Report for max length
	_ = v.RegisterTranslation("max", trans, func(ut ut.Translator) error {
		return ut.Add("max", "{0} is too long", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("max", fe.Field())
		return t
	})

	// Report for PhoneNum
	_ = v.RegisterTranslation("phone_num", trans, func(ut ut.Translator) error {
		return ut.Add("phone_num", "Phone Number NEED to be in format: 123-456-7890", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("phone_num", fe.Field())
		return t
	})

	// Report for state abbrev.
	_ = v.RegisterTranslation("state", trans, func(ut ut.Translator) error {
		return ut.Add("state", "State must be 2 CAPITAL LETTER abbrev., e.g.: AZ", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("state", fe.Field())
		return t
	})

	// Report for ZipCode
	_ = v.RegisterTranslation("zip_code", trans, func(ut ut.Translator) error {
		return ut.Add("zip_code", "ZipCode must be 5 digits, e.g.: 90001", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("zip_code", fe.Field())
		return t
	})

	// Report for TagString
	_ = v.RegisterTranslation("tag_string", trans, func(ut ut.Translator) error {
		return ut.Add("tag_string", "TagString not comply with rule!", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("tag_string", fe.Field())
		return t
	})

	s.Validator.v = v
	s.Validator.trans = &trans
}

// RegisterTagStringValidator - dynamically load tag_string rule for each institution
func (s *CCServer) RegisterTagStringValidator(tagStringRule string) {
	// Convert to Regex String
	tagStringRegex := strings.ReplaceAll(tagStringRule, `\\`, `\`)
	log.Println("TagStringRegex - ", tagStringRegex)

	// Rule for TagString
	v := s.Validator.v
	var regexpTagString = regexp.MustCompile(tagStringRegex)
	_ = v.RegisterValidation("tag_string", func(fl validator.FieldLevel) bool {
		return regexpTagString.Match([]byte(fl.Field().String()))
	})

}
