package server

import (
	"errors"
	"fmt"
	"net/http"

	c "github.com/jenujari/go-astro-re/config"
	rulesengine "github.com/jenujari/go-astro-re/rules-engine"
	rtc "github.com/jenujari/runtime-context"
)

var (
	server      *http.Server
	router      *http.ServeMux
	ruleRuntime *rulesengine.Runtime
)

const (
	ruleRootDir    = "rules"
	ruleSetName    = "example"
	ruleSetVersion = "0.0.1"
)

func init() {
	runtime, err := rulesengine.NewRuntimeFromDir(ruleRootDir, ruleSetName, ruleSetVersion)
	if err != nil {
		panic(fmt.Errorf("init rule runtime: %w", err))
	}
	ruleRuntime = runtime

	server = &http.Server{
		Addr:              ":8899",
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		MaxHeaderBytes:    0,
	}

	router = http.NewServeMux()

	router.Handle("/static/", staticHander())

	router.HandleFunc("/", indexhandler)
	router.HandleFunc("/rule", ruleHandler)

	server.Handler = GlobalRequestContextSetter(router)
	c.GetLogger().Println("server initialization complete.")
}

func RunServer() {
	pc := rtc.GetMainProcess()

	go func(cmdx *rtc.ProcessContext) {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			cmdx.FatalErrorChan <- fmt.Errorf("ListenAndServe(): %v", err)
		}
	}(pc)

	<-pc.CTX.Done()
	c.GetLogger().Println("shutting down server...")
	if err := server.Shutdown(pc.CTX); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
	c.GetLogger().Println("server shutdown complete...")
}

func GetServer() *http.Server {
	return server
}

func GlobalRequestContextSetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// ctx = context.WithValue(ctx, "services", lib.GetAllServices())

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
