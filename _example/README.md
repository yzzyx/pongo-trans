
Translation example project
============================

This project shows how the pongo-trans package can be utilised to add
translation to HTML templates in go.

Updating translations
---------------------

If changes are made to the templates, and new translatable strings
are added, run the following commands to update the translations:

```
## NOTE! Make sure that gettext is installed before running makemessage
$ go install github.com/yzzyx/makemessage
$ makemessage -l sv_SE -t templates
```

New languages can also be added by suppling the new language-code to makemessage