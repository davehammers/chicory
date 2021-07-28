// util package contains common/convenient routines to aid in development and debugging
package util

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// environment variable name set to true/false to output log entries in JSON format
	// defaults to true
	LOG_JSON = "LOG_JSON"

	// environment variable name set the true/false to log debug output
	// defaults to false
	LOG_DEBUG = "LOG_DEBUG"
)

// LogSetup - initialize logging configuration
func LogSetup() {
	// get our logging enviornment
	logJSON := strings.ToLower(LookupEnv(LOG_JSON, "true"))
	logDebug := strings.ToLower(LookupEnv(LOG_DEBUG, "false"))

	// setup JSON logger for SumoLogic parsing
	if logJSON == "true" {
		jsonFmt := &log.JSONFormatter{
			DisableTimestamp: true,
			FieldMap: log.FieldMap{
				log.FieldKeyLevel: "level",
				log.FieldKeyFunc:  "caller",
				log.FieldKeyMsg:   "message",
			},
			PrettyPrint: false,
		}
		log.SetFormatter(jsonFmt)
	}
	if logDebug == "true" {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	}
}
