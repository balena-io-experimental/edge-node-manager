package hook

import (
	"io/ioutil"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/config"
	"github.com/resin-io/edge-node-manager/supervisor"
)

type Hook struct {
	ResinUUID string
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	serialised, _ := entry.Logger.Formatter.Format(entry)
	message := regexp.MustCompile(`\r?\n`).ReplaceAllString((string)(serialised), "")
	supervisor.DependentDeviceLog(h.ResinUUID, message)

	return nil
}

func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

func Create(resinUUID string) *logrus.Logger {
	log := logrus.New()
	log.Out = ioutil.Discard
	log.Level = config.GetDependentLogLevel()
	log.Formatter = &logrus.TextFormatter{ForceColors: true, DisableTimestamp: true}
	log.Hooks.Add(&Hook{
		ResinUUID: resinUUID,
	})

	return log
}
