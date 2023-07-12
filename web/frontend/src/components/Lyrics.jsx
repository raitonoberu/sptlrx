import { onMount, createSignal, createEffect, batch, For } from "solid-js";
import styles from "./Lyrics.module.css";

// TODO: use window.location.host
const URL = "ws://localhost:8080/ws";

function Lyrics() {
  const [lines, setLines] = createSignal([]);
  const [index, setIndex] = createSignal(0);
  const [playing, setPlaying] = createSignal(false);
  const [error, setError] = createSignal("");

  const onMessage = (event) => {
    const data = JSON.parse(event.data);
    batch(() => {
      if (data.lines !== undefined) setLines(data.lines);
      if (data.index !== undefined) setIndex(data.index);
      if (data.playing !== undefined) setPlaying(data.playing);
      if (data.err !== undefined) setError(data.err);
    });
  };

  let list;
  const scrollTo = (index) => {
    if (list.childElementCount !== 0)
      list.childNodes[index].scrollIntoView({
        behavior: "smooth",
        block: "center",
        inline: "center",
      });
  };

  onMount(() => {
    // TODO: reopen connection?
    let socket = new WebSocket(URL);
    socket.onopen = () => console.log("[open] Connection opened successfully");
    socket.onerror = () => console.log(`[error] An error has occured`);
    socket.onclose = (event) => {
      if (event.wasClean)
        console.log(
          `[close] Connection closed, code=${event.code} reason=${event.reason}`
        );
      else console.log("[close] Connection closed unexpectedly");
    };
    socket.onmessage = onMessage;

    createEffect(() => scrollTo(index()));
  });

  return (
    <div ref={list} class={styles.Lyrics}>
      <For each={lines()}>
        {(line, i) => {
          return (
            <div
              classList={{
                [styles.Line]: true,
                [styles.After]: i() > index(),
              }}
            >
              {line.words}
            </div>
          );
        }}
      </For>
    </div>
  );
}

export default Lyrics;
