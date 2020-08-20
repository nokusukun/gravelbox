package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"github.com/nokusukun/roggy"
	"time"
)

var engine *gin.Engine
var log = roggy.Printer("gravel")
var cfgDocker *ini.Section
var cfg *ini.File

var CommandTimeout = time.Second * 30

func init() {
	roggy.LogLevel = 5
	engine = gin.Default()

	var err error
	cfg, err = ini.Load("gravel.ini")
	if err != nil {
		panic(err)
	}
	cfgDocker = cfg.Section("docker")

	version, err := GetDockerVersion()
	if err != nil {
		panic(fmt.Sprintf("Docker cannot be found or can't be executed through '%v': %v", cfgDocker.Key("command").String(), err))
		//panic("Docker cannot be found or can't be executed through '" + cfgDocker.Key("command").String() + "'")
	}
	log.Info("Docker version: ", version)

	CommandTimeout, err = time.ParseDuration(cfgDocker.Key("timeout").String())
	if err != nil {
		panic("Failed to parse configuration: timeout")
	}
}

func main() {

	api := engine.Group("/api")
	{
		api.GET("/docker/version", func(g *gin.Context) {
			version, err := GetDockerVersion()
			JSONReturn{
				Data:  version,
				Error: err,
			}.Send(g)
		})

		api.GET("/atoms/create/:name", func(g *gin.Context) {
			log.Info("Creating new atom, the first time usually takes a while")
			command, err := BuildAtom(g.Param("name"))
			JSONReturn{
				Data:  command,
				Error: err,
			}.Send(g)
		})

		api.GET("/atoms/list", func(g *gin.Context) {
			command, err := ListAtoms()
			JSONReturn{
				Data:  command,
				Error: err,
			}.Send(g)
		})

		api.POST("/atoms/execute", func(g *gin.Context) {
			exec := Executor{}
			err := g.ShouldBindJSON(&exec)
			if err != nil {
				JSONReturn{Error: err}.Send(g)
				return
			}

			command, err := exec.Start()
			JSONReturn{
				Data:  command,
				Error: err,
			}.Send(g)
		})
	}

	engine.Run("localhost:12375")
	roggy.Wait()
}

type JSONReturn struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error"`

	code int
}

func (j JSONReturn) Send(ctx *gin.Context) {
	if j.code == 0 {
		j.code = 200
		if j.Error != nil {
			j.code = 500
		}
	}

	if j.Error != nil {
		j.Error = fmt.Sprintf("%v", j.Error)
	}
	ctx.JSON(j.code, j)
}
