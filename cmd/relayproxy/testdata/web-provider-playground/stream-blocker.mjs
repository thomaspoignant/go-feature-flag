// Pass-through proxy in front of the relay-proxy that can cut only the flag-change streams
// (SSE + WebSocket) while /v1/allflags & co keep working, to exercise the polling fallback.
//   node stream-blocker.mjs            -> listens on :1032, forwards to localhost:1031
//   GET /__blocker?block=1|0           -> toggle (also driven by the playground UI)
import http from 'node:http';
import net from 'node:net';

const PORT = Number(process.env.PORT ?? 1032);
const TARGET = new URL(process.env.TARGET ?? 'http://localhost:1031');
const isStream = (url = '') => url.startsWith('/stream/') || url.startsWith('/ws/');

let blocked = false;
const openStreams = new Set();

const cors = { 'Access-Control-Allow-Origin': '*', 'Content-Type': 'application/json' };

const server = http.createServer((req, res) => {
  if (req.url.startsWith('/__blocker')) {
    const q = new URL(req.url, 'http://x').searchParams.get('block');
    if (q !== null) {
      blocked = q === '1';
      if (blocked) for (const s of openStreams) s.destroy();
      console.log(`streams ${blocked ? 'BLOCKED' : 'allowed'}`);
    }
    res.writeHead(200, cors).end(JSON.stringify({ blocked }));
    return;
  }
  if (blocked && isStream(req.url)) {
    res.writeHead(503, { 'Access-Control-Allow-Origin': '*' }).end('stream blocked by stream-blocker');
    return;
  }
  const upstream = http.request(
    { host: TARGET.hostname, port: TARGET.port, path: req.url, method: req.method, headers: req.headers },
    (up) => {
      res.writeHead(up.statusCode ?? 502, up.headers);
      res.flushHeaders(); // SSE: send headers now so EventSource fires onopen
      up.pipe(res);
    },
  );
  upstream.on('error', () => res.writeHead(502).end());
  if (isStream(req.url)) {
    openStreams.add(res.socket);
    res.socket.once('close', () => openStreams.delete(res.socket));
  }
  req.pipe(upstream);
});

// WebSocket upgrade: raw TCP tunnel
server.on('upgrade', (req, socket, head) => {
  if (blocked) return socket.destroy();
  const up = net.connect(Number(TARGET.port), TARGET.hostname, () => {
    up.write(`${req.method} ${req.url} HTTP/1.1\r\n`);
    for (let i = 0; i < req.rawHeaders.length; i += 2) up.write(`${req.rawHeaders[i]}: ${req.rawHeaders[i + 1]}\r\n`);
    up.write('\r\n');
    up.write(head);
    up.pipe(socket).pipe(up);
  });
  const close = () => (socket.destroy(), up.destroy(), openStreams.delete(socket));
  up.on('error', close);
  socket.on('error', close).on('close', close);
  openStreams.add(socket);
});

server.listen(PORT, () => console.log(`stream-blocker on :${PORT} -> ${TARGET.origin}`));
