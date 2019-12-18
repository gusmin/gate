package i18n

import (
	"path"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Translator translates messages in different languages
// according to translations files.
type Translator struct {
	// contains filtered or unexported fields
	localizer *i18n.Localizer
}

// NewTranslatorFromFile returns a new Translator translating message
// to lang based on .toml translation files located in the translationsPath.
func NewTranslatorFromFile(lang, translationsPath string) *Translator {
	bndl := i18n.NewBundle(language.English)
	bndl.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bndl.MustLoadMessageFile(path.Join(translationsPath, lang+".toml"))

	return &Translator{localizer: i18n.NewLocalizer(bndl, lang)}
}

// Translate translates a templated message.
func (t Translator) Translate(message string) string {
	return t.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message,
	})
}
