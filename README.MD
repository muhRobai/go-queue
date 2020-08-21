## How to start ? 
- you must install docker, docker-compose, and makefile on your computer
- i'm using ubuntu 20

## write in your CLI 
- make prepare 
- make migrate
- make run-api
- make run-dispatcher
- make run-worker

## to send message 
- send to `/api/create-queue` with payload
```
 {  
    {
    "events": "S0001", --> initial events for first message events
    "item": {
        "message": "omama olala", --> your message
        "email": "bambang@getnada.com", --> email will send the message
        "number": "202020011" --> number customer
    }
}
```

- for delete message `/api/delete-queue` with payload
```
{
    "events": "M0001", --> events name
    "times": 1597999435, --> time scheduled
    "number": "202020012" --> customer number
}
```

- for call message `/api/call-queue` with payload (same with delete)
```
{
    "events": "M0001", --> events name
    "times": 1597999435, --> time scheduled
    "number": "202020012" --> customer number
}
```

