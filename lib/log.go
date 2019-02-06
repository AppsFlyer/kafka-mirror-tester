package lib

import log "github.com/sirupsen/logrus"

// initialize logs
func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}
