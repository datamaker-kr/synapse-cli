package i18n

import (
	"embed"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed messages/*.yaml
var messageFS embed.FS

var (
	localizer   *i18n.Localizer
	currentLang string
	mu          sync.RWMutex
)

// Init initializes the i18n system with the given language tag.
func Init(lang string) {
	mu.Lock()
	defer mu.Unlock()

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Load embedded message files (errors are ignored — fallback to message ID)
	_, _ = bundle.LoadMessageFileFS(messageFS, "messages/en.yaml")
	_, _ = bundle.LoadMessageFileFS(messageFS, "messages/ko.yaml")

	localizer = i18n.NewLocalizer(bundle, lang, "en")
	currentLang = lang
}

// T translates a message by ID with optional template data.
func T(messageID string, data ...map[string]interface{}) string {
	mu.RLock()
	l := localizer
	mu.RUnlock()

	if l == nil {
		return messageID
	}

	cfg := &i18n.LocalizeConfig{MessageID: messageID}
	if len(data) > 0 && data[0] != nil {
		cfg.TemplateData = data[0]
	}

	msg, err := l.Localize(cfg)
	if err != nil {
		return messageID
	}
	return msg
}

// CurrentLang returns the currently active language tag.
func CurrentLang() string {
	mu.RLock()
	defer mu.RUnlock()
	if currentLang == "" {
		return "en"
	}
	return currentLang
}
