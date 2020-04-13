package main

import (
	"flag"
	"os"
	"strings"

	"TerminalBuddyServer/config"
	"TerminalBuddyServer/internal"

	log "github.com/sirupsen/logrus"
)

func main() {
	port := flag.Int("port", 8080, "port number")
	dbTypeParam := flag.String("db-type", "ps", "in memory DB (mem) or Postgres (ps)")
	recreateDb := flag.Bool("recreate-db", false, "drop current DB and create from scratch")
	configPath := flag.String("cfg-path", "cmd/config.yaml", "yaml config file path")
	flag.Parse()
	log.Debugf("using port %d", *port)
	if *recreateDb {
		log.Warn("will recreate DB")
	}

	configData, err := config.ReadYamlConfig(*configPath)
	if err != nil {
		log.Fatalf("cannot open/read yaml conf file: %s", err.Error())
	}

	tbConfig, err := config.NewTbConfig(configData)
	if err != nil {
		log.Fatalf("error getting config: %s", err.Error())
	}
	if tbConfig == nil {
		panic("received config is nil")
	}

	if tbConfig.LogOutput() == config.FileLogOutput {
		loggingSetup(tbConfig.LogFilePath(), tbConfig.LogLevel())
	} else {
		setupTraceLoggingToStdout()
	}
	log.Debug("starting ...")

	if *dbTypeParam != "ps" && *dbTypeParam != "mem" {
		panic("unknown db type: " + *dbTypeParam)
	}

	var dbType = internal.InMemDB
	if *dbTypeParam == "ps" {
		dbType = internal.PsDB
	}

	dbPassword := os.Getenv("TB_DB_PASSWORD")
	if dbType == internal.PsDB && len(dbPassword) == 0 {
		log.Fatal("DB password not set. use env var TB_DB_PASSWORD to set it")
	}

	server := internal.NewServer(tbConfig, dbType, dbPassword, *recreateDb)
	server.Serve()
}

func setupTraceLoggingToStdout() {
	loggingSetup("", "")
}

func loggingSetup(logFileName string, logLevel string) {
	switch strings.ToLower(logLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.TraceLevel)
	}

	if logFileName == "" {
		log.SetOutput(os.Stdout)
		return
	}

	if !strings.HasSuffix(logFileName, ".log") {
		logFileName += ".log"
	}

	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Panicf("failed to open log file %q: %s", logFileName, err)
	}

	log.SetOutput(logFile)
}
