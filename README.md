# GuessWhat

Online Pictionary-like game: a Go websocket server and an HTML/CSS/JS front.
One player draws a random word, the first one to guess it wins points proportional to their speed.

## Run

```sh
go run .            # http://localhost:8080
PORT=3000 go run .  # custom port
go test -race ./...
```

The front is embedded in the binary (`go build` produces a standalone executable).
On a fatal error the server writes to stderr and exits with code 84.

## HTTP API

| Method | Path | Body | Response |
|---|---|---|---|
| `GET` | `/api/rooms` | | `[{code, players, state}]` public rooms |
| `POST` | `/api/rooms` | `{private}` | `201 {code}` |
| `GET` | `/ws?room=CODE` | | websocket |

## Websocket protocol

Every packet is a JSON object with a `type` field.

Client to server:

| type | fields |
|---|---|
| `nickname` | `nickname` |
| `message` | `message` (a correct guess is not broadcast) |
| `draw` | `stroke: {from: {x, y}, to: {x, y}, color, size}`, coordinates in `[0, 1]` |
| `clear` | |
| `start` | admin only |

Server to client:

| type | fields |
|---|---|
| `nickname` | `id`, `nickname` |
| `history` | `messages` |
| `drawing` | `strokes` |
| `players` | `players: [{id, nickname, admin, drawer, score}]` |
| `game` | `state`: `waiting` or `playing` |
| `round` | `drawer`, `word` (drawer only), `hint`, `duration`, `remaining` (ms), `round`, `rounds` |
| `roundEnd` | `word`, `winner`, `points` |
| `gameEnd` | `players` ranking, then the room is closed |
| `message`, `system`, `error` | `message`, `nickname` |
| `draw`, `clear` | same as client |

## Rules

- The first player in a room is its admin and can start the game (2 to 10 players).
- A game has 3 rounds of 80 seconds; the drawer is drawn at random each round.
- Points: 100 scaled by the remaining time, 10 minimum.
- Empty rooms are removed after one minute.
