# GuessWhat

Online Pictionary-like game: a Go websocket server and an HTML/CSS/JS front.
One player draws a random word, the first one to guess it wins points proportional to their speed.

## Run

```sh
go run .                          # http://localhost:8080
PORT=3000 go run .                # custom port
ADMIN_PASSWORD=secret go run .    # enables http://localhost:8080/admin.html
go test -race ./...
```

| Variable | Default | |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `WORDS_DIR` | `data/words` | word lists, seeded with the default ones on first run |
| `ADMIN_PASSWORD` | | admin page and API are disabled when empty |

The front is embedded in the binary (`go build` produces a standalone executable).
On a fatal error the server writes to stderr and exits with code 84.

## Deploy

`render.yaml` is a [Render](https://render.com) blueprint: create a Blueprint from the repository, then set `ADMIN_PASSWORD` in the dashboard.
On the free plan the disk is not persistent, so words edited from the admin page are reset on each deploy or restart.

## HTTP API

| Method | Path | Body | Response |
|---|---|---|---|
| `GET` | `/api/rooms` | | `[{code, players, state, theme, teams}]` public rooms |
| `POST` | `/api/rooms` | `{private, theme, rounds, duration, teams}`, all optional | `201 {code}` |
| `GET` | `/api/themes` | | `["animaux", ...]` |
| `GET` | `/ws?room=CODE` | | websocket |

Admin API, with an `Authorization: Bearer <ADMIN_PASSWORD>` header:

| Method | Path | Body | Response |
|---|---|---|---|
| `GET` | `/api/admin/themes` | | `{theme: [words]}` |
| `POST` | `/api/admin/themes` | `{name, words}` | `201 [words]` |
| `DELETE` | `/api/admin/themes/{theme}` | | `204` |
| `POST` | `/api/admin/themes/{theme}/words` | `{word}` | `201 [words]` |
| `DELETE` | `/api/admin/themes/{theme}/words/{word}` | | `204` |

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
| `avatar` | `avatar: {color, eyes, mouth}`, indexes in `[0, 12)`, `[0, 6)`, `[0, 6)` |
| `team` | `team`: 1 or 2, before the game |
| `start` | admin only |
| `settings` | `theme` (empty for all), `rounds` (1-10), `duration` (30-180 s), `teams` (bool), admin only before the game |

Server to client:

| type | fields |
|---|---|
| `nickname` | `id`, `nickname` |
| `history` | `messages` |
| `drawing` | `strokes`, full redraw (on join and after an undo) |
| `players` | `players: [{id, nickname, admin, drawer, score, avatar, team}]`, `teams` scores when teams are on |
| `game` | `state`: `waiting` or `playing` |
| `settings` | `theme`, `rounds`, `duration` (s), `teams`, `themes` |
| `round` | `drawer`, `word` (drawer only), `hint`, `theme`, `duration`, `remaining` (ms), `round`, `rounds`, `team` (drawing team) |
| `roundEnd` | `word`, `winner`, `points` |
| `gameEnd` | `players` ranking, `teams` scores, then the room is closed |
| `message`, `system`, `error` | `message`, `nickname` |
| `draw`, `clear` | same as client |

A stroke is one drawing operation, coordinates are in `[0, 1]`:

| kind | fields |
|---|---|
| `line` | `from`, `to`, `color`, `size` (brush and eraser) |
| `spray` | `points` (up to 100 dots), `color`, `size` |
| `fill` | `from`, `color` (flood fill) |

`begin: true` marks the first stroke of an action, which is what `undo` removes.

## Interface

- Every page (lobby, room, admin) has a 🌙 / ☀️ button in its header to switch between the light and dark themes.
- The choice is saved in the browser (`localStorage` key `color-scheme`); without one, the system preference is used.
- The canvas stays white in both themes so that drawing colors look the same for everyone.
- Colors are CSS variables in `web/style.css`, redefined under `:root[data-theme="dark"]`; `web/theme.js` sets the theme before the page renders.

## Rules

- The first player in a room is its admin and can start the game (2 to 10 players).
- A game has 3 rounds of 80 seconds by default, the admin can change it before starting; the drawer is drawn at random each round.
- Points: 100 scaled by the remaining time, 10 minimum.
- Team mode: players are split into two teams (each needs 2 players), teams draw in turn and only the drawer's teammates can guess; the guesser's points go to the team.
- Empty rooms are removed after one minute.
- Words live in `WORDS_DIR/<theme>.txt`, one per line, editable from the admin page; the defaults are in `words/`.
- A word is unique across all themes (case and accents ignored).
