import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';

const TONE_VALUES = ['good', 'warn', 'info', 'bad'];

// Tone presets. Full class strings (no dynamic concatenation) so Tailwind's
// JIT picks them up.
const TONES = {
  good: {
    bar: 'bg-[#18b192]',
    chip: 'bg-[#18b192]/15 text-[#0c8f77] dark:text-[#3ccbaa]',
    icon: 'text-[#0c8f77] dark:text-[#3ccbaa]',
  },
  warn: {
    bar: 'bg-amber-400',
    chip: 'bg-amber-400/20 text-amber-700 dark:text-amber-400',
    icon: 'text-amber-600 dark:text-amber-400',
  },
  info: {
    bar: 'bg-slate-400',
    chip: 'bg-slate-400/15 text-slate-600 dark:text-slate-300',
    icon: 'text-slate-500 dark:text-slate-300',
  },
  bad: {
    bar: 'bg-red-400',
    chip: 'bg-red-400/15 text-red-700 dark:text-red-400',
    icon: 'text-red-600 dark:text-red-400',
  },
};

// Small pill that states a tool's OpenFeature posture at a glance.
export function Posture({label, tone = 'good'}) {
  const t = TONES[tone] || TONES.info;
  return (
    <span
      className={clsx(
        'inline-flex items-center gap-1.5 self-start px-2.5 py-1 rounded-full text-xs font-bold',
        t.chip
      )}>
      <i className="fa-solid fa-plug" aria-hidden />
      {label}
    </span>
  );
}

// Big-logo header card that opens each tool section.
export function ToolHeader({
  logo,
  alt,
  posture,
  postureTone = 'good',
  children,
}) {
  return (
    <div className="not-prose flex items-center gap-4 my-5 p-4 md:p-5 rounded-2xl border border-solid border-[#273437]/10 dark:border-white/10 bg-gray-50 dark:bg-white/[0.04]">
      <img
        src={logo}
        alt={alt ? `${alt} logo` : ''}
        loading="lazy"
        className="w-14 h-14 md:w-16 md:h-16 shrink-0 rounded-xl object-contain"
      />
      <div className="flex flex-col gap-2">
        {posture && <Posture label={posture} tone={postureTone} />}
        <p className="m-0 text-[1.02rem] leading-snug font-semibold text-gray-800 dark:text-gray-100">
          {children}
        </p>
      </div>
    </div>
  );
}

// Color-coded callout row used for Best-for / Watch-outs / License lines.
export function Field({label, tone = 'info', icon, children}) {
  const t = TONES[tone] || TONES.info;
  return (
    <div className="not-prose flex items-stretch gap-3 my-2.5">
      <span className={clsx('w-1 shrink-0 rounded-full', t.bar)} aria-hidden />
      <span className={clsx('mt-0.5 shrink-0 text-base', t.icon)} aria-hidden>
        <i className={clsx('fa-solid', icon)} />
      </span>
      <p className="m-0 text-[0.95rem] leading-relaxed text-gray-700 dark:text-gray-200">
        <strong className="text-gray-900 dark:text-gray-50">{label}:</strong>{' '}
        {children}
      </p>
    </div>
  );
}

// Selection-criteria grid: scannable icon cards instead of a bullet list.
const CRITERIA = [
  {
    icon: 'fa-plug',
    title: 'OpenFeature support',
    text: 'Native/ecosystem, official provider, community provider, or none.',
  },
  {
    icon: 'fa-lock-open',
    title: 'License & openness',
    text: 'Fully open vs. open-core with features behind a paid tier.',
  },
  {
    icon: 'fa-database',
    title: 'Self-host footprint',
    text: 'Does running it require a database — and how heavy?',
  },
  {
    icon: 'fa-bullseye',
    title: 'Targeting & rollouts',
    text: 'Segments, percentage/progressive, and scheduled changes.',
  },
  {
    icon: 'fa-flask',
    title: 'Experimentation',
    text: 'Built-in A/B with stats, or bring-your-own analytics?',
  },
  {
    icon: 'fa-window-maximize',
    title: 'Built-in UI',
    text: 'Is there a management dashboard — and is it the whole story?',
  },
  {
    icon: 'fa-people-group',
    title: 'Community health',
    text: 'Ecosystem momentum, maintenance, and adoption.',
  },
];

export function Criteria() {
  return (
    <div className="not-prose my-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {CRITERIA.map(c => (
        <div
          key={c.title}
          className="flex items-start gap-3 p-4 rounded-xl border border-solid border-[#273437]/10 dark:border-white/10 bg-gray-50 dark:bg-white/[0.04]">
          <span className="flex items-center justify-center w-9 h-9 shrink-0 rounded-lg bg-titles-500/20 text-titles-500">
            <i className={clsx('fa-solid', c.icon)} aria-hidden />
          </span>
          <div>
            <div className="font-bold text-sm text-gray-900 dark:text-gray-50">
              {c.title}
            </div>
            <div className="text-[0.82rem] leading-snug text-gray-500 dark:text-gray-400 mt-0.5">
              {c.text}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

// Concept diagram for the OpenFeature thesis: one SDK, swappable backends.
const BACKENDS = [
  'GO Feature Flag',
  'flagd',
  'Unleash',
  'Flagsmith',
  'GrowthBook',
  'Flipt',
];

function Arrow() {
  return (
    <i
      className="fa-solid fa-arrow-down text-titles-500 text-lg my-1"
      aria-hidden
    />
  );
}

export function OpenFeatureFlow() {
  return (
    <div className="not-prose my-8 p-5 md:p-6 rounded-2xl border border-solid border-[#273437]/10 dark:border-white/10 bg-gradient-to-b from-titles-500/[0.07] to-transparent">
      <div className="flex flex-col items-center text-center">
        <div className="w-full max-w-sm px-4 py-3 rounded-xl bg-white dark:bg-[#2a2a2a] border border-solid border-[#273437]/10 dark:border-white/10 font-semibold text-gray-800 dark:text-gray-100">
          <i className="fa-solid fa-code mr-2 text-gray-400" aria-hidden />
          Your application code
        </div>
        <Arrow />
        <div className="w-full max-w-sm px-4 py-3 rounded-xl bg-titles-500/90 text-[#273437] font-bold shadow-sm">
          <i className="fa-solid fa-plug mr-2" aria-hidden />
          OpenFeature SDK
          <span className="block text-xs font-semibold opacity-80">
            write against this once
          </span>
        </div>
        <Arrow />
        <div className="w-full">
          <div className="text-xs font-bold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-2">
            Swap the backend — no app code changes
          </div>
          <div className="flex flex-wrap justify-center gap-2">
            {BACKENDS.map(b => (
              <span
                key={b}
                className="px-3 py-1.5 rounded-lg text-sm font-semibold bg-white dark:bg-[#2a2a2a] border border-solid border-[#273437]/10 dark:border-white/10 text-gray-700 dark:text-gray-200">
                {b}
              </span>
            ))}
            <span className="px-3 py-1.5 rounded-lg text-sm font-semibold text-gray-400 dark:text-gray-500">
              …
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}

Posture.propTypes = {
  label: PropTypes.string.isRequired,
  tone: PropTypes.oneOf(TONE_VALUES),
};

ToolHeader.propTypes = {
  logo: PropTypes.string.isRequired,
  alt: PropTypes.string,
  posture: PropTypes.string,
  postureTone: PropTypes.oneOf(TONE_VALUES),
  children: PropTypes.node,
};

Field.propTypes = {
  label: PropTypes.string.isRequired,
  tone: PropTypes.oneOf(TONE_VALUES),
  icon: PropTypes.string,
  children: PropTypes.node,
};
