# sptlrx websocket specification

This is a specification that can be used to embed sptlrx into your applications using websockets.

`sptlrx web` starts a websocket connection. The default port is random, use the `--port` flag to specify it. Also, pass `--no-browser` if you don't want your browser to open at startup.

The endpoint is located in `/ws`.

## Message structure

The messages are encoded in json format and look like this:

```json
{
  "lines": [
    {
      "time": 1000, // line's time in milliseconds
      "words": "Hello world"
    },
    ...
  ],
  "index": 0, // index of the current line
  "playing": true, // whether the track is currently being played
  "error": "" // error, if any
}
```

Note that the message may contain one or more fields. For example, you will get lines only when the track starts, then you will get messages like `{"index": 2}` indicating the beginning of the next line; if the track is paused, you will get `{"playing": false}`; if an error occurred, you will get `{"error": "some text"}` and so on.
