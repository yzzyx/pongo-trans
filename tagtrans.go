package trans

import (
	"github.com/flosch/pongo2/v6"
)

type tagTransNode struct {
	translator Translator

	countEval pongo2.IEvaluator
	withEval  map[string]pongo2.IEvaluator

	// Should context be updated?
	asValue string

	transCtx   string
	transText  string
	transEval  pongo2.IEvaluator
	pluralText string
}

func (node *tagTransNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) (transError *pongo2.Error) {
	language, ok := ctx.Private["_language"].(string)
	if !ok {
		language, _ = ctx.Public["_language"].(string)
	}

	domain, ok := ctx.Private["_domain"].(string)
	if !ok {
		domain, _ = ctx.Public["_domain"].(string)
	}

	transCtx := TransCtx{
		Language: language,
		Domain:   domain,
	}

	// Do we need to evaluate the string to be translated?
	if node.transEval != nil {
		val, evalErr := node.transEval.Evaluate(ctx)
		if evalErr != nil {
			return evalErr
		}
		node.transText = val.String()
	}

	var content string
	var err error
	if node.pluralText != "" && node.countEval != nil {
		countVal, evalErr := node.countEval.Evaluate(ctx)
		if evalErr != nil {
			return evalErr
		}

		if node.transCtx != "" {
			content = node.translator.GetNC(transCtx, node.transText, node.pluralText, countVal.Integer(), node.transCtx)
		} else {
			content = node.translator.GetN(transCtx, node.transText, node.pluralText, countVal.Integer())
		}
	} else {
		if node.transCtx != "" {
			content = node.translator.GetC(transCtx, node.transText, node.transCtx)
		} else {
			content = node.translator.Get(transCtx, node.transText)
		}
	}

	if err != nil {
		return ctx.Error(err.Error(), nil)
	}

	subCtx := pongo2.NewChildExecutionContext(ctx)

	for key, eval := range node.withEval {
		val, evalErr := eval.Evaluate(ctx)
		if evalErr != nil {
			return evalErr
		}

		subCtx.Public[key] = val.Interface()
	}

	content, err = pongo2.RenderTemplateString(content, subCtx.Public)
	if err != nil {
		return ctx.Error(err.Error(), nil)
	}

	if node.asValue != "" {
		ctx.Public[node.asValue] = content
	} else {
		_, err = writer.WriteString(content)
		if err != nil {
			return ctx.Error(err.Error(), nil)
		}
	}

	return nil
}

// NewTransTag creates a pongo2 tag for handling translations
//
// Usage:
//
//	pongo2.RegisterTag("trans", trans.NewBlockTransTag(tr))
//
//	// and the, in your templates
//	{% trans "This should be translated" %}
//	{% trans "This should be translated, and has context" with ctx="example" %}
//	{% trans "Save translation to var" as "myvar" %}{{myvar}}
func NewTransTag(translator Translator) pongo2.TagParser {
	fn := func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (tag pongo2.INodeTag, err *pongo2.Error) {
		transNode := &tagTransNode{translator: translator}

		transNode.withEval = make(map[string]pongo2.IEvaluator)

		if strToken := arguments.PeekType(pongo2.TokenString); strToken != nil {
			transNode.transText = strToken.Val
		} else if identifierToken := arguments.PeekType(pongo2.TokenIdentifier); identifierToken != nil {
			transNode.transEval, err = arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, arguments.Error("Tag 'trans' requires at least one argument, which must be a string or identifier", nil)
		}
		arguments.Consume()

		if t := arguments.Peek(pongo2.TokenKeyword, "as"); t != nil {
			arguments.Consume()
			asTag := arguments.Current()
			if asTag == nil || asTag.Typ != pongo2.TokenIdentifier {
				return nil, arguments.Error("Expected 'as' to be follow by an identifier", nil)
			}
			arguments.Consume()
			transNode.asValue = asTag.Val
		}

		if t := arguments.Peek(pongo2.TokenIdentifier, "context"); t != nil {
			arguments.Consume()
			transCtx := arguments.Current()

			if transCtx == nil || transCtx.Typ != pongo2.TokenString {
				return nil, arguments.Error("Expected 'context' to be followed by a string", nil)
			}
			transNode.transCtx = transCtx.Val
		}
		return transNode, nil
	}
	return fn
}
