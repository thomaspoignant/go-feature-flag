import React, {useState} from 'react';
import clsx from 'clsx';
import PropTypes from 'prop-types';
import {CodeCard} from './CodeCard';
import {sdk} from '../../../../data/sdk';

const dockerSnippet = String.raw`docker run \
    -v $(pwd)/flags.goff.yaml:/goff/flags.goff.yaml \
    -p 1031:1031 \
    -e RETRIEVERS_0_KIND=file \
    -e RETRIEVERS_0_PATH=/goff/flags.goff.yaml \
    gofeatureflag/go-feature-flag:latest`;

const yamlSnippet = `# Roll out to 10% of users — increase anytime
my-new-feature:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    percentage:
      enabled: 10
      disabled: 90`;

const sdkTabs = sdk
  .sort((a, b) => {
    const aIsClient = a.paradigm?.includes('Client');
    const bIsClient = b.paradigm?.includes('Client');
    if (aIsClient === bIsClient) return 0;
    return aIsClient ? -1 : 1;
  })
  .filter(s => s.snippets !== undefined)
  .map(s => {
    const paradigmSuffix = s.paradigm?.includes('Client')
      ? ' (Client)'
      : ' (Server)';
    return {
      value: s.key,
      label: s.name,
      language: s.language,
      displayName: `${s.name} ${paradigmSuffix}`,
      icon: s.icon,
      code: s.snippets,
    };
  });

const steps = [
  {
    id: 1,
    label: 'Define your flag',
    description: 'Create a YAML file with your flag and rollout rules.',
  },
  {
    id: 2,
    label: 'Run the server',
    description: 'Start GO Feature Flag locally with a single Docker command.',
  },
  {
    id: 3,
    label: 'Evaluate in your app',
    description: 'Use any OpenFeature SDK to evaluate your flag.',
  },
];

StepCard.propTypes = {
  step: PropTypes.shape({
    id: PropTypes.number.isRequired,
    label: PropTypes.string.isRequired,
    description: PropTypes.string.isRequired,
  }).isRequired,
  active: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
};

function StepCard({step, active, onClick}) {
  return (
    <div className="relative">
      <button
        type="button"
        onClick={onClick}
        aria-pressed={active}
        className={clsx(
          'w-full h-full text-left p-2 rounded-xl font-poppins transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--ifm-color-primary)] focus:ring-offset-white dark:focus:ring-offset-[#242526]',
          'bg-gray-50 dark:bg-[#1f1f20] border-2',
          active
            ? 'border-[var(--ifm-color-primary)] shadow-[0_0_0_1px_var(--ifm-color-primary)]'
            : 'border-transparent hover:border-gray-300 dark:hover:border-gray-700'
        )}>
        <div className="flex items-center gap-3">
          <span className="flex items-center justify-center w-7 h-7 shrink-0 rounded-full bg-[var(--ifm-color-primary-lightest)] text-gray-900 text-sm font-bold">
            {step.id}
          </span>
          <span className="text-base md:text-lg font-semibold text-gray-900 dark:text-gray-50">
            {step.label}
          </span>
        </div>
        <p className="mt-3 text-sm text-gray-600 dark:text-gray-400 leading-relaxed font-poppins">
          {step.description}
        </p>
      </button>
      {active && (
        <div
          aria-hidden="true"
          className="absolute left-1/2 -translate-x-1/2 -bottom-4 w-8 h-8 rounded-full bg-gray-200 dark:bg-[#363636] flex items-center justify-center shadow-md">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
            className="w-4 h-4 text-gray-600 dark:text-gray-300">
            <path
              fillRule="evenodd"
              d="M5.23 7.21a.75.75 0 011.06.02L10 11.06l3.71-3.83a.75.75 0 111.08 1.04l-4.25 4.39a.75.75 0 01-1.08 0L5.21 8.27a.75.75 0 01.02-1.06z"
              clipRule="evenodd"
            />
          </svg>
        </div>
      )}
    </div>
  );
}

export function QuickStart() {
  const [step, setStep] = useState(1);

  return (
    <section className="py-[3rem] text-center bg-white dark:bg-[#242526]">
      <span className="text-gray-800 dark:text-gray-50 text-5xl font-poppins font-bold tracking-[-0.18rem]">
        Up and running in 60 seconds
      </span>
      <p className="mt-4 text-gray-600 dark:text-gray-300 font-poppins text-lg">
        A Docker container, a YAML file, and two lines of SDK code.
      </p>

      <div className="max-w-5xl mx-auto px-4 mt-10">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-2">
          {steps.map(s => (
            <StepCard
              key={s.id}
              step={s}
              active={step === s.id}
              onClick={() => setStep(s.id)}
            />
          ))}
        </div>
        {step === 1 && (
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={yamlSnippet}
            callout={
              <>
                Change percentages, targeting rules, or add new flags —{' '}
                <strong>no redeploy needed</strong>.
              </>
            }
          />
        )}
        {step === 2 && (
          <CodeCard
            filename="Terminal"
            language="shell"
            code={dockerSnippet}
            analyticsEvent="relay_proxy_download"
            analyticsMethod="docker"
            callout={
              <>
                Relay proxy listening on{' '}
                <code>http://localhost:1031 — no signup, low infra.</code>
              </>
            }
          />
        )}
        {step === 3 && (
          <CodeCard
            tabs={sdkTabs}
            moreLink={{
              to: '/docs/sdk',
              label: 'More (NodeJS, Python, React, Angular, PHP, Ruby ...)',
              ariaLabel: 'See all supported SDKs',
            }}
            callout={
              <>
                Built on the <strong>OpenFeature standard</strong> — swap
                providers without rewriting evaluation code.
              </>
            }
          />
        )}
      </div>
    </section>
  );
}
