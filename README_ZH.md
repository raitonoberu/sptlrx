<div align="center">

<h1><a href="https://github.com/raitonoberu/sptlrx">sptlrx</a></h1>
<h4>在终端中显示同步歌词</h4>

<a href="https://www.youtube.com/watch?v=qR2QIJdtgiU">
  <img title="demo" src="./demo.gif" width="450"/>
</a>

</div>

## 功能

- 支持 Spotify、MPD、Mopidy、MPRIS 与浏览器扩展。
- 适配长行与 Unicode 字符。
- 易于自定义样式。
- 可将歌词通过管道输出到 stdout。
- 单一二进制，跨平台。

## 安装

Linux、Windows、macOS 等平台可从 [Releases](https://github.com/raitonoberu/sptlrx/releases/latest) 下载，或参见 [building.md](./building.md) 自行编译。

## 配置

首次运行会自动在配置目录创建 `config.yaml`。在 Linux 上路径通常为 `~/.config/sptlrx/config.yaml`。运行 `sptlrx -h` 可查看完整路径。

示例配置（节选）：

```yaml
### 全局设置 ###
# 使用的播放器。可选：spotify, mpd, mopidy, mpris, browser
# 也可以使用英文逗号分隔按顺序尝试多个播放器，例如：
#   player: mopidy ,mpris
# 程序会按顺序初始化，直到有一个成功为止；未列出的播放器不会被尝试。
player: spotify
# 是否忽略错误（为 true 时不在界面显示错误）。
ignoreErrors: true
# 内部计时器间隔（毫秒），决定终端刷新频率。
timerInterval: 200
# 位置更新间隔（毫秒），不显著影响精度。
updateInterval: 2000

### 样式设置 ###
style:
  # 行的水平对齐方式：left, center, right
  hAlignment: center
  before:
    background: ""
    foreground: ""
    bold: true
    italic: false
    underline: false
    strikethrough: false
    blink: false
    faint: false
  current:
    background: ""
    foreground: ""
    bold: true
    italic: false
    underline: false
    strikethrough: false
    blink: false
    faint: false
  after:
    background: ""
    foreground: ""
    bold: false
    italic: false
    underline: false
    strikethrough: false
    blink: false
    faint: true

### MPD 设置 ###
mpd:
  address: 127.0.0.1:6600
  password: ""

### Mopidy 设置 ###
mopidy:
  address: 127.0.0.1:6680

### MPRIS 设置 ###
mpris:
  # MPRIS 播放器白名单；为空则自动选第一个可用的。
  players: []

### 浏览器扩展 ###
browser:
  port: 8974

### 本地歌词 ###
local:
  # 扫描 .lrc 文件的目录，例如："~/Music"。设置后将仅使用本地歌词源。
  folder: ""
```

### 多播放器顺序（回退）示例

```yaml
# config.yaml
player: mopidy ,mpris
mopidy:
  address: 127.0.0.1:6680
mpris:
  players: []
```

程序会按照 `player` 字段给定的顺序依次尝试初始化播放器，使用第一个初始化成功的播放器。仅尝试用户列出的播放器。

### Spotify 登录

将 `player` 设为 `spotify` 后，首次使用需要登录：执行 `sptlrx login`，根据提示提供 `Client ID`/`Client Secret`，或通过环境变量 `SPOTIFY_CLIENT_ID`/`SPOTIFY_CLIENT_SECRET` 或命令行参数 `--client-id`/`--client-secret` 提供。登录成功后凭据将保存到 `$XDG_STATE_HOME/sptlrx/spotify-auth.json`。

### 输出到管道

执行 `sptlrx pipe` 可将当前行输出到标准输出，便于与状态栏等其他程序集成。

## 许可证

**MIT License**，详见 [LICENSE](./LICENSE)。
