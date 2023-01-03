package trans

import (
	"strings"

	"github.com/flosch/pongo2/v6"
)

// NewBlockTransTag creates a new pongo2 block translator tag
//
// Usage:
//
//	pongo2.RegisterTag("blocktrans", trans.NewBlockTransTag(tr))
//
//	// and the, in your templates
//	{% blocktrans %}This is a block that should be translated.
//	It can contain newlines and {{variables}}
//	{% endblocktrans %}
func NewBlockTransTag(translator Translator) pongo2.TagParser {
	fn := func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (tag pongo2.INodeTag, err *pongo2.Error) {
		transNode := &tagTransNode{translator: translator}

		transNode.withEval = make(map[string]pongo2.IEvaluator)

		countToken := arguments.Peek(pongo2.TokenIdentifier, "count")
		if countToken != nil {
			arguments.Consume()

			keyToken := arguments.MatchType(pongo2.TokenIdentifier)
			if keyToken == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}

			if arguments.Match(pongo2.TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}

			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			transNode.withEval[keyToken.Val] = valueExpr
			transNode.countEval = valueExpr
		}

		if t := arguments.Peek(pongo2.TokenIdentifier, "context"); t != nil {
			arguments.Consume()
			transCtx := arguments.Current()

			if transCtx == nil || transCtx.Typ != pongo2.TokenString {
				return nil, arguments.Error("Expected 'context' to be followed by a string", nil)
			}
			transNode.transCtx = transCtx.Val
		}

		if t := arguments.Peek(pongo2.TokenIdentifier, "asvar"); t != nil {
			arguments.Consume()
			asTag := arguments.Current()
			if asTag == nil || asTag.Typ != pongo2.TokenIdentifier {
				return nil, arguments.Error("Expected 'as' to be follow by an identifier", nil)
			}
			arguments.Consume()
			transNode.asValue = asTag.Val
		}

		text, endTag, err := getTextUntil(doc, "plural", "endblocktrans")
		if err != nil {
			return nil, err
		}
		transNode.transText = text

		if endTag == "plural" {
			text, _, err = getTextUntil(doc, "endblocktrans")
			if err != nil {
				return nil, err
			}
			transNode.pluralText = text
		}

		return transNode, nil
	}
	return fn
}

func getTextUntil(doc *pongo2.Parser, names ...string) (str string, endTagName string, err *pongo2.Error) {
	var prevLine, prevEndCol int
	for doc.Remaining() > 0 {
		// New tag, check whether we have to stop wrapping here
		if doc.Peek(pongo2.TokenSymbol, "{%") != nil {
			tagIdent := doc.PeekTypeN(1, pongo2.TokenIdentifier)

			if tagIdent != nil {
				// We've found a (!) end-tag

				found := false
				for _, n := range names {
					if tagIdent.Val == n {
						found = true
						endTagName = n
						break
					}
				}

				// We only process the tag if we've found an end tag
				if found {
					// Okay, endtag found.
					doc.ConsumeN(2) // '{%' tagname

					for {
						if doc.Match(pongo2.TokenSymbol, "%}") != nil {
							// Done skipping, exit.
							return str, endTagName, nil
						}
					}
				}
			}
		}

		t := doc.Current()
		if t == nil {
			break
		}

		doc.Consume()

		// Do not remove white-space between '{{'-symbol and identifier
		//  e.g. do not turn '{{ blah }}' into '{{blah}}'
		if (t.Typ == pongo2.TokenIdentifier || t.Typ == pongo2.TokenSymbol) && t.Line == prevLine && t.Col != prevEndCol {
			str += strings.Repeat(" ", t.Col-prevEndCol)
		}

		str += t.Val

		prevLine = t.Line
		prevEndCol = t.Col + len(t.Val)
	}
	return str, endTagName, doc.Error("Unexpected EOF.", nil)
}
