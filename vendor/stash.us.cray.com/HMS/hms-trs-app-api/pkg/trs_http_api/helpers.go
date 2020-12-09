package trs_http_api

import "github.com/sirupsen/logrus"

func SendDelayedError(tct taskChannelTuple, logger *logrus.Logger) {
	err := *tct.task.Err
	err.Error()
	logger.Errorf("Receieved error for task: %+v, error: %s", tct.task.ToHttpKafkaTx(), err.Error())
	tct.taskListChannel <- tct.task
	return
}