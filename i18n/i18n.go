package i18n

import (
	"embed"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

var bundle *i18n.Bundle

func Init() error {
	bundle = i18n.NewBundle(language.Japanese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load Japanese translations
	if _, err := bundle.LoadMessageFileFS(localeFS, "locales/ja.json"); err != nil {
		return err
	}

	// Load English translations
	if _, err := bundle.LoadMessageFileFS(localeFS, "locales/en.json"); err != nil {
		return err
	}

	return nil
}

// Middleware adds localizer to context
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get language from cookie, default to Japanese
			lang := "ja"
			if cookie, err := c.Cookie("lang"); err == nil {
				lang = cookie.Value
			}

			// Create localizer
			localizer := i18n.NewLocalizer(bundle, lang)
			c.Set("localizer", localizer)
			c.Set("lang", lang)

			return next(c)
		}
	}
}

// T translates a message
func T(c echo.Context, key string) string {
	localizer, ok := c.Get("localizer").(*i18n.Localizer)
	if !ok {
		return key
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	if err != nil {
		return key
	}

	return msg
}

// GetLocalizer returns the localizer from context
func GetLocalizer(c echo.Context) *i18n.Localizer {
	localizer, _ := c.Get("localizer").(*i18n.Localizer)
	return localizer
}

// GetCurrentLang returns the current language
func GetCurrentLang(c echo.Context) string {
	lang, _ := c.Get("lang").(string)
	if lang == "" {
		return "ja"
	}
	return lang
}
