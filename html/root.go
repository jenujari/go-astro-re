package html

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"time"
)

var (
	//go:embed template/*.html
	htmlTemplate embed.FS

	//go:embed static/js/*.js static/css/*.css
	staticFS embed.FS

	tpl *template.Template
)

func init() {
	tpl = template.New("init.html").Funcs(getFuncMap())
}

func GetViewsFs() fs.FS {
	f, _ := fs.Sub(htmlTemplate, "template")
	return f
}

func GetAssetsFs() fs.FS {
	f, _ := fs.Sub(staticFS, "static")
	return f
}

func GetTpl() *template.Template {
	return tpl
}

func getFuncMap() template.FuncMap {
	var funcMap = template.FuncMap{}

	funcMap["title"] = func(v, d string) string {
		if v == "" {
			return d
		}
		return v
	}

	funcMap["formatDate"] = func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	}

	funcMap["toJson"] = func(obj any) string {
		b, e := json.MarshalIndent(obj, "", "  ")
		if e != nil {
			return e.Error()
		}
		return string(b)
	}

	funcMap["toJsonJS"] = func(obj any) template.JS {
		b, e := json.Marshal(obj)
		if e != nil {
			return template.JS("[]")
		}
		return template.JS(b)
	}

	return funcMap
}
