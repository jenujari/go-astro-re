package main

import (
	"github.com/hyperjumptech/grule-rule-engine/logger"
	c "github.com/jenujari/go-astro-re/config"
	"github.com/jenujari/go-astro-re/server"

	rtc "github.com/jenujari/runtime-context"
)

var pc *rtc.ProcessContext

func init() {
	rtc.InitProcessContext(c.GetLogger())
	logger.SetLogger(c.GetLogger())
}

func main() {
	pc = rtc.GetMainProcess()
	srv := server.GetServer()
	pc.Run(server.RunServer)
	c.GetLogger().Println("Server is running at ", srv.Addr)

	pc.WaitForFinish()
}
