sptlrx 5
========

## NAME

**sptlrx** - Synchronized lyrics in your terminal

## LOCATION

The config file will be created at the first launch. It is located in `~/.config/sptlrx/config.yaml`. Run sptlrx -h to see the full path.

## SPOTIFY

### FORMAT
```
# config.yaml
player: spotify
```

### NOTES

If you want to use Spotify as your player, you will need to log in first.

1. Go to [developer.spotify.com](https://developer.spotify.com/dashboard), create a new app, and set the redirect URI to `http://127.0.0.1:8888/callback`. Grab your Client ID and Client Secret.
2. Run `sptlrx login`. You can pass Client ID and Client Secret in one of three ways:
  - As environmental variables: `SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET`
  - As CLI parameters: `--client-id` and `--client-secret`
  - Interactively: run `sptlrx login` without providing credentials and you will be prompted to enter them
3. Spotify login page will open. Log in and wait for the success message.

You only need to do this once. Your credentials will then be saved to `$XDG_STATE_HOME/sptlrx/spotify-auth.json`.

## MPD

### FORMAT

```
# config.yaml
player: mpd
mpd:
  address: 127.0.0.1:6600
  password: ""
```

## MOPIDY

### FORMAT

```
# config.yaml
player: mopidy
mopidy:
  address: 127.0.0.1:6680
```

## MPRIS

### FORMAT

```
# config.yaml
player: mpris
mpris:
  players: []
```

### NOTES

System player that supports MPRIS protocol will be used. You can also specify a whitelist of players to use, example: `players: [rhythmbox, spotifyd, ncspot]`. Run `playerctl -l` to get the names.

## BROWSER

### FORMAT

```
# config.yaml
player: browser
browser:
  port: 8974
```

### NOTES

You need to install a [browser extension](https://wnp.keifufu.dev/extension/getting-started). If you don't change the default port, no further configuration is required. Otherwise, create a custom adapter in the extension settings. **You can only run one instance on one port.**

## LOCAL

### FORMAT

```
# config.yaml
local:
  folder: ""
```

### NOTES

If you want to use your local collection of `.lrc` files to display lyrics, specify the folder to scan. The application will use files with the most similar name. All other lyrics sources will be disabled.
