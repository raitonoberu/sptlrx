<div align="center">

<h1><a href="https://github.com/raitonoberu/sptlrx">sptlrx</a></h1>
<h4>Timesynced lyrics in your terminal</h4>

![Crystal Castles - Not In Love](./demo.svg "Crystal Castles - Not In Love")

</div>

## Features

- Compatible with Spotify, MPD, Mopidy and MPRIS.
- Works well with long lines & Unicode characters.
- Easy to use customization.
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
nix-env -iA sptlrx
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
# Interval of the internal timer. Determines how often the terminal will be updated.
timerInterval: 200
# Interval for checking the position. Doesn't really affect the precision.
updateInterval: 3000

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
    undeline: false
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
    undeline: false
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
    undeline: false
    strikethrough: false
    blink: false
    faint: true

### Pipe settings ###
pipe:
  # Maximum line length. 0 - unlimited.
  length: 0
  # How to handle overflowing strings. Possible values: word, none, ellipsis.
  overflow: word
  # Whether to ignore errors instead of printing to stdout.
  ignoreErrors: true

### MPD settings ###
mpd:
  # MPD server address with port
  address: 127.0.0.1:6600
  # MPD server password (if any)
  password: ""

### Mopidy settings ###
mopidy:
  # Mopidy server address with port
  address: 127.0.0.1:6680
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

You can also set the `SPOTIFY_COOKIE` enviroment variable or pass the `--cookie` flag.

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
```

Linux only. System player that supports MPRIS protocol will be used.

## Information

### Piping

Run `sptlrx pipe` to start printing the current lines to stdout. This can be used in various status bars and other applications.

## License

**MIT License**, see [LICENSE](./LICENSE) for additional information.
