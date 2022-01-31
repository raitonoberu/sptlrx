<div align="center">

<h1><a href="https://github.com/raitonoberu/sptlrx">sptlrx</a></h1>
<h4>Spotify lyrics in your terminal.</h4>

<br/>

<img src="./demo.svg"/>

</div>
<br/>

## Features

- Timesynced lyrics in your terminal.
- Fully compatible with [spotifyd](https://github.com/Spotifyd/spotifyd).
- Works well with long lines & Unicode characters.
- Single binary & cross-plaftorm.

## Installation

#### Ubuntu / Debian

```sh
sudo apt install sptlrx
```

#### Arch (AUR)

```sh
yay -S sptlrx
```

### Other

Download the binary from the [Releases](https://github.com/raitonoberu/sptlrx/releases) page.

## Configuration

Since Spotify requires a special web token to display song lyrics, you need to specify your cookie when you first launch.

1. Open your browser.

2. Press F12, open the `Network` tab and go to [open.spotify.com](https://open.spotify.com/).

3. Click on the first request to `open.spotify.com`.

4. Scroll down to the `Request Headers`, right click the `cookie` field and select `Copy value`.

5. Paste it when you are asked.

Another way to set cookie is to set the `SPOTIFY_COOKIE` enviroment variable. You can always clear cookie by running `sptlrx --clear`.

## Information

### In development

`sptlrx` is pretty much ready to use, however you may encounter some bugs. Please open a new issue so that we can fix it quickly. Also, I plan to add additional settings, such as PC-like mode, color text and more. Stay tuned 😉

### Delay

For some reason unknown to me, there is a delay in the lyrics on some devices. You can manually adjust the delay by using the "**+**" and "**-**" symbols on the keyboard (adds or subtracts 100 ms).

## License

**MIT License**, see [LICENSE](./LICENSE) for additional information.
