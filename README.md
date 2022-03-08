<div align="center">

<h1><a href="https://github.com/raitonoberu/sptlrx">sptlrx</a></h1>
<h4>Spotify lyrics in your terminal.</h4>

![Crystal Castles - Not In Love](./demo.svg "Crystal Castles - Not In Love")

</div>

## Features

- Timesynced lyrics in your terminal.
- Fully compatible with [spotifyd](https://github.com/Spotifyd/spotifyd).
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
````

**Windows**
````
iwr instl.sh/raitonoberu/sptlrx/windows | iex  
````

**macOS**
````
curl -sSL instl.sh/raitonoberu/sptlrx/macos | bash   
````

You can also download the binary from the [Releases](https://github.com/raitonoberu/sptlrx/releases/latest) page or [build it yourself](./building.md).

## Configuration

Since Spotify requires a special web token to display song lyrics, you need to specify your cookie when you first launch.

1. Open your browser.
2. Press F12, open the `Network` tab and go to [open.spotify.com](https://open.spotify.com/).
3. Click on the first request to `open.spotify.com`.
4. Scroll down to the `Request Headers`, right click the `cookie` field and select `Copy value`.
5. Paste it when you are asked.

You can also set the `SPOTIFY_COOKIE` enviroment variable or pass the `--cookie` flag, and your cookie will be saved on the next run. You can always clear cookie by running `sptlrx clear`.

## Information

### Styling

There are three special flags for applying custom colors and styles to lines: `--current`, `--before` and `--after`. The syntax for all flags is the same - pass styles and colors separated by commas. Example:
```sh
sptlrx --current "bold,#FFDFD3,#957DAD" --before "104,faint,italic" --after "104,faint"
```
List of allowed styles: `bold`, `italic`, `underline`, `strikethrough`, `blink`, `faint`. The colors can be either in HEX format, or ANSI 0-255. The first color represents the foreground, the second represents the background. **Note that styles will not work if your terminal does not support them.**

### Piping

Run `sptlrx pipe` to start printing the current lines in stdout. This can be used in various status bars and other applications. You can specify the maximum line length and overflow (run `sptlrx pipe -h` for help).

## License

**MIT License**, see [LICENSE](./LICENSE) for additional information.
