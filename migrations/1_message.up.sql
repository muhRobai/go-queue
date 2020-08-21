DO $$ BEGIN
  CREATE EXTENSION pgcrypto;
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;


CREATE TABLE worker_queue (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    queue_id TEXT NOT NULL
);

CREATE TABLE events (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule TEXT NOT NULL,
    description TEXT NOT NULL,
    activation INT NOT NULL,
    canceled TEXT NULL,
    trigger_by TEXT NOT NULL
);

CREATE TABLE queue_item (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    activation_time TIMESTAMP NOT NULL,
    canceled_time TIMESTAMP NULL,
    message_id TEXT NOT NULL,
    status TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_name ON queue_item(name);

CREATE TABLE message_item (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    message TEXT NOT NULL,
    email TEXT NOT NULL,
    number TEXT NOT NULL
);


INSERT INTO events (schedule, description, activation, canceled, trigger_by) VALUES ('M0001', 'message 1', 1, 'C0001', 'S0001');
INSERT INTO events (schedule, description, activation, canceled, trigger_by) VALUES ('M0002', 'message 2', 1, 'C0002', 'S0002');
INSERT INTO events (schedule, description, activation, canceled, trigger_by) VALUES ('M0003', 'message 3', 1, 'C0003', 'S0003');

INSERT INTO events (schedule, description, activation, trigger_by) VALUES ('S0001', 'start message 1', 1, 'EXT');
INSERT INTO events (schedule, description, activation, trigger_by) VALUES ('S0002', 'start message 2', 15, 'S0001');
INSERT INTO events (schedule, description, activation, trigger_by) VALUES ('S0003', 'start message 3', 30, 'S0001');

INSERT INTO events (schedule, description, activation, trigger_by) VALUES ('C0001', 'cancel message 1', 1, 'EXT');
INSERT INTO events (schedule, description, activation, trigger_by) VALUES ('C0002', 'cancel message 2', 1, 'EXT');
INSERT INTO events (schedule, description, activation, trigger_by) VALUES ('C0003', 'cancel message 3', 1, 'EXT');




