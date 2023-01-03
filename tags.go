package trans

// TransCtx describes a translation context specifying the language and domain to use when translating
type TransCtx struct {
	Language string
	Domain   string
}

// LanguageInfo describes a language.  It's used by the 'language' and 'get_available_languages'-tags
type LanguageInfo struct {
	Code string
	Name string
}

// Translator describes the interface that a translator implementation must fulfill
type Translator interface {
	Get(ctx TransCtx, str string, values ...interface{}) string
	GetC(ctx TransCtx, str string, transCtx string, values ...interface{}) string
	GetN(ctx TransCtx, str string, plural string, count int, values ...interface{}) string
	GetNC(ctx TransCtx, str string, plural string, count int, transCtx string, values ...interface{}) string
}
