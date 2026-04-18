package main

import (
	"github.com/hyperjumptech/grule-rule-engine/logger"
	c "github.com/jenujari/go-astro-re/config"
	"github.com/jenujari/go-astro-re/server"

	rtc "github.com/jenujari/runtime-context"
)

func main() {
	appLogger := c.GetLogger()
	rtc.InitProcessContext(appLogger)
	logger.SetLogger(appLogger)

	pc := rtc.GetMainProcess()
	srv := server.GetServer()
	pc.Run(server.RunServer)
	appLogger.Println("Server is running at ", srv.Addr)

	pc.WaitForFinish()
}
