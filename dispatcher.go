package api

import (
	"context"
	"log"
	"time"
)

func (c *initWorker) DispatchQueue(ctx context.Context) {
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
		err := c.DispatchJob()
		if err != nil {
			if err.Error() != "no rows in result set" {
				log.Println(err)
				return
			}
		}

		counter = counter + 1
		diff := time.Now().Sub(last)
		if counter > 100 {
			counter = 0

			if diff < time.Second {
				time.Sleep(30) // sleep if it happens too fast
			}
		}
		last = time.Now()
	}
}

func (c *initWorker) DispatchJob() error {
	now := time.Now()
	t1 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	t2 := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	rows, err := c.api.Db.Query(`
		SELECT id, 
			name,
			message_id,
			number,
		FROM queue_item
			WHERE activation_time > $1 AND activation_time < $2 
			AND canceled_time is NULL AND status = 'NEW'
	`, t1, t2)

	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var id, name, messageId, number string
		err = rows.Scan(&id, &name, &messageId, &number)
		if err != nil {
			log.Println(err)
			return err
		}
		status := "DONE"
		log.Println(name[:4])
		if name[:4] == "M000" {
			err = c.InsertIntoQueue(name, id)
			if err != nil {
				log.Println(err)
				return err
			}

			status = "INQUEUE"
		}

		if name[:5] != "S0001" {
			items, err := c.api.GetEventsByTrigger(name[:5])
			if err != nil {
				log.Println(err)
				return err
			}

			err = c.api.ProcessQueue(items, id, number)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		err = c.api.UpdateQueueItem(id, status)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (c *initWorker) InsertIntoQueue(name, id string) error {
	_, err := c.api.Db.Exec(`
		INSERT INTO worker_queue (name, queue_id) VALUES ($1, $2)
	`, name, id)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
