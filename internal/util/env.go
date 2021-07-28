// util package contains common/convenient routines to aid in development and debugging
package util

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// LookupEnv returns an envrionment variable value.
// if the env variable is not found, returns the default
func LookupEnv(env, defaultVal string) (val string) {
	var ok bool
	if val = defaultVal; env == "" {
		log.Debug("Caller provided blank environment variable name")
		return
	}
	if val, ok = os.LookupEnv(env); !ok {
		log.Debug(env, " Notfound using default value ", defaultVal)
		val = defaultVal
	}
	log.Debug(env, " value is ", val)
	return
}

// LookupEnvInt returns an envrionment variable value as an int.
// if the env variable is not found, returns the default
func LookupEnvInt(env string, defaultVal int) (n int) {
	var ok bool
	var val string
	var err error

	if n = defaultVal; env == "" {
		log.Debug("Caller provided blank environment variable name")
		return
	}
	if val, ok = os.LookupEnv(env); !ok {
		log.Debug(env, " Notfound using default value ", defaultVal)
		n = defaultVal
		return
	}
	if n, err = strconv.Atoi(val); err != nil {
		log.Error(env, " value is not an integer", val, err)
	}
	log.Debug(env, " value is ", n)
	return
}
