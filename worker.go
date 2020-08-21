package api

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"gopkg.in/gomail.v2"
)

func (c *initWorker) ProcessMessage(ctx context.Context) {
	last := time.Now()
	counter := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Dispatcher worker canceled")
			return
		default:
			// empty
		}
		err := c.SendMessage()
		if err != nil {
			if err.Error() != "no rows in result set" {
				log.Println(err)
				break
			}
		}

		counter = counter + 1
		diff := time.Now().Sub(last)
		if counter > 100 {
			counter = 0

			if diff < time.Second {
				time.Sleep(60) // sleep if it happens too fast
			}
		}
		last = time.Now()
	}
}

func (c *initWorker) SendMessage() error {
	tx, err := c.api.Db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	defer tx.Rollback()
	var queueId string
	err = tx.QueryRow(`
		DELETE FROM worker_queue
			WHERE id = (
				SELECT id FROM worker_queue
					FOR UPDATE SKIP LOCKED
					LIMIT 1
			)
		RETURNING queue_id
	`).Scan(&queueId)

	if err != nil {
		return err
	}

	if queueId != "" {
		status := "DONE"
		messageId, err := c.GetQueueItemByID(queueId)
		if err != nil {
			status = "ERROR"
		}

		err = c.GetMessageById(messageId)
		if err != nil {
			log.Println(err)
			status = "ERROR"
		}

		err = c.api.UpdateQueueItem(queueId, status)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

func (c *initWorker) GetQueueItemByID(id string) (string, error) {
	var messageId string
	err := c.api.Db.QueryRow(`
		SELECT message_id FROM queue_item WHERE id = $1 AND status = 'INQUEUE'
	`, id).Scan(&messageId)

	if err != nil {
		return "", err
	}

	return messageId, nil
}

func (c *initWorker) GetMessageById(id string) error {
	rows, err := c.api.Db.Query(`
		SELECT id,
			message,
			email
		FROM message_item
			WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var id, message, email string
		err = rows.Scan(&id, &message, &email)
		if err != nil {
			log.Println(err)
			return err
		}

		m := gomail.NewMessage()
		m.SetHeader("From", "example@getnada.com")
		m.SetHeader("To", email)
		m.SetHeader("Subject", "ini email")

		m.SetBody("text/plain", message)
		err := c.send(m)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (c *initWorker) send(m *gomail.Message) error {

	d := gomail.NewDialer(c.api.config.smtpHost, c.api.config.smtpPort, c.api.config.smtpUser, c.api.config.smtpPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)

}
