import React from 'react';
import Layout from '@theme/Layout';
import Head from '@docusaurus/Head';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaRocket,
  FaBook,
  FaGithub,
  FaCodeBranch,
  FaShieldAlt,
  FaPowerOff,
} from 'react-icons/fa';
import {CodeCard} from '@site/src/components/home/HomepageQuickStart/CodeCard';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';
import aiHero from '@site/static/img/landing/ai/hero.png';
import aiCode from '@site/static/img/landing/ai/promptb.png';
import aiModelRollout from '@site/static/img/landing/ai/promptc.png';
import aiExperiment from '@site/static/img/landing/ai/promptd.png';
import aiKillSwitch from '@site/static/img/landing/ai/prompte.png';

const PAGE_TITLE = 'Feature flags for AI';
const PAGE_DESCRIPTION =
  'Why feature flags are the safety net for AI-generated code and AI features: ship agent-written code dark, roll a model out to a percentage, A/B test prompts, and keep an instant kill switch - no redeploy.';

const ROLLOUT_DOCS = '/docs/configure_flag/rollout-strategies/';

const ASSISTANT_CODE_YAML = `ai-summary:
  variations:
    on: true
    off: false
  targeting:
    # Internal staff get the agent-written path first
    - query: email ew "@your-company.com"
      variation: on
  defaultRule:
    variation: off`;

const MODEL_ROLLOUT_YAML = `llm-model:
  variations:
    current: "gpt-4o-mini"
    candidate: "gpt-4o"
  defaultRule:
    percentage:
      candidate: 10 # 10% of users hit the new model
      current: 90`;

const EXPERIMENT_YAML = `support-prompt:
  variations:
    promptA: "concise-v1"
    promptB: "detailed-v2"
  defaultRule:
    percentage:
      promptA: 50
      promptB: 50
  experimentation:
    start: 2026-07-01T00:00:00.1-05:00
    end: 2026-07-08T00:00:00.1-05:00`;

const KILL_SWITCH_YAML = `ai-chat:
  variations:
    live: "llm"
    fallback: "rules-engine"
  targeting:
    - query: beta eq "true"
      variation: live
  defaultRule:
    # Flip this one line to "fallback" to kill the AI path
    variation: live`;

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const codeWrap = 'mx-auto w-full max-w-3xl';
const docLink =
  'mt-4 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const CAPABILITY_CARDS = [
  {
    icon: <FaCodeBranch />,
    title: 'Merge AI code dark',
    description:
      'Agent-written code lands behind a flag, off by default, instead of going straight to every user.',
  },
  {
    icon: <FaShieldAlt />,
    title: 'Contain the blast radius',
    description:
      'A hallucination or a bad model only ever reaches the slice of traffic you opened it to.',
  },
  {
    icon: <FaPowerOff />,
    title: 'Kill it instantly',
    description:
      'One YAML change turns the AI path off and falls back to the safe one - no rollback, no redeploy.',
  },
];

const WHEN_ROWS = [
  {
    name: 'AI-written code path',
    href: '/docs/configure_flag/target-with-flags',
    when: 'A coding agent wrote a new path you have not fully reviewed in production.',
    instead:
      'Gate it behind a boolean flag, default off, and open it to internal users first.',
  },
  {
    name: 'New / upgraded model',
    href: `${ROLLOUT_DOCS}progressive`,
    when: 'You are swapping in a new LLM and want to prove it before everyone hits it.',
    instead: 'Canary it with a percentage, then ramp it progressively.',
  },
  {
    name: 'Prompt change',
    href: `${ROLLOUT_DOCS}experimentation`,
    when: 'You changed a prompt and need to know it is actually better, not just different.',
    instead:
      'Run an experimentation rollout and measure the two against each other.',
  },
  {
    name: 'New AI feature',
    href: '/product/test_in_production',
    when: 'A user-facing AI feature is ready to ship but unproven at scale.',
    instead: 'Release it dark, then widen exposure as the data stays green.',
  },
  {
    name: 'Deterministic fallback',
    href: '/docs/configure_flag/create-flags',
    when: 'The AI path can fail, time out, or get expensive without warning.',
    instead:
      'Keep a non-AI variation and a kill switch that flips to it in one edit.',
  },
];

const FAQ_ITEMS = [
  {
    question: 'Why do AI features need a feature flag?',
    answer:
      'AI is non-deterministic: the same input can produce a different output for every user and every run. A flag lets you release that behavior to a small slice first, measure it, and reverse it instantly if quality, latency, or cost regresses - without shipping new code.',
  },
  {
    question: 'How do I roll out a new LLM model safely?',
    answer:
      'Put the model name in a flag variation and use a percentage or progressive rollout. Start at a few percent, watch your metrics, then ramp to everyone. Because evaluation is deterministic on the targeting key, a given user keeps a consistent experience while you widen the rollout.',
  },
  {
    question: 'Can I A/B test two prompts or two models?',
    answer:
      'Yes. Define a variation per prompt or model, split traffic with a percentage, and wrap it in an experimentation rollout so it runs for a fixed window. Pair it with the data export to measure which variation actually won.',
  },
  {
    question: 'What happens when an AI feature misbehaves?',
    answer:
      'Keep a deterministic fallback variation - a rules engine, a cached answer, or the previous model - and a kill switch. Flipping the flag routes everyone to the safe path on the relay proxy’s next poll, with no redeploy and no rollback.',
  },
  {
    question: 'Do I need to redeploy to change which model is live?',
    answer:
      'No. The model lives in the flag configuration. Edit the YAML and the relay proxy picks it up on its next poll, so swapping models, changing the split, or killing the feature never requires a deploy.',
  },
  {
    question: 'Can I gate code my coding agent wrote?',
    answer: (
      <>
        Yes - that is one of the strongest uses. Wrap the agent-written path in
        a <Link to="/product/what-are-feature-flags">feature flag</Link> that
        defaults off, merge it dark, and turn it on for internal users before
        widening. The unreviewed code reaches production behind a switch instead
        of going live to everyone in a single deploy.
      </>
    ),
  },
];

export default function AIPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  const siteUrl = siteConfig.url;
  const pageUrl = `${siteUrl}/product/ai`;
  // Use the default GO Feature Flag social card (themeConfig.image) for sharing.
  const ogImage = `${siteUrl}/img/logo/x-card.png`;
  const structuredData = [
    {
      '@context': 'https://schema.org',
      '@type': 'TechArticle',
      headline: PAGE_TITLE,
      description: PAGE_DESCRIPTION,
      image: ogImage,
      author: {'@type': 'Organization', name: siteConfig.title, url: siteUrl},
      publisher: {
        '@type': 'Organization',
        name: siteConfig.title,
        logo: {'@type': 'ImageObject', url: `${siteUrl}/img/logo/logo.png`},
      },
      datePublished: '2026-06-30',
      dateModified: '2026-06-30',
      mainEntityOfPage: {'@type': 'WebPage', '@id': pageUrl},
    },
    {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      itemListElement: [
        {'@type': 'ListItem', position: 1, name: 'Home', item: `${siteUrl}/`},
        {'@type': 'ListItem', position: 2, name: PAGE_TITLE, item: pageUrl},
      ],
    },
  ];

  return (
    <Layout
      title={PAGE_TITLE}
      description={PAGE_DESCRIPTION}
      keywords={[
        'feature flags for AI',
        'AI feature flag',
        'release AI code',
        'LLM rollout',
        'AI model rollout',
        'prompt A/B test',
        'AI kill switch',
        'ship AI safely',
      ]}>
      <Head>
        <link rel="canonical" href={pageUrl} />
        <script type="application/ld+json">
          {JSON.stringify(structuredData)}
        </script>
      </Head>

      <Title
        title={PAGE_TITLE}
        description="AI writes code faster than you can review it, and AI features behave differently for every user. A feature flag puts a switch in front of both - ship dark, release to a slice, and kill it the moment it misbehaves. No redeploy."
        actions={[
          {
            label: 'Get started',
            href: '/docs/getting-started',
            icon: <FaRocket />,
          },
          {
            label: 'Read the docs',
            href: '/docs',
            icon: <FaBook />,
            variant: 'secondary',
          },
        ]}
      />

      <FeatureRow
        eyebrow="The core idea"
        title="AI ships faster than you can verify"
        imageSrc={aiHero}
        imageAlt="AI code firehose funneling through a single flag switch"
        imageWidth={1200}
        imageHeight={655}>
        <p className={prose}>
          Coding agents and LLM features generate more code and more behavior
          than any team can review line by line. The bottleneck is no longer
          writing it - it is being sure it is safe in front of real users.
        </p>
        <p className="mb-0">
          A <Link to="/product/what-are-feature-flags">feature flag</Link> wraps
          the new path so it merges <strong>dark</strong>: the code ships, but
          exposure stays under your control, decoupled from deploy. You decide{' '}
          <strong>who</strong> sees the AI and <strong>when</strong> - and you
          can take it back in seconds.
        </p>
      </FeatureRow>

      <Cards
        title="Feature flags let your AI work safely:"
        cards={CAPABILITY_CARDS}
      />

      {/* Why it is risky without a flag */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Why shipping AI without a flag is risky
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          Traditional code is deterministic - you can review it once and trust
          it. AI is not. That breaks the assumptions a big-bang release relies
          on.
        </p>
        <ul className="mx-auto mb-0 max-w-3xl list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Non-determinism.</strong> The same prompt can return a
            different answer per user and per run - you cannot fully test it
            before release.
          </li>
          <li>
            <strong>Silent quality regressions.</strong> A new model or a
            tweaked prompt can get subtly worse in ways no unit test catches.
          </li>
          <li>
            <strong>Cost and latency spikes.</strong> A bigger model can quietly
            multiply your bill and your p99 the moment it goes live.
          </li>
          <li>
            <strong>Unreviewed agent code at 100%.</strong> Without a flag,
            machine-written code reaches every user in a single deploy.
          </li>
          <li>
            <strong>No fast way back.</strong> Reverting means a redeploy or a
            rollback - minutes to hours - instead of a one-line flag flip.
          </li>
        </ul>
      </section>

      {/* AI-generated code */}
      <FeatureRow
        eyebrow="AI-generated code"
        title="Wrap agent output in a flag"
        reverse
        imageSrc={aiCode}
        imageAlt="Agent-written code entering a gated pipe behind a flag"
        imageWidth={1200}
        imageHeight={896}>
        <p className="mb-4">
          When Copilot, Cursor, or Claude Code writes a new path, gate it behind
          a boolean flag that defaults <strong>off</strong>. The code merges and
          deploys with everything else, but no user reaches it until you say so.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> turn it on for internal staff first,
          watch it in production, then widen to everyone - or flip it back if it
          misbehaves.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={ASSISTANT_CODE_YAML}
            callout="Default off. Only your own team gets the agent-written summary until you widen it."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to="/docs/configure_flag/target-with-flags">
            Targeting docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Model rollout */}
      <FeatureRow
        eyebrow="AI features"
        title="Roll a model out to a percentage"
        imageSrc={aiModelRollout}
        imageAlt="User traffic split between two models on a rising ramp"
        imageWidth={1200}
        imageHeight={896}>
        <p className="mb-4">
          Put the model name in a flag variation and send a small slice of
          traffic to the new one. The split is deterministic on the targeting
          key, so a given user keeps a consistent experience while you ramp.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> a canary - start at a few percent,
          watch quality, cost, and latency, then ramp to everyone with a{' '}
          <Link to={`${ROLLOUT_DOCS}progressive`}>progressive rollout</Link>.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={MODEL_ROLLOUT_YAML}
            callout="Bump candidate to 25, then 100, as the new model proves out. No redeploy."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={`${ROLLOUT_DOCS}progressive`}>
            Progressive rollout docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Compare models / prompts */}
      <FeatureRow
        eyebrow="Compare models & prompts"
        title="A/B test which AI is actually better"
        reverse
        imageSrc={aiExperiment}
        imageAlt="Two AI variants compared inside a bracketed start-end window"
        imageWidth={1200}
        imageHeight={896}>
        <p className="mb-4">
          Different is not the same as better. Define a variation per prompt or
          model, split traffic 50/50, and wrap it in an experimentation rollout
          so it runs for a fixed window.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> pair it with the{' '}
          <Link to="/docs/integrations/export-evaluation-data">
            data export
          </Link>{' '}
          to measure which variation won, then keep the winner.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={EXPERIMENT_YAML}
            callout="A clean one-week measurement of two prompts, then an automatic return to the default."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={`${ROLLOUT_DOCS}experimentation`}>
            Experimentation docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Kill switch */}
      <FeatureRow
        eyebrow="Kill switch"
        title="Turn the AI off in seconds"
        imageSrc={aiKillSwitch}
        imageAlt="A large switch cutting an AI path back to a safe deterministic fallback"
        imageWidth={1200}
        imageHeight={896}>
        <p className="mb-4">
          Keep a deterministic fallback variation - a rules engine, a cached
          answer, or the previous model - alongside the live AI path. When
          something goes wrong, you do not debug under fire; you flip the
          switch.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> change one line in the flag and every
          user drops to the safe path on the relay proxy&rsquo;s next poll - no
          redeploy, no rollback.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={KILL_SWITCH_YAML}
            callout="Flip defaultRule to fallback and the whole AI feature is off for everyone in seconds."
          />
        </div>
      </section>

      {/* What to flag */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          What to flag in your AI stack
        </h2>
        <p className={`${prose} text-center`}>
          Anywhere AI touches production is a place to put a switch - here is
          what to reach for and when.
        </p>
        <div className="mx-auto w-full max-w-4xl overflow-x-auto">
          <table className="w-full border-collapse text-left align-top">
            <thead>
              <tr className="border-0 border-b-2 border-solid border-gray-200 dark:border-gray-700">
                <th className="py-3 pr-4 font-bold text-gray-800 dark:text-gray-50">
                  What
                </th>
                <th className="py-3 pr-4 font-bold text-gray-800 dark:text-gray-50">
                  Reach for it when
                </th>
                <th className="py-3 font-bold text-gray-800 dark:text-gray-50">
                  How to flag it
                </th>
              </tr>
            </thead>
            <tbody className="text-gray-700 dark:text-gray-300">
              {WHEN_ROWS.map(row => (
                <tr
                  key={row.name}
                  className="border-0 border-b border-solid border-gray-200 align-top dark:border-gray-700">
                  <td className="py-3 pr-4 font-semibold">
                    <Link
                      className="text-[color:var(--ifm-color-primary-dark)] no-underline hover:underline dark:text-[color:var(--ifm-color-primary)]"
                      to={row.href}>
                      {row.name}
                    </Link>
                  </td>
                  <td className="py-3 pr-4">{row.when}</td>
                  <td className="py-3">{row.instead}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {/* Pitfalls */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>Pitfalls to avoid</h2>
        <ul className="mx-auto mb-0 max-w-3xl list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Shipping agent code straight to 100%.</strong>{' '}
            Machine-written code deserves the same dark launch as anything else
            - default the flag off and widen on evidence.
          </li>
          <li>
            <strong>No deterministic fallback.</strong> If the only path is the
            AI path, an outage or a bad answer has nowhere to fall back to.
          </li>
          <li>
            <strong>No kill switch.</strong> Always keep the one-line flip that
            takes the feature off without a deploy.
          </li>
          <li>
            <strong>Bucketing experiments on an unstable key.</strong>{' '}
            Consistency rides on the targeting key; a value that changes per
            request flips users between models mid-session.
          </li>
          <li>
            <strong>Leaving a finished AI rollout in the code.</strong> Once a
            model is at 100% for everyone, the flag is debt - clean it up.
          </li>
        </ul>
      </section>

      <CtaBand
        title="Ship AI with a safety net"
        description="Self-hosted, OpenFeature-native, MIT-licensed. Gate every model, prompt, and agent-written path behind a flag - and kill it with a one-line YAML change."
        actions={[
          {
            label: 'Get started',
            href: '/docs/getting-started',
            icon: <FaRocket />,
          },
          {
            label: 'View on GitHub',
            href: githubUrl,
            icon: <FaGithub />,
            variant: 'secondary',
          },
        ]}
      />

      <Faq title="Frequently asked questions" items={FAQ_ITEMS} withJsonLd />
    </Layout>
  );
}
