package schedule

import (
	log "github.com/sirupsen/logrus"
)

type gocronLogger struct {
}

func (l gocronLogger) Debug(msg string, args ...any) {
	log.Debugf("gocron: %s: %v", msg, args)
}
func (l gocronLogger) Info(msg string, args ...any) {
	log.Infof("gocron: %s: %v", msg, args)
}
func (l gocronLogger) Warn(msg string, args ...any) {
	log.Warnf("gocron: %s: %v", msg, args)
}
func (l gocronLogger) Error(msg string, args ...any) {
	log.Errorf("gocron: %s: %v", msg, args)
}
