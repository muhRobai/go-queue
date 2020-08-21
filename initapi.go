package api

import (
	"log"
	"os"
	"strconv"
)

func (c *initAPI) initConfig() {
	c.config.smtpHost = os.Getenv("SMTP_HOST")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))

	if err != nil {
		port = 587
	}
	c.config.smtpPort = port
	c.config.smtpUser = os.Getenv("SMTP_USER")
	c.config.smtpPassword = os.Getenv("SMTP_PASS")
}

func CreateAPI() (*initAPI, error) {
	c := initAPI{}
	c.initConfig()

	return &c, nil
}

func CreateWorker() (*initWorker, error) {
	c := initWorker{}
	api, err := CreateAPI()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	api.initDB()
	c.api = api

	return &c, nil
}
