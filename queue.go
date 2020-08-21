package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

func (b *initAPI) nowTs(now time.Time) string {
	return now.Format(time.RFC3339)
}

func (c *initAPI) CreateQueue(ctx context.Context, req *QueueRequest) (*QueueResponse, error) {
	if req.Event == "" {
		return nil, errors.New("missing-events")
	}

	id, err := c.CreateMessage(req.Item)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var items []*QueueItem
	events, err := c.GetEventsBySchedule(req.Event)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if events.TriggerBy == "EXT" {
		items, err = c.GetEventsByTrigger(req.Event)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	items = append(items, events)

	err = c.ProcessQueue(items, id, req.Item.Number)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &QueueResponse{
		Id:        id,
		Timestamp: time.Now().Unix(),
	}, nil
}

func (c *initAPI) ProcessQueue(req []*QueueItem, id, number string) error {
	if len(req) == 0 {
		return errors.New("missing-queue-item")
	}

	for _, item := range req {
		err := c.CreateQueueItem(item, id, number, "NEW")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (c *initAPI) CreateQueueItem(req *QueueItem, id, number, status string) error {
	now := time.Now()
	name := fmt.Sprintf("%s%s%s", req.Schedule, number, dateFormat(now.Unix()))
	_, err := c.Db.Exec(`
		INSERT INTO queue_item (name, activation_time, message_id, status) VALUES ($1, $2, $3, $4)
	`, name, c.nowTs(time.Unix(req.ActivationTime, 0)), id, status)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *initAPI) GetEventsByTrigger(name string) ([]*QueueItem, error) {
	rows, err := c.Db.Query(`
		SELECT
			schedule, 
			description, 
			activation,
			trigger_by
		FROM events
			WHERE trigger_by = $1
	`, name)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	var items []*QueueItem
	for rows.Next() {
		var item QueueItem
		err = rows.Scan(
			&item.Schedule,
			&item.Description,
			&item.Activation,
			&item.TriggerBy,
		)

		if err != nil {
			log.Println(err)
			return nil, err
		}
		now := time.Now()
		scheduled := time.Date(now.Year(), now.Month(), now.Day()+(int(item.Activation)-1), now.Hour(), 0, 0, 0, time.UTC)

		item.ActivationTime = scheduled.Unix()

		items = append(items, &item)
	}

	if len(items) == 0 {
		return nil, errors.New("missing-events")
	}

	return items, nil
}

func (c *initAPI) GetEventsBySchedule(name string) (*QueueItem, error) {
	rows, err := c.Db.Query(`
		SELECT
			schedule, 
			description, 
			activation,
			trigger_by
		FROM events
			WHERE schedule = $1
	`, name)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	var items []*QueueItem
	for rows.Next() {
		var item QueueItem
		err = rows.Scan(
			&item.Schedule,
			&item.Description,
			&item.Activation,
			&item.TriggerBy,
		)

		if err != nil {
			log.Println(err)
			return nil, err
		}

		now := time.Now()
		scheduled := time.Date(now.Year(), now.Month(), now.Day()+(int(item.Activation)-1), now.Hour(), 0, 0, 0, time.UTC)

		item.ActivationTime = scheduled.Unix()
		items = append(items, &item)
	}

	if len(items) == 0 {
		return nil, errors.New("missing-events")
	}

	return items[0], nil
}

func (c *initAPI) CreateMessage(req *MessageItem) (string, error) {
	if req.Message == "" {
		return "", errors.New("missing-message")
	}

	if req.Email == "" {
		return "", errors.New("missing-email")
	}

	if req.Number == "" {
		return "", errors.New("missing-number")
	}
	var id string
	err := c.Db.QueryRow(`
		INSERT INTO message_item (message, email, number) VALUES ($1, $2, $3) RETURNING id
	`, req.Message, req.Email, req.Number).Scan(&id)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return id, nil
}

func (c *initAPI) CreateAuthenticate() {

}

func (c *initAPI) DeleteQueue(ctx context.Context, req *MessageRequest) (*QueueResponse, error) {
	name := fmt.Sprintf("%s%s%s", req.Events, req.Number, dateFormat(req.Times))
	now := time.Now()
	var id string
	err := c.Db.QueryRow(`
		UPDATE queue_item SET canceled_time = $1 WHERE name = $2 RETURNING id
	`, c.nowTs(now), name).Scan(&id)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = c.UpdateQueueItem(id, "CANCELED")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &QueueResponse{
		Id:        id,
		Timestamp: now.Unix(),
	}, nil
}

func (c *initAPI) CallMessage(ctx context.Context, req *MessageRequest) (*QueueResponse, error) {
	name := fmt.Sprintf("%s%s%s", req.Events, req.Number, dateFormat(req.Times))
	now := time.Now()
	var id string
	err := c.Db.QueryRow(`
		UPDATE queue_item SET activation_time = $1, canceled_time = NULL WHERE name = $2 RETURNING id
	`, c.nowTs(now), name).Scan(&id)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = c.UpdateQueueItem(id, "NEW")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &QueueResponse{
		Id:        id,
		Timestamp: now.Unix(),
	}, nil
}

func dateFormat(now int64) string {
	date := time.Unix(now, 0)
	WIB := time.FixedZone("UTC+7", 7*60*60)
	date = date.In(WIB)
	month := fmt.Sprintf("%d", int(date.Month()))
	if int(date.Month()) < 10 {
		month = fmt.Sprintf("0%s", month)
	}

	days := date.Day() + 1
	if date.Hour() > 16 {
		days += 1
	}

	day := fmt.Sprintf("%d", days)
	if days < 10 {
		day = fmt.Sprintf("0%s", day)
	}

	return fmt.Sprintf("%d%s%s",
		date.Year(),
		month,
		day,
	)
}

func (c *initAPI) UpdateQueueItem(id, status string) error {
	_, err := c.Db.Exec(`
		UPDATE queue_item SET status = $1 WHERE id = $2
	`, status, id)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
