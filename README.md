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
| `GET` | `/api/rooms` | | `[{code, players, state, theme}]` public rooms |
| `POST` | `/api/rooms` | `{private, theme, rounds, duration}`, all optional | `201 {code}` |
| `GET` | `/api/themes` | | `["animaux", ...]` |
| `GET` | `/ws?room=CODE` | | websocket |

## Websocket protocol

Every packet is a JSON object with a `type` field.

Client to server:

| type | fields |
|---|---|
| `nickname` | `nickname` |
| `message` | `message` (a correct guess is not broadcast) |
| `draw` | `stroke`, see below |
| `undo` | removes the last action |
| `clear` | |
| `start` | admin only |
| `settings` | `theme` (empty for all), `rounds` (1-10), `duration` (30-180 s), admin only before the game |

Server to client:

| type | fields |
|---|---|
| `nickname` | `id`, `nickname` |
| `history` | `messages` |
| `drawing` | `strokes`, full redraw (on join and after an undo) |
| `players` | `players: [{id, nickname, admin, drawer, score}]` |
| `game` | `state`: `waiting` or `playing` |
| `settings` | `theme`, `rounds`, `duration` (s), `themes` |
| `round` | `drawer`, `word` (drawer only), `hint`, `theme`, `duration`, `remaining` (ms), `round`, `rounds` |
| `roundEnd` | `word`, `winner`, `points` |
| `gameEnd` | `players` ranking, then the room is closed |
| `message`, `system`, `error` | `message`, `nickname` |
| `draw`, `clear` | same as client |

A stroke is one drawing operation, coordinates are in `[0, 1]`:

| kind | fields |
|---|---|
| `line` | `from`, `to`, `color`, `size` (brush and eraser) |
| `spray` | `points` (up to 100 dots), `color`, `size` |
| `fill` | `from`, `color` (flood fill) |

`begin: true` marks the first stroke of an action, which is what `undo` removes.

## Rules

- The first player in a room is its admin and can start the game (2 to 10 players).
- A game has 3 rounds of 80 seconds by default, the admin can change it before starting; the drawer is drawn at random each round.
- Points: 100 scaled by the remaining time, 10 minimum.
- Empty rooms are removed after one minute.
- Words live in `words/<theme>.txt`, one per line: adding a file adds a theme.
