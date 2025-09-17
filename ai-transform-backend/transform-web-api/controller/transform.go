package controllers

import (
	"ai-transform-backend/data"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/log"
)

type Transform struct {
	conf *config.Config
	log  log.ILogger
	data data.IData
}

func NewTransform(conf *config.Config, log log.ILogger, data data.IData) *Transform {
	return &Transform{
		conf: conf,
		log:  log,
		data: data,
	}
}
