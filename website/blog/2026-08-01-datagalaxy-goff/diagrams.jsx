import React from 'react';

// Diagrams for the DataGalaxy case study.
//
// Built to match the site's existing card idiom (see
// src/components/comparison/ToolMeta.js): rounded/bordered containers,
// brand green (#18b192) + sage (titles-500) accents, FontAwesome icons, and
// full dark-mode support driven by Tailwind `dark:` classes. Class strings are
// written out in full (no dynamic concatenation) so Tailwind's JIT keeps them.

// ---------------------------------------------------------------------------
// Shared primitives
// ---------------------------------------------------------------------------

// Tone presets, mirroring ToolMeta's TONES so the diagrams read as part of the
// same visual family.
const TONES = {
  good: {
    bar: 'bg-[#18b192]',
    icon: 'text-[#0c8f77] dark:text-[#3ccbaa]',
    ring: 'border-[#18b192]/40 dark:border-[#3ccbaa]/40',
  },
  warn: {
    bar: 'bg-amber-400',
    icon: 'text-amber-600 dark:text-amber-400',
    ring: 'border-amber-400/40',
  },
  info: {
    bar: 'bg-slate-400',
    icon: 'text-slate-500 dark:text-slate-300',
    ring: 'border-[#273437]/10 dark:border-white/10',
  },
  bad: {
    bar: 'bg-red-400',
    icon: 'text-red-500 dark:text-red-400',
    ring: 'border-red-400/40',
  },
};

// A single labelled node card.
function Node({icon, title, sub, tone = 'info', className = ''}) {
  const t = TONES[tone] || TONES.info;
  return (
    <div
      className={
        'flex items-start gap-3 px-4 py-3 rounded-xl bg-white dark:bg-[#2a2a2a] border border-solid ' +
        t.ring +
        ' ' +
        className
      }>
      <span className={'mt-0.5 shrink-0 text-lg ' + t.icon} aria-hidden>
        <i className={'fa-solid ' + icon} />
      </span>
      <span className="flex flex-col leading-tight">
        <span className="font-bold text-[0.9rem] text-gray-900 dark:text-gray-50">
          {title}
        </span>
        {sub && (
          <span className="text-[0.78rem] text-gray-500 dark:text-gray-400 mt-0.5">
            {sub}
          </span>
        )}
      </span>
    </div>
  );
}

// Small uppercase caption used above each diagram body.
function Caption({children}) {
  return (
    <div className="text-xs font-bold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-3 text-center">
      {children}
    </div>
  );
}

// Gradient tint presets for the outer Frame, keyed by the `tint` prop.
const FRAME_TINTS = {
  good: 'from-[#18b192]/[0.08]',
  warn: 'from-amber-400/[0.08]',
};

// Outer card wrapper shared by all three diagrams.
function Frame({tint = 'titles', children}) {
  const from = FRAME_TINTS[tint] || 'from-titles-500/[0.07]';
  return (
    <div
      className={
        'not-prose my-8 p-5 md:p-6 rounded-2xl border border-solid border-[#273437]/10 dark:border-white/10 bg-gradient-to-b ' +
        from +
        ' to-transparent'
      }>
      {children}
    </div>
  );
}

// Downward flow arrow (vertical stacks).
function DownArrow() {
  return (
    <i
      className="fa-solid fa-arrow-down-long text-titles-500 text-lg my-2"
      aria-hidden
    />
  );
}

// ---------------------------------------------------------------------------
// 1. Before — four disconnected flag mechanisms
// ---------------------------------------------------------------------------
export function BeforeDiagram() {
  return (
    <Frame tint="warn">
      <Caption>Before · four systems that never talked to each other</Caption>
      <div className="grid gap-3 sm:grid-cols-2">
        <Node
          tone="warn"
          icon="fa-database"
          title="SQL table rows"
          sub="Read at request time by the backend"
        />
        <Node
          tone="warn"
          icon="fa-server"
          title="Backend env vars"
          sub="Baked into deploys — toggling means a redeploy"
        />
        <Node
          tone="warn"
          icon="fa-window-maximize"
          title="Frontend build-time vars"
          sub="Separate set, no link to the backend flag"
        />
        <Node
          tone="warn"
          icon="fa-network-wired"
          title="Infra / ingress toggles"
          sub="Feature gates & canary weights, unrelated again"
        />
      </div>
      <div className="mt-4 flex justify-center">
        <span className="inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm font-semibold bg-red-400/10 text-red-600 dark:text-red-400 border border-solid border-red-400/30">
          <i className="fa-solid fa-circle-question" aria-hidden />
          “Is this feature actually live for this customer?” — no single source
          of truth
        </span>
      </div>
    </Frame>
  );
}

// ---------------------------------------------------------------------------
// 2. After — one evaluation, everywhere, through the RelayProxy
// ---------------------------------------------------------------------------
export function AfterDiagram() {
  return (
    <Frame tint="good">
      <Caption>After · one evaluation, everywhere</Caption>
      <div className="flex flex-col items-center text-center">
        {/* Three consumers, one contract */}
        <div className="grid grid-cols-3 gap-2 w-full">
          <Node tone="info" icon="fa-server" title=".NET backend" className="justify-center" />
          <Node
            tone="info"
            icon="fa-window-maximize"
            title="Angular frontend"
            className="justify-center"
          />
          <Node
            tone="info"
            icon="fa-network-wired"
            title="Infra check"
            className="justify-center"
          />
        </div>
        <DownArrow />
        <div className="w-full max-w-md px-4 py-3 rounded-xl bg-titles-500/90 text-[#273437] font-bold shadow-sm">
          <i className="fa-solid fa-plug mr-2" aria-hidden />
          OpenFeature API
          <span className="block text-xs font-semibold opacity-80">
            one contract — write against it once
          </span>
        </div>
        <DownArrow />
        <div className="w-full max-w-md px-4 py-3 rounded-xl bg-[#18b192] text-white font-bold shadow-sm">
          <i className="fa-solid fa-bolt mr-2" aria-hidden />
          GO Feature Flag RelayProxy
          <span className="block text-xs font-semibold opacity-90">
            single HTTP service — same rules, rollout & kill switch
          </span>
        </div>
        <DownArrow />
        <Node
          tone="good"
          icon="fa-flag"
          title="Flag configuration"
          sub="Evaluated once, consistently, for every consumer"
          className="max-w-md w-full justify-center"
        />
      </div>
    </Frame>
  );
}
