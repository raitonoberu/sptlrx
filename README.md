<div align="center">

<h1><a href="https://github.com/raitonoberu/sptlrx">sptlrx</a></h1>
<h4>Synchronized lyrics in your terminal</h4>

<a href="https://www.youtube.com/watch?v=qR2QIJdtgiU">
  <img title="Crystal Castles — Kerosene" src="./demo.gif" width="450"/>
</a>

</div>

## Features

- Compatible with Spotify, MPD, Mopidy, MPRIS and browsers.
- Works well with long lines & Unicode characters.
- Easy to customize.
- Allows piping to stdout.
- Single binary & cross-plaftorm.

## Installation

**Linux**

- Arch Linux ([@BachoSeven](https://github.com/BachoSeven))

```
yay -S sptlrx-bin
```

- NixOS ([@MoritzBoehme](https://github.com/MoritzBoehme))

```
nix-env -iA nixos.sptlrx
```

or if using nixpkgs

```
nix-env -iA nixpkgs.sptlrx
```

- Other

```
curl -sSL instl.sh/raitonoberu/sptlrx/linux | bash
```

**Windows**

```
iwr instl.sh/raitonoberu/sptlrx/windows | iex
```

**macOS**

```
curl -sSL instl.sh/raitonoberu/sptlrx/macos | bash
```

You can also download the binary from the [Releases](https://github.com/raitonoberu/sptlrx/releases/latest) page or [build it yourself](./building.md).

## Configuration

Config file will be created at the first launch. On Linux it's located in `~/.config/sptlrx/config.yaml`. Run `sptlrx -h` to see the full path.

<details>
<summary>Show config contents (with descriptions)</summary>

```yaml
### Global settings ###
# Your Spotify cookie. Only needed if you are going to use Spotify as a player.
cookie: ""
# Player that will be used. Possible values: spotify, mpd, mopidy, mpris.
player: spotify
# Host of lyrics API to be used in case the cookie is not provided.
host: lyricsapi.vercel.app
# Whether to ignore errors instead of showing them.
ignoreErrors: true
# Interval of the internal timer. Determines how often the terminal will be updated.
timerInterval: 200
# Interval for checking the position. Doesn't really affect the precision.
updateInterval: 2000

### Style settings ###
style:
  # Horizontal alignment of lines. Possible values: left, center, right.
  hAlignment: center
  # Style of the lines before the current one.
  before:
    # The colors can be either in HEX format, or ANSI 0-255.
    background: ""
    foreground: ""
    bold: true
    italic: false
    underline: false
    strikethrough: false
    blink: false
    faint: false
  # Style of the current line.
  current:
    # The colors can be either in HEX format, or ANSI 0-255.
    background: ""
    foreground: ""
    bold: true
    italic: false
    underline: false
    strikethrough: false
    blink: false
    faint: false
  # Style of the lines after the current one.
  after:
    # The colors can be either in HEX format, or ANSI 0-255.
    background: ""
    foreground: ""
    bold: false
    italic: false
    underline: false
    strikethrough: false
    blink: false
    faint: true

### Pipe settings ###
pipe:
  # Maximum line length. 0 - unlimited.
  length: 0
  # How to handle overflowing strings. Possible values: word, none, ellipsis.
  overflow: word

### MPD settings ###
mpd:
  # MPD server address with port.
  address: 127.0.0.1:6600
  # MPD server password (if any).
  password: ""

### Mopidy settings ###
mopidy:
  # Mopidy server address with port.
  address: 127.0.0.1:6680

### MPRIS settings ###
mpris:
  # Whitelist of MPRIS players. First available is used if empty.
  players: []

### Browser extension settings ###
browser:
  # Port on which the server will be started.
  port: 8974

### Local lyrics source ###
local:
  # Enable the local lyrics source.
  # For backwards compatibility reasons setting the folder also enables this source.
  enabled: false
  # Folder for scanning .lrc files. Example: "~/Music".
  folder: ""
```

</details>

### Spotify

```yaml
# config.yaml
cookie: <your cookie>
player: spotify
```

If you want to use Spotify as your player or lyrics source, you need to specify your cookie.

1. Open your browser.
2. Press F12, open the `Network` tab and go to [open.spotify.com](https://open.spotify.com/).
3. Click on the first request to `open.spotify.com`.
4. Scroll down to the `Request Headers`, right click the `cookie` field and select `Copy value`.
5. Paste it to your config.

You can also set the `SPOTIFY_COOKIE` environment variable or pass the `--cookie` flag.

**TREAT YOUR COOKIE LIKE A PASSWORD AND NEVER SHARE IT**

### MPD

```yaml
# config.yaml
player: mpd
mpd:
  address: 127.0.0.1:6600
  password: ""
```

MPD server will be used as a player.

### Mopidy

```yaml
# config.yaml
player: mopidy
mopidy:
  address: 127.0.0.1:6680
```

Mopidy server will be used as a player.

### MPRIS

```yaml
# config.yaml
player: mpris
mpris:
  players: []
```

Linux only. System player that supports MPRIS protocol will be used. You can also specify a whitelist of players to use, example: `players: [rhythmbox, spotifyd, ncspot]`. Run `playerctl -l` to get the names.

### Browser

```yaml
# config.yaml
player: browser
browser:
  port: 8974
```

You need to install a [browser extension](https://wnp.keifufu.dev/extension/getting-started). If you don't change the default port, no further configuration is required. Otherwise, create a custom adapter in the extension settings. **You can only run one instance on one port.**

### Local

```yaml
# config.yaml
local:
  enabled: true
  folder: ""
```

Display lyrics from local `.lrc` files.

By default, the application will look for a file that is a sibling of a local music file (e.g. local player via mpris), i.e. with the same path, with the extension replaced by `.lrc`.

If the `folder` config option is set, it will additionally search for files within that folder. If the player provides a relative path to the music file (e.g. mpd), an exact match is attempted first as described above. If that fails, a best-effort search will be performed, returning a `.lrc` file in the folder (can be nested) with the most similar name.

All other lyrics sources will be disabled.

## Information

### Source

If you specify your Spotify cookie, the lyrics will be fetched using your account. Otherwise, the API [hosted by me](https://github.com/raitonoberu/lyricsapi) will be used. It is also possible to host your own API or use local `.lrc` files.

### Piping

Run `sptlrx pipe` to start printing the current lines to stdout. This can be used in various status bars and other applications.

### Flags

You can pass flags to override the style parameters defined in the config. Example:

```sh
sptlrx --current "bold,#FFDFD3,#957DAD" --before "104,faint,italic" --after "104,faint"
```

List of allowed styles: `bold`, `italic`, `underline`, `strikethrough`, `blink`, `faint`. The colors can be either in HEX format, or ANSI 0-255. The first color represents the foreground, the second represents the background.

Run `sptlrx --help` to see all the flags.

## License

**MIT License**, see [LICENSE](./LICENSE) for additional information.
