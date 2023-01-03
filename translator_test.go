package trans

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/locales
var localeTestdata embed.FS

func TestNewTemplateTranslator(t *testing.T) {
	tt, err := NewTemplateTranslator(localeTestdata, "testdata/locales")
	require.Nil(t, err)

	// Check that both of our locales was loaded
	require.NotNil(t, tt.locales["en_GB"])
	require.NotNil(t, tt.locales["sv_SE"])

	// Check that our domains was loaded
	require.Equal(t, "Hej världen!", tt.Get(TransCtx{Language: "sv_SE"}, "Hello world!"))
	require.Equal(t, "Hej från other!", tt.Get(TransCtx{Language: "sv_SE", Domain: "other"}, "Hello world!"))

	// en_GB only contains the 'other'-domain, so translations without domain should not be changed
	require.Equal(t, "Hello world!", tt.Get(TransCtx{Language: "en_GB"}, "Hello world!"))
	// But the other domain should still work as expected
	require.Equal(t, "Hello from the other domain!", tt.Get(TransCtx{Language: "en_GB", Domain: "other"}, "Hello world!"))
}
