import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaRocket,
  FaBook,
  FaGithub,
  FaLockOpen,
  FaServer,
  FaPlug,
  FaCheck,
  FaTimes,
} from 'react-icons/fa';
import {CodeCard} from '@site/src/components/home/HomepageQuickStart/CodeCard';
import SeoHead from '../../../components/landingPage/seo';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';
import suiteImage from '@site/static/img/landing/feature-flag/multi-ff.png';
import deployImage from '@site/static/img/landing/feature-flag/deploy.png';
import rolloutImage from '@site/static/img/landing/rollouts/progressive.png';
import killSwitchImage from '@site/static/img/landing/feature-flag/kill-switch.png';

const PAGE_TITLE = 'Why GO Feature Flag?';
const PAGE_DESCRIPTION =
  'Why choose GO Feature Flag: a lightweight, 100% open-source, self-hosted feature flag solution built on OpenFeature — complete feature management with no vendor lock-in and no per-seat pricing.';

const ROLLOUT_DOCS = '/docs/configure_flag/rollout-strategies/';

const DEPLOY_YAML = `new-checkout:
  variations:
    on: true
    off: false
  targeting:
    # Ship the code now, open it only to your own team
    - query: email ew "@your-company.com"
      variation: on
  defaultRule:
    variation: off`;

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const codeWrap = 'mx-auto w-full max-w-3xl';
const docLink =
  'mt-4 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const DIFFERENTIATOR_CARDS = [
  {
    icon: <FaLockOpen />,
    title: '100% open source',
    description:
      'MIT-licensed and free, forever. Read the code, contribute, and run it without a contract, a sales call, or a per-seat bill.',
    link: '/product/open_source',
    linkLabel: 'Open source',
  },
  {
    icon: <FaServer />,
    title: 'Self-hosted, your data stays yours',
    description:
      'Run it on your own infrastructure. Flag evaluation and usage data never leave your environment — no third party in the request path.',
    link: '/docs/relay-proxy/getting_started',
    linkLabel: 'Relay proxy',
  },
  {
    icon: <FaPlug />,
    title: 'Built on OpenFeature',
    description:
      'Evaluate flags through the vendor-neutral OpenFeature SDKs. No proprietary client to rip out later — swap the backend without rewriting your code.',
    link: '/product/open-feature',
    linkLabel: 'OpenFeature',
  },
];

// "GO Feature Flag vs. a SaaS feature-flag service" comparison rows.
const COMPARISON_ROWS = [
  {
    capability: 'Pricing',
    goff: 'Free — MIT licensed, no per-seat or per-MAU cost',
    saas: 'Paid, usually per seat or per monthly active user',
  },
  {
    capability: 'Hosting',
    goff: 'Self-hosted on your own infrastructure',
    saas: 'Vendor-hosted; your data leaves your environment',
  },
  {
    capability: 'SDK standard',
    goff: 'OpenFeature-native — vendor-neutral SDKs',
    saas: 'Proprietary SDK; switching means a rewrite',
  },
  {
    capability: 'Data ownership',
    goff: 'Evaluation and usage data stay in your stack',
    saas: 'Flag data flows through the vendor',
  },
  {
    capability: 'Source code',
    goff: 'Fully open — audit, fork, and contribute',
    saas: 'Closed source',
  },
];

const FAQ_ITEMS = [
  {
    question: 'Is GO Feature Flag really free?',
    answer:
      'Yes. GO Feature Flag is 100% open source under the MIT license. There is no paid tier, no per-seat pricing, and no per-monthly-active-user billing. You run it yourself and it stays free at any scale.',
  },
  {
    question: 'Do I have to self-host it?',
    answer: (
      <>
        GO Feature Flag is built to be self-hosted: you run the{' '}
        <Link to="/docs/relay-proxy/getting_started">relay proxy</Link> (or
        embed the Go module directly) on your own infrastructure, so flag
        evaluation and usage data never leave your environment. You keep full
        control of where your data lives and who can see it.
      </>
    ),
  },
  {
    question: 'What is OpenFeature and why does it matter?',
    answer: (
      <>
        <Link to="https://openfeature.dev">OpenFeature</Link> is a CNCF
        vendor-neutral standard and SDK for feature flagging. Because GO Feature
        Flag is OpenFeature-native, your application code talks to the standard
        SDK, not a proprietary client — so you are never locked in and can swap
        the backend without rewriting your integration.
      </>
    ),
  },
  {
    question: 'How is it different from LaunchDarkly or Flagsmith?',
    answer:
      'GO Feature Flag is open source, self-hosted, and free, where most commercial services are vendor-hosted and billed per seat or per active user. You get the same core capabilities — targeting, progressive and scheduled rollouts, experimentation, and data export — without sending your flag data to a third party or paying as you grow.',
  },
  {
    question: 'What can I export flag usage data to?',
    answer: (
      <>
        Flag evaluation events can be exported to S3, Google Cloud Storage,
        local files, Kafka, Kinesis, Pub/Sub, SQS, webhooks, OpenTelemetry, and
        more, so you can measure rollouts and run experiments with your own
        analytics. See the{' '}
        <Link to="/docs/integrations/export-evaluation-data">
          data export docs
        </Link>
        .
      </>
    ),
  },
  {
    question: 'Can I use it without the relay proxy?',
    answer:
      'Yes. You can embed the Go module directly in a Go application for in-process evaluation, or run the relay proxy to serve any language through the OpenFeature SDKs. Both share the same flag configuration and rollout features.',
  },
];

export default function WhyGoFeatureFlagPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  return (
    <Layout
      title={PAGE_TITLE}
      description={PAGE_DESCRIPTION}
      keywords={[
        'open source feature flags',
        'self-hosted feature flags',
        'OpenFeature',
        'free feature flag tool',
        'feature management',
        'LaunchDarkly alternative',
        'feature flag solution',
        'progressive rollout',
      ]}>
      <SeoHead
        title={PAGE_TITLE}
        description={PAGE_DESCRIPTION}
        path="/product/why_go_feature_flag"
        image={suiteImage}
        imageMeta
        imageWidth={1200}
        imageHeight={896}
      />

      <Title
        title={PAGE_TITLE}
        description="Ship faster, reduce risk, and stay in control. GO Feature Flag gives you complete feature management — open source, self-hosted, and built on OpenFeature — without vendor lock-in or a per-seat bill."
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
        title="Feature management without the lock-in or the bill"
        imageSrc={suiteImage}
        imageAlt="A single flag configuration powering targeting, rollouts, and data export"
        imageWidth={1200}
        imageHeight={896}>
        <p className={prose}>
          Feature flags should be simple, accessible, and yours. Most teams hit
          a wall where the tooling means a contract, a per-seat bill, and their
          flag data flowing through someone else&rsquo;s servers.
        </p>
        <p className="mb-0">
          GO Feature Flag takes the other path: a <strong>lightweight</strong>,{' '}
          <strong>100% open-source</strong> solution you{' '}
          <strong>self-host</strong>, built on the{' '}
          <Link to="/product/open-feature">OpenFeature</Link> standard so
          you&rsquo;re never tied to a vendor. Start in minutes and keep every
          advanced capability as you grow.
        </p>
      </FeatureRow>

      <Cards
        title="Why teams choose GO Feature Flag:"
        cards={DIFFERENTIATOR_CARDS}
      />

      {/* Complete feature management */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Everything you need, nothing you don&rsquo;t
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          GO Feature Flag is not just an on/off switch — it&rsquo;s a complete
          feature-management suite that stays free and self-hosted.
        </p>
        <ul className="mx-auto mb-0 max-w-3xl list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>
              <Link to="/docs/configure_flag/target-with-flags">
                Targeting &amp; segmentation
              </Link>
              .
            </strong>{' '}
            Show a feature only to the users who should see it, with rich rules
            on any context attribute.
          </li>
          <li>
            <strong>
              <Link to="/product/rollouts">Rollout strategies</Link>.
            </strong>{' '}
            A/B testing, progressive ramps, and scheduled flag changes — all
            from configuration.
          </li>
          <li>
            <strong>
              <Link to={`${ROLLOUT_DOCS}experimentation`}>Experimentation</Link>
              .
            </strong>{' '}
            Run measured experiments over a fixed window and keep the variation
            that actually wins.
          </li>
          <li>
            <strong>
              <Link to="/docs/integrations/export-evaluation-data">
                Data export
              </Link>
              .
            </strong>{' '}
            Stream evaluation data to S3, Google Cloud Storage, local files,
            Kafka, and more for your own analytics.
          </li>
          <li>
            <strong>
              <Link to="/docs/integrations/notify-flags-changes">
                Change notifications
              </Link>
              .
            </strong>{' '}
            Get alerted in Slack, Microsoft Teams, Discord, or a webhook the
            moment a flag changes.
          </li>
        </ul>
      </section>

      {/* Decouple deploy from release */}
      <FeatureRow
        eyebrow="Ship faster"
        title="Decouple deploy from release"
        reverse
        imageSrc={deployImage}
        imageAlt="Code shipping to production while a flag controls who sees the feature"
        imageWidth={1200}
        imageHeight={896}>
        <p className="mb-4">
          Merge and deploy whenever you&rsquo;re ready, then decide separately
          who sees the feature and when. The code goes out <strong>dark</strong>
          , gated behind a flag, so releasing is a configuration change — not
          another deploy.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> turn it on for internal users first,
          watch it in production, then widen to everyone.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={DEPLOY_YAML}
            callout="Default off. The new checkout ships with everything else, but only your team sees it until you widen."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to="/product/test_in_production">
            Test in production <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Reduce risk */}
      <FeatureRow
        eyebrow="Reduce risk"
        title="Roll out gradually, prove it as you go"
        imageSrc={rolloutImage}
        imageAlt="A feature ramping from a small percentage of traffic to everyone"
        imageWidth={1200}
        imageHeight={896}>
        <p className="mb-4">
          Release to a small slice of traffic, watch your metrics, and ramp up
          only when the data stays green. Evaluation is deterministic on the
          targeting key, so a given user keeps a consistent experience while you
          widen the rollout.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> start at a few percent, then move to a{' '}
          <Link to={`${ROLLOUT_DOCS}progressive`}>progressive rollout</Link> to
          reach everyone safely.
        </p>
      </FeatureRow>

      {/* Instant kill switch */}
      <FeatureRow
        eyebrow="Stay in control"
        title="Turn anything off in seconds"
        reverse
        imageSrc={killSwitchImage}
        imageAlt="A switch cutting a risky feature back to its safe default"
        imageWidth={1200}
        imageHeight={747}>
        <p className="mb-4">
          When something misbehaves, you don&rsquo;t debug under fire — you flip
          the switch. Change one line in the flag and every user drops back to
          the safe path on the relay proxy&rsquo;s next poll.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> keep a safe default variation and a
          kill switch on every risky feature — no redeploy, no rollback.
        </p>
      </FeatureRow>

      {/* Comparison */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          GO Feature Flag vs. a SaaS feature-flag service
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          The same core capabilities as a commercial service — without the bill,
          the lock-in, or handing over your data.
        </p>
        <div className="mx-auto w-full max-w-4xl overflow-x-auto">
          <table className="w-full border-collapse text-left align-top">
            <thead>
              <tr className="border-0 border-b-2 border-solid border-gray-200 dark:border-gray-700">
                <th className="py-3 pr-4 font-bold text-gray-800 dark:text-gray-50">
                  Capability
                </th>
                <th className="py-3 pr-4 font-bold text-gray-800 dark:text-gray-50">
                  <span className="inline-flex items-center gap-2">
                    <FaCheck className="text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]" />
                    GO Feature Flag
                  </span>
                </th>
                <th className="py-3 font-bold text-gray-800 dark:text-gray-50">
                  <span className="inline-flex items-center gap-2">
                    <FaTimes className="text-gray-400" />
                    Typical SaaS service
                  </span>
                </th>
              </tr>
            </thead>
            <tbody className="text-gray-700 dark:text-gray-300">
              {COMPARISON_ROWS.map(row => (
                <tr
                  key={row.capability}
                  className="border-0 border-b border-solid border-gray-200 align-top dark:border-gray-700">
                  <td className="py-3 pr-4 font-semibold">{row.capability}</td>
                  <td className="py-3 pr-4">{row.goff}</td>
                  <td className="py-3">{row.saas}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <CtaBand
        title="Start flagging in minutes — for free"
        description="Self-hosted, OpenFeature-native, MIT-licensed. Get complete feature management without a contract or a per-seat bill."
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
