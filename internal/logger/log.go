// MIT License
//
// (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func Init()(logy *logrus.Logger) {

	logy = logrus.New()
	logLevel := ""
	envstr := os.Getenv("LOG_LEVEL")
	if envstr != "" {
		logLevel = strings.ToUpper(envstr)
		logy.Infof("Setting log level to: %s\n", envstr)
	}

	switch logLevel {
	case "TRACE":
		logy.SetLevel(logrus.TraceLevel)
	case "DEBUG":
		logy.SetLevel(logrus.DebugLevel)
	case "INFO":
		logy.SetLevel(logrus.InfoLevel)
	case "WARN":
		logy.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logy.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		logy.SetLevel(logrus.FatalLevel)
	case "PANIC":
		logy.SetLevel(logrus.PanicLevel)
	default:
		logy.SetLevel(logrus.ErrorLevel)
	}

	return logy
}
