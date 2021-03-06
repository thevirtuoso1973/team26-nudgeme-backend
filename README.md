# Team 26's Back-end Go Server

This was initially started as a rewrite of the previous (node.js) back-end since I anticipated
we would need to write some back-end code for some extra features. As expected,
we did, and the back-end now holds the code for three parts:

- wellbeing visualisation, with Google Maps
- data/message passing between mobile users (i.e. for user wellbeing sharing and
sending step goals)
- add friend page, to serve a deeplink

## Deployment Guide & Usage

You need an installation of `go`, using your package manager. You can verify
you have it with `go version`.

Follow these steps on the Linode server:
- Copy the latest code into ct's home directory in 'team26-goviz/'.
- `go build` in that project directory (team26-goviz/)
- Restart the systemd service with `sudo systemctl restart goviz`

Of course, this can be done on any server, but you'd need to set up your own
`goviz.service` file.

You can modify the goviz.service just like any systemd service, e.g. if you want to change
where it searches for the binary.
The service config/unit file is located in `/lib/systemd/system/goviz.service`.

Current domain that links to the server:
`https://comp0016.cyberchris.xyz/`.
This is retrieved from the environment variable in the `goviz.service` file.

## API Docs

See the `WellbeingRecord` struct in `models.go` for the latest. Fields that are marked `omitempty`
are intended to be optional in the future, e.g. weeklySteps (so one doesn't need a pedometer).

### Wellbeing Data & Steps for Map

Endpoint: *.../add-wellbeing-record*

- postCode: string e.g. TW6
- wellbeingScore: integer
- weeklySteps: integer
- errorRate: integer, this is abs(score-userScore), where score is our estimate
of their score
- supportCode: String
- date_sent: string, 'yyyy-MM-dd'

### User Wellbeing Sharing

#### .../user

check if user exists.

Request example:
``` json
{
"identifier": "abc1337",
"password": "battery horse staple"
}
```

Response example:

``` json
{
"success": true,
"exists": true,
}
```

#### .../user/new

add new user

Request example:
``` json
{
"identifier": "abc1337",
"password": "battery horse staple"
}
```

Response example:

``` json
{
"success": true,
}
```

#### .../user/message

get unread 'messages' for user

Request example:
``` json
{
"identifier": "abc1337",
"password": "battery horse staple"
}
```

Response example:

``` json
[{"identifier_from":"blahblah", "data":"123"}, {"identifier_from":"blahblah2", "data":"1234"}, ...]
```

#### .../user/message/new

send message/data to a user

Request example:
``` json
{
"identifier_from": "abc1337",
"password": "battery horse staple",
"identifier_to": "bobby420",
"data": "UXVvdGggdGhlIFJhdmVuIOKAnE5ldmVybW9yZS7igJ0="
}
```

Response example:

``` json
{
"success": true,
}
```

### P2P nudging

Nothing special on the back-end, uses the same structure as wellbeing
sharing. It's up to the client to define the different spec.

Only difference is that it is using a different table.

#### .../user/nudge

See .../user/message section.

#### .../user/nudge/new

See .../user/message/new section.
