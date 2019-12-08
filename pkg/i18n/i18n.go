package i18n

import (
	"path"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Translator translates messages.
type Translator struct {
	// contains filtered or unexported fields
	localizer *i18n.Localizer
}

// NewTranslatorFromTOMLFile instanciates a newly initialized Translator
// from the corresponding .toml translation file located in translationsPath.
func NewTranslatorFromTOMLFile(lang, translationsPath string) *Translator {
	bndl := i18n.NewBundle(language.English)
	bndl.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bndl.MustLoadMessageFile(path.Join(translationsPath, lang+".toml"))

	return &Translator{localizer: i18n.NewLocalizer(bndl, lang)}
}

// Translate translates the formatted message with template.
func (t *Translator) Translate(message string, template map[string]interface{}) string {
	return t.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    message,
		TemplateData: template,
	})
}
