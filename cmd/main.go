package main

import (
	"TerminalBuddyServer/internal"
	"flag"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	setupTraceLoggingToStdout()
	log.Debug("starting ...")

	port := flag.Int("port", 8080, "port number")
	recreateDb := flag.Bool("recreate-db", false, "drop current DB and create from scratch")
	flag.Parse()
	log.Debugf("using port %d", *port)
	if *recreateDb {
		log.Warn("will recreate DB")
	}

	server := internal.NewServer(*recreateDb)
	server.Serve(*port)
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
