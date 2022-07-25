package logutil

import (
	"context"
	"fmt"
	"github.com/anon-org/developing-api-services-with-golang/util/idutil"
	"log"
	"os"
)

const (
	defaultPrefix string = "[MAIN] "
	defaultFlag   int    = log.LstdFlags | log.Lshortfile | log.Lmsgprefix
)

type ctxLogger struct{}

var (
	ctxLoggerKey      *ctxLogger  = &ctxLogger{}
	stdLoggerInstance *log.Logger = log.New(os.Stdout, defaultPrefix, defaultFlag)
)

func NewStdLogger() *log.Logger {
	return stdLoggerInstance
}

func NewCtxLogger() *log.Logger {
	id := idutil.MustGenerateID(8)
	prefix := fmt.Sprintf("[%s] ", id)
	return log.New(os.Stdout, prefix, defaultFlag)
}

func PutCtxLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, logger)
}

func GetCtxLogger(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value(ctxLoggerKey).(*log.Logger)
	if !ok {
		logger = NewCtxLogger()
	}
	return logger
}
