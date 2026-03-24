package Log

import (
	"go.uber.org/zap"
)

var ZLog *zap.Logger

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		ZLog = zap.NewNop()
		return
	}
	ZLog = logger

}
