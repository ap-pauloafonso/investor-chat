package utils

import (
	"log/slog"
	"os"
)

type ErrorMessage struct {
	ErrorMessage string `json:"errorMessage"`
}

func LogErrorFatal(err error) {
	slog.Error(err.Error())
	os.Exit(1)
}

func ExecAndPrintErr(fn func() error) {
	err := fn()
	if err != nil {
		slog.Error("error while executing fn", err)
	}
}
