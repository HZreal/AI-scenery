export function parseSSEEvent(event) {
  let type = "";
  const data = [];
  for (const line of event.replace(/\r\n/g, "\n").split("\n")) {
    if (line.startsWith("event:")) type = line.slice(6).trim();
    if (line.startsWith("data:")) data.push(line.slice(5).trimStart());
  }
  return { type, data: data.join("\n") };
}
