import { OpenFeature, ProviderEvents, type EvaluationContext, type Logger } from '@openfeature/web-sdk';
import { GoFeatureFlagWebProvider, type GoFeatureFlagWebProviderOptions } from '@openfeature/go-feature-flag-web-provider';

type FlagType = 'boolean' | 'string' | 'number' | 'object';
type FlagRow = { key: string; type: FlagType };

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
declare const __PROVIDER_SOURCE__: string;
$('branch').textContent = `(provider: ${__PROVIDER_SOURCE__})`;

const FLAGS_STORAGE = 'goff-playground-flags';
const OPTS_STORAGE = 'goff-playground-opts';

// ---------- log ----------
const logEl = $<HTMLDivElement>('log');
function log(level: 'debug' | 'info' | 'warn' | 'error' | 'event', ...args: unknown[]) {
  const line = document.createElement('div');
  line.className = `l-${level}`;
  line.dataset.level = level;
  const text = args
    .map((a) => {
      if (a instanceof Error) return `${a.name}: ${a.message}`;
      if (typeof a === 'object') {
        try {
          return JSON.stringify(a);
        } catch {
          return String(a);
        }
      }
      return String(a);
    })
    .join(' ');
  line.textContent = `${new Date().toISOString().slice(11, 23)} [${level}] ${text}`;
  if (level === 'debug' && !$<HTMLInputElement>('showDebug').checked) line.style.display = 'none';
  const atBottom = logEl.scrollTop + logEl.clientHeight >= logEl.scrollHeight - 4;
  logEl.appendChild(line);
  if (atBottom) logEl.scrollTop = logEl.scrollHeight;
  (level === 'event' ? console.info : console[level])('[goff-playground]', ...args);
}
$('clearLog').onclick = () => (logEl.innerHTML = '');
$<HTMLInputElement>('showDebug').onchange = (e) => {
  const show = (e.target as HTMLInputElement).checked;
  logEl.querySelectorAll<HTMLElement>('[data-level="debug"]').forEach((el) => (el.style.display = show ? '' : 'none'));
};

const logger: Logger = {
  debug: (...a) => log('debug', ...a),
  info: (...a) => log('info', ...a),
  warn: (...a) => log('warn', ...a),
  error: (...a) => log('error', ...a),
};

// ---------- flags table ----------
let flags: FlagRow[] = JSON.parse(localStorage.getItem(FLAGS_STORAGE) ?? 'null') ?? [
  { key: 'my-bool-flag', type: 'boolean' },
  { key: 'my-string-flag', type: 'string' },
  { key: 'my-number-flag', type: 'number' },
  { key: 'my-object-flag', type: 'object' },
];
const client = OpenFeature.getClient('playground');
const lastValues = new Map<string, string>();

function saveFlags() {
  localStorage.setItem(FLAGS_STORAGE, JSON.stringify(flags));
}

function evaluate(row: FlagRow) {
  switch (row.type) {
    case 'boolean':
      return client.getBooleanDetails(row.key, false);
    case 'string':
      return client.getStringDetails(row.key, '<default>');
    case 'number':
      return client.getNumberDetails(row.key, -1);
    case 'object':
      return client.getObjectDetails(row.key, {});
  }
}

function renderFlags(changed: string[] = []) {
  const tbody = $<HTMLTableSectionElement>('flags');
  tbody.innerHTML = '';
  for (const row of flags) {
    const d = evaluate(row);
    const value = JSON.stringify(d.value);
    const tr = document.createElement('tr');
    const cells = [row.key, row.type, value, d.variant ?? '', d.reason ?? '', d.errorCode ? `${d.errorCode} ${d.errorMessage ?? ''}` : ''];
    for (const c of cells) {
      const td = document.createElement('td');
      td.textContent = c;
      tr.appendChild(td);
    }
    const rm = document.createElement('button');
    rm.textContent = '×';
    rm.onclick = () => {
      flags = flags.filter((f) => f !== row);
      saveFlags();
      renderFlags();
    };
    tr.appendChild(document.createElement('td')).appendChild(rm);
    if (changed.includes(row.key) || (lastValues.has(row.key) && lastValues.get(row.key) !== value)) {
      tr.classList.add('flash');
    }
    lastValues.set(row.key, value);
    tbody.appendChild(tr);
  }
}

$('addFlag').onclick = () => {
  const key = $<HTMLInputElement>('newKey').value.trim();
  if (!key) return;
  flags.push({ key, type: $<HTMLSelectElement>('newType').value as FlagType });
  $<HTMLInputElement>('newKey').value = '';
  saveFlags();
  renderFlags();
};
$('evalAll').onclick = () => renderFlags();

// ---------- status ----------
function setBadge(id: string, text: string) {
  const el = $(id);
  el.textContent = text;
  el.className = 'badge';
  if (/ready|connected/i.test(text)) el.classList.add('ok');
  else if (/stale|reconcil|connecting|not_ready|idle/i.test(text)) el.classList.add('warn');
  else if (/error|fatal|closed/i.test(text)) el.classList.add('err');
}

let provider: GoFeatureFlagWebProvider | undefined;
setInterval(() => {
  setBadge('sdkStatus', client.providerStatus);
  // changeStrategy only exists in providers with SSE support (> 0.2.10).
  const strategy = (provider as unknown as { changeStrategy?: { name: string; status: string } } | undefined)?.changeStrategy;
  setBadge('strategyStatus', provider ? (strategy ? `${strategy.name}: ${strategy.status}` : 'n/a') : '-');
}, 300);

for (const ev of Object.values(ProviderEvents)) {
  OpenFeature.addHandler(ev, (details) => {
    setBadge('lastEvent', ev);
    const { domain: _d, providerName: _p, ...rest } = (details ?? {}) as Record<string, unknown>;
    log('event', `${ev}`, rest);
    renderFlags(ev === ProviderEvents.ConfigurationChanged ? ((details?.flagsChanged as string[]) ?? []) : []);
  });
}

// ---------- options ----------
const form = $<HTMLFormElement>('opts');
const savedOpts = JSON.parse(localStorage.getItem(OPTS_STORAGE) ?? '{}') as Record<string, string | boolean>;
for (const [k, v] of Object.entries(savedOpts)) {
  const input = form.elements.namedItem(k) as HTMLInputElement | null;
  if (!input) continue;
  if (input.type === 'checkbox') input.checked = Boolean(v);
  else input.value = String(v);
}

function readOptions(): GoFeatureFlagWebProviderOptions {
  const fd = new FormData(form);
  const raw: Record<string, string | boolean> = {};
  for (const el of Array.from(form.elements) as HTMLInputElement[]) {
    if (!el.name) continue;
    raw[el.name] = el.type === 'checkbox' ? el.checked : el.value;
  }
  localStorage.setItem(OPTS_STORAGE, JSON.stringify(raw));
  const num = (k: string) => (fd.get(k) === '' ? undefined : Number(fd.get(k)));
  return {
    endpoint: String(fd.get('endpoint')),
    mode: fd.get('mode') as 'ws' | 'sse',
    apiKey: String(fd.get('apiKey') || '') || undefined,
    pollingIntervalMs: num('pollingIntervalMs'),
    maxRetries: num('maxRetries'),
    retryInitialDelay: num('retryInitialDelay'),
    retryDelayMultiplier: num('retryDelayMultiplier'),
    apiTimeout: num('apiTimeout'),
    dataFlushInterval: num('dataFlushInterval'),
    disableDataCollection: raw.disableDataCollection === true,
    // cast: `mode` / `pollingIntervalMs` are unknown to providers released before SSE support.
  } as GoFeatureFlagWebProviderOptions;
}

function readContext(): EvaluationContext {
  const custom = JSON.parse($<HTMLTextAreaElement>('ctx').value || '{}');
  return { targetingKey: $<HTMLInputElement>('targetingKey').value, ...custom };
}

// ---------- actions ----------
$('connect').onclick = async () => {
  const opts = readOptions();
  log('info', 'Creating provider with options', opts);
  ($('connect') as HTMLButtonElement).disabled = true;
  ($('disconnect') as HTMLButtonElement).disabled = false;
  try {
    await OpenFeature.setContext(readContext());
    provider = new GoFeatureFlagWebProvider(opts, logger);
    (window as unknown as { provider: unknown }).provider = provider;
    await OpenFeature.setProviderAndWait(provider);
    log('info', 'setProviderAndWait() resolved');
  } catch (e) {
    log('error', 'setProviderAndWait() rejected', e);
  }
  renderFlags();
};

$('disconnect').onclick = async () => {
  await OpenFeature.clearProviders();
  log('info', 'clearProviders() done (provider.onClose called)');
  provider = undefined;
  ($('connect') as HTMLButtonElement).disabled = false;
  ($('disconnect') as HTMLButtonElement).disabled = true;
  renderFlags();
};

$('setCtx').onclick = async () => {
  try {
    const ctx = readContext();
    log('info', 'setContext()', ctx);
    await OpenFeature.setContext(ctx);
    log('info', 'setContext() resolved');
  } catch (e) {
    log('error', 'setContext() failed', e);
  }
  renderFlags();
};

// ---------- stream-blocker (polling fallback) ----------
const BLOCKER_URL = 'http://localhost:1032/__blocker';
async function blocker(block?: boolean) {
  try {
    const res = await fetch(block === undefined ? BLOCKER_URL : `${BLOCKER_URL}?block=${block ? 1 : 0}`);
    const { blocked } = (await res.json()) as { blocked: boolean };
    setBadge('blockerStatus', blocked ? 'blocked (error)' : 'allowed (connected)');
    if (block !== undefined) log('info', `stream-blocker: streams ${blocked ? 'BLOCKED' : 'allowed'}`);
  } catch {
    setBadge('blockerStatus', 'blocker not running');
  }
}
$('blockStreams').onclick = () => blocker(true);
$('allowStreams').onclick = () => blocker(false);
blocker();

(window as unknown as { OpenFeature: unknown }).OpenFeature = OpenFeature;
renderFlags();
