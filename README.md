
# Translation tags for pongo2

This project adds support for the 'trans' and 'blocktrans'-tags in [pongo2](https://github.com/flosch/pongo2).
It includes a translator based on [gotext](https://github.com/leonelquinteros/gotext) to  facilitate
the actual translation.

## Installation

```
go get github.com/yzzyx/pongo-trans
```

## Usage

The tags require an underlying translator. This package includes
an implementation that uses 'gotext' to perform the actual translations.

The gotext translator can be created with the function `NewTemplateTranslator`

When a translator is available, the tags can be created and registered by running

```
pongo2.RegisterTag("trans", trans.NewTransTag(tr))
pongo2.RegisterTag("blocktrans", trans.NewBlockTransTag(tr))
```

See template tag usage below.

## Example

```
package main

import (
    "fmt"

    "github.com/flosch/pongo2/v6"
    "github.com/yzzyx/pongo-trans"
)

func main() {
    // Read all translations from directory "locales"
    // The expected layout of this directory is as follows:
    // locales/
    //   |- langcode1/
    //     |- domain1.po
    //     |- domain2.po
    //   |- langcode2/
    //     |- domain1.po
    //     |- domain2.po
    //
    //  Where <langcode> is a language code (e.g. 'en_GB', 'sv_SE' or just 'en')
    
    t, err := NewTemplateTranslator(os.DirFS("locales"), ".")
    if err != nil {
        panic(err)
    }
    
    // If the files should be loaded from an embedded FS, add this instead:
    // //go:embed locales
    // var localeFS embed.FS
    // t, err := trans.NewTemplateTranslator(localeFS, "locales")

    pongo2.RegisterTag("trans", trans.NewTransTag(t))
    pongo2.RegisterTag("blocktrans", trans.NewBlockTransTag(t))

    tmpl, _ := pongo2.FromString(`{% trans "Please translate this!" %}`)
    result, _ := tmpl.Execute(pongo2.Context{
        "_language": "sv_SE", // No default, uses string as-is if not specified
        "_domain": "default", // Defaults to 'default', expecting a file called 'default.po' or 'default.mo'
    })

    fmt.Println("Translated result:", result)
}
```

## trans template tag


The `{% trans %}` template tag translates either a constant string (enclosed in single or double quotes) or variable content:

```
<title>{% trans "This is the title." %}</title>
```

It’s not possible to mix a template variable inside a string within {% trans %}. If your translations require strings with variables (placeholders), use {% blocktrans %} instead.

If you’d like to retrieve a translated string without displaying it, you can use the following syntax:

```
{% trans "This is the title" as the_title %}

<title>{{ the_title }}</title>
<meta name="description" content="{{ the_title }}">
```


In practice you’ll use this to get strings that are used in multiple places or should be used as arguments for other template tags or filters:

```
{% trans "starting point" as start %}
{% trans "end point" as end %}
{% trans "La Grande Boucle" as race %}

<h1>
  <a href="/" title="{% blocktrans %}Back to '{{ race }}' homepage{% endblocktrans %}">{{ race }}</a>
</h1>
<p>
{% for stage in tour_stages %}
    {% cycle start end %}: {{ stage }}{% if forloop.counter|divisibleby:2 %}<br />{% else %}, {% endif %}
{% endfor %}
</p>
```

`{% trans %}` also supports contextual markers using the context keyword:

```
{% trans "May" context "month name" %}
```

## blocktrans template tag

Contrarily to the trans tag, the blocktrans tag allows you to mark complex sentences consisting of literals and variable content for translation by making use of placeholders:

```
{% blocktrans %}This string will have {{ value }} inside.{% endblocktrans %}
```

To translate a template expression – say, accessing object attributes or using template filters – you need to bind the expression to a local variable for use within the translation block. Examples:

```
{% blocktrans with amount=article.price %}
That will cost $ {{ amount }}.
{% endblocktrans %}

{% blocktrans with myvar=value %}
This will have {{ myvar }} inside.
{% endblocktrans %}
```

You can use multiple expressions inside a single blocktrans tag:

```
{% blocktrans with book_t=book.title author_t=author.name %}
This is {{ book_t }} by {{ author_t }}
{% endblocktrans %}
```

This tag also provides for pluralization. To use it:

Designate and bind a counter value with the name count. This value will be the one used to select the right plural form.
Specify both the singular and plural forms separating them with the {% plural %} tag within the {% blocktrans %} and {% endblocktrans %} tags.

An example:

```
{% blocktrans count counter=cnt %}
There is only one {{ name }} object.
{% plural %}
There are {{ counter }} {{ name }} objects.
{% endblocktrans %}
```

A more complex example:

```
{% blocktrans with amount=article.price count years=i.length %}
That will cost $ {{ amount }} per year.
{% plural %}
That will cost $ {{ amount }} per {{ years }} years.
{% endblocktrans %}
```

(documentation adapted from the original Django documentation)
