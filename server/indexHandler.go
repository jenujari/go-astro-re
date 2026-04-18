package server

import (
	"net/http"

	"github.com/jenujari/go-astro-re/config"
	c "github.com/jenujari/go-astro-re/config"
	"github.com/jenujari/go-astro-re/html"
)

func staticHander() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.FS(html.GetAssetsFs())))
}

func indexhandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := html.GetTpl().Clone()
	if err != nil {
		c.GetLogger().Println("template clone failed", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tpl, err = tpl.ParseFS(html.GetViewsFs(), "index.html")

	if err != nil {
		config.GetLogger().Println("template not found", err)
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	err = tpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		config.GetLogger().Println("template execution failed", err)
		http.Error(w, "template execution failed", http.StatusInternalServerError)
	}
}
