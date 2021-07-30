package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func init() {
	cfg := zap.NewProductionConfig()
	ec := zap.NewProductionEncoderConfig()
	ec.TimeKey = "Timestamp"
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig = ec
	zlp, _ := cfg.Build()
	Log = zlp.Sugar()
}
