package trans

import (
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTagTransNode_Execute(t *testing.T) {
	testTrans := MockTranslator{}
	testTrans.On("Get", mock.Anything, "test").Return("ok", nil)
	testTrans.On("Get", mock.Anything, "test with {{ var }}").Return("ok-var", nil)
	testTrans.On("Get", mock.Anything, "test with {{var2}}").Return("ok-var2", nil)
	testTrans.On("Get", mock.Anything, "test with {{ var3 }} {{ var4 }} post").Return("ok-var3", nil)
	testTrans.On("GetC", mock.Anything, "test", "myctx").Return("ok-ctx", nil)
	testTrans.On("GetN", mock.Anything, "test-1", "test-2", 1).Return("ok-1", nil)
	testTrans.On("GetN", mock.Anything, "test-1", "test-2", 2).Return("ok-2", nil)
	testTrans.On("GetNC", mock.Anything, "test-1", "test-2", 1, "myctx").Return("ok-1-ctx", nil)
	testTrans.On("GetNC", mock.Anything, "test-1", "test-2", 2, "myctx").Return("ok-2-ctx", nil)
	err := pongo2.RegisterTag("trans", NewTransTag(&testTrans))
	if err != nil {
		err = pongo2.ReplaceTag("trans", NewTransTag(&testTrans))
	}
	require.Nil(t, err)

	err = pongo2.RegisterTag("blocktrans", NewBlockTransTag(&testTrans))
	if err != nil {
		err = pongo2.ReplaceTag("blocktrans", NewBlockTransTag(&testTrans))
	}
	require.Nil(t, err)

	type T struct {
		input    string
		expected string
		err      bool
	}

	tests := []T{
		{input: `{% trans "test" %}`, expected: "ok"},
		{input: `{% trans "test" %} post`, expected: "ok post"},
		{input: `pre {% trans "test" %}`, expected: "pre ok"},
		{input: `{% trans "test" as myvar %}{{myvar}}`, expected: "ok"},
		{input: `{% trans "test" context "myctx" %}`, expected: "ok-ctx"},
		{input: `{% trans "test" as othervar context "myctx" %}{{othervar}}`, expected: "ok-ctx"},
		{input: `{% trans "test" as %}`, err: true},
		{input: `{% trans "test" context blah %}`, err: true},

		{input: `{% blocktrans %}test{% endblocktrans %}`, expected: "ok"},
		{input: `{% blocktrans %}test{% endblocktrans %} post`, expected: "ok post"},
		{input: `pre {% blocktrans %}test{% endblocktrans %}`, expected: "pre ok"},
		{input: `{% blocktrans %}test with {{ var }}{% endblocktrans %}`, expected: "ok-var"},
		{input: `{% blocktrans %}test with {{var2}}{% endblocktrans %}`, expected: "ok-var2"},
		{input: `{% blocktrans %}test with {{ var3 }} {{ var4 }} post{% endblocktrans %}`, expected: "ok-var3"},
		{input: `{% blocktrans count cnt=1 %}test-1{% plural %}test-2{% endblocktrans %}`, expected: "ok-1"},
		{input: `{% blocktrans count cnt=2 %}test-1{% plural %}test-2{% endblocktrans %}`, expected: "ok-2"},
		{input: `{% blocktrans count cnt=1 context "myctx" %}test-1{% plural %}test-2{% endblocktrans %}`, expected: "ok-1-ctx"},
		{input: `{% blocktrans count cnt=2 context "myctx" %}test-1{% plural %}test-2{% endblocktrans %}`, expected: "ok-2-ctx"},
		{input: `{% blocktrans asvar the_title %}test{% endblocktrans%}{{the_title}}`, expected: "ok"},
		{input: `{% blocktrans context "myctx" asvar the_title %}test{% endblocktrans%}{{the_title}}`, expected: "ok-ctx"},
		{input: `{% blocktrans count cnt=1 asvar the_title %}test-1{% plural %}test-2{% endblocktrans%}{{the_title}}`, expected: "ok-1"},
		{input: `{% blocktrans count cnt=1 context "myctx" asvar the_title %}test-1{% plural %}test-2{% endblocktrans%}{{the_title}}`, expected: "ok-1-ctx"},
		{input: `{% blocktrans count cnt=2 context "myctx" asvar the_title %}test-1{% plural %}test-2{% endblocktrans%}{{the_title}}`, expected: "ok-2-ctx"},
	}

	for k, tst := range tests {
		tmpl, err := pongo2.FromString(tst.input)
		if tst.err {
			require.NotNilf(t, err, "test: %d, input: %s", k, tst.input)
			continue
		}
		require.Nilf(t, err, "test: %d, input: %s", k, tst.input)
		result, err := tmpl.Execute(pongo2.Context{})
		require.Nilf(t, err, "test: %s input: %s", k, tst.input)
		require.Equalf(t, tst.expected, result, "test: %s input: %s", k, tst.input)
	}
}

// Check that context is accessible
func TestTagTransNode_ContextAccess(t *testing.T) {
	testTrans := TestTranslator{}
	err := pongo2.RegisterTag("trans", NewTransTag(&testTrans))
	if err != nil {
		err = pongo2.ReplaceTag("trans", NewTransTag(&testTrans))
	}
	require.Nil(t, err)

	err = pongo2.RegisterTag("blocktrans", NewBlockTransTag(&testTrans))
	if err != nil {
		err = pongo2.ReplaceTag("blocktrans", NewBlockTransTag(&testTrans))
	}
	require.Nil(t, err)

	type T struct {
		input    string
		expected string
		err      bool
	}

	tests := []T{
		{input: `{% trans "test" %}`, expected: "domain:language:test"},
		{input: `{% trans "test" %} post`, expected: "domain:language:test post"},
		{input: `pre {% trans "test" %}`, expected: "pre domain:language:test"},
		{input: `{% trans "test" as myvar %}{{myvar}}`, expected: "domain:language:test"},
		{input: `{% trans "test" context "myctx" %}`, expected: "domain:language:myctx:test"},
		{input: `{% trans "test" as othervar context "myctx" %}{{othervar}}`, expected: "domain:language:myctx:test"},

		{input: `{% blocktrans %}test{% endblocktrans %}`, expected: "domain:language:test"},
		{input: `{% blocktrans context "myctx" asvar the_title %}test{% endblocktrans%}{{the_title}}`, expected: "domain:language:myctx:test"},
	}

	for k, tst := range tests {
		tmpl, err := pongo2.FromString(tst.input)
		if tst.err {
			require.NotNilf(t, err, "test: %d, input: %s", k, tst.input)
			continue
		}
		require.Nilf(t, err, "test: %d, input: %s", k, tst.input)

		// These should be accessible in the output
		result, err := tmpl.Execute(pongo2.Context{
			"_domain":   "domain",
			"_language": "language",
		})
		require.Nilf(t, err, "test: %s input: %s", k, tst.input)
		require.Equalf(t, tst.expected, result, "test: %s input: %s", k, tst.input)
	}
}
