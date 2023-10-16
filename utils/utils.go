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
