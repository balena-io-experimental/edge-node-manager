package hook

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/resin-io/edge-node-manager/supervisor"
)

type Hook struct {
	ResinUUID string
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	serialised, _ := entry.Logger.Formatter.Format(entry)
	supervisor.DependantDeviceLog(h.ResinUUID, (string)(serialised))

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
	log.Hooks.Add(&Hook{
		ResinUUID: resinUUID,
	})

	return log
}
