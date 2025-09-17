package machine_translate

type Translator interface {
	TextTranslateBatch(textList []string, sourceLanguage, targetLanguage string) ([]*string, error)
}
type TranslatorFactory interface {
	CreateTranslator() (Translator, error)
}
