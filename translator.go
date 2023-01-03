package trans

import (
	"io/fs"
	"path"
	"strings"

	"github.com/leonelquinteros/gotext"
)

// TemplateTranslator wraps gotext in an interface compatible with tagtrans
type TemplateTranslator struct {
	locales map[string]*gotext.Locale
}

// Get translates a string using gotext
func (t *TemplateTranslator) Get(ctx TransCtx, str string, values ...interface{}) string {
	// Always return empty string as translation for empty string,
	// otherwise the gettext header will be returned instead
	if str == "" {
		return ""
	}

	l, ok := t.locales[ctx.Language]
	if !ok {
		return str
	}

	dom := ctx.Domain
	if dom == "" {
		dom = "default"
	}
	return l.GetD(dom, str, values...)
}

// GetC translates a string using gotext, with a specific translation context
func (t *TemplateTranslator) GetC(ctx TransCtx, str string, transctx string, values ...interface{}) string {
	// Always return empty string as translation for empty string,
	// otherwise the gettext header will be returned instead
	if str == "" {
		return ""
	}

	l, ok := t.locales[ctx.Language]
	if !ok {
		return str
	}

	dom := ctx.Domain
	if dom == "" {
		dom = "default"
	}
	return l.GetDC(dom, str, transctx, values...)
}

// GetN translates a string using gotext, with support for plurals
func (t *TemplateTranslator) GetN(ctx TransCtx, str string, plural string, count int, values ...interface{}) string {
	// Always return empty string as translation for empty string,
	// otherwise the gettext header will be returned instead
	if str == "" {
		return ""
	}

	l, ok := t.locales[ctx.Language]
	if !ok {
		return str
	}

	dom := ctx.Domain
	if dom == "" {
		dom = "default"
	}
	return l.GetND(dom, str, plural, count, values...)
}

// GetNC translates a string using gotext, with a specific translation context, with support for plurals
func (t *TemplateTranslator) GetNC(ctx TransCtx, str string, plural string, count int, transctx string, values ...interface{}) string {
	// Always return empty string as translation for empty string,
	// otherwise the gettext header will be returned instead
	if str == "" {
		return ""
	}

	l, ok := t.locales[ctx.Language]
	if !ok {
		return str
	}

	dom := ctx.Domain
	if dom == "" {
		dom = "default"
	}
	return l.GetNDC(dom, str, plural, count, transctx, values...)
}

// NewTemplateTranslator creates a new translator, initialized with available locales
func NewTemplateTranslator(localeFS fs.FS, localePath string) (*TemplateTranslator, error) {
	localeDirs, err := fs.ReadDir(localeFS, localePath)
	if err != nil {
		return nil, err
	}

	locales := map[string]*gotext.Locale{}
	for _, dirEntry := range localeDirs {
		localeName := dirEntry.Name()

		lp := path.Join(localePath, localeName)
		domainFiles, err := fs.ReadDir(localeFS, lp)
		if err != nil {
			return nil, err
		}

		l := gotext.NewLocale("", localeName)
		if err != nil {
			return nil, err
		}

		domainMap := map[string]bool{}
		for _, domainFile := range domainFiles {
			if domainFile.IsDir() {
				continue
			}

			n := domainFile.Name()
			isPo := strings.HasSuffix(n, ".po")
			isMo := strings.HasSuffix(n, ".mo")

			if strings.HasPrefix(n, ".") ||
				(!isPo && !(isMo)) {
				continue
			}

			// trim suffix
			domainName := n[:len(n)-len(".po")]

			if domainMap[domainName] {
				// We've already loaded either the .PO or .MO file
				continue
			}

			contents, err := fs.ReadFile(localeFS, path.Join(lp, n))
			if err != nil {
				return nil, err
			}

			var tr gotext.Translator
			if isPo {
				po := gotext.NewPo()
				po.Parse(contents)
				tr = po
			} else {
				mo := gotext.NewMo()
				mo.Parse(contents)
				tr = mo
			}
			l.AddTranslator(domainName, tr)
			domainMap[domainName] = true
		}

		//// If we don't have a default domain, we'll set the first one as default
		//if !domainMap["default"] {
		//	dom := l.GetDomain()
		//	if dom != "" && l.Domains[dom] != nil {
		//		l.AddTranslator("default", l.Domains[dom])
		//	}
		//}

		locales[localeName] = l
	}

	// If our directory matches a regional locale, we'll have to
	// map it to a general locale as well (if we don't have one)
	additionalLocales := map[string]*gotext.Locale{}
	for name, locale := range locales {
		parts := strings.Split(name, "_")
		if len(parts) > 1 {
			// Do not create a new mapping if we already have one
			if _, ok := locales[parts[0]]; ok {
				continue
			}
			additionalLocales[parts[0]] = locale
		}
	}

	// Add the new locales to the default mapping
	for name, locale := range additionalLocales {
		locales[name] = locale
	}

	t := &TemplateTranslator{locales: locales}
	return t, nil
}
