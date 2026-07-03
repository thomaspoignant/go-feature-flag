import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaRocket,
  FaBook,
  FaGithub,
  FaRandom,
  FaLock,
  FaClock,
  FaDatabase,
  FaCloud,
  FaStream,
  FaBolt,
} from 'react-icons/fa';
import {CodeCard} from '@site/src/components/home/HomepageQuickStart/CodeCard';
import SeoHead from '../../../components/landingPage/seo';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';

const PAGE_TITLE = 'A/B testing with feature flags';
const PAGE_DESCRIPTION =
  'How to run A/B tests with GO Feature Flag: split users into A and B with the ' +
  'evaluation engine, then export the data so you can measure which variation won.';

const EXPERIMENTATION_DOCS =
  '/docs/configure_flag/rollout-strategies/experimentation';
const EXPORTER_DOCS = '/docs/concepts/exporter';
const TRACKING_DOCS = '/docs/tracking/tracking-api';

// Step 1 - evaluation: a deterministic 50/50 split, bounded to a test window.
const SPLIT_YAML = `checkout-experiment:
  variations:
    control: "current-checkout"
    candidate: "new-checkout"
  defaultRule:
    percentage:
      control: 50
      candidate: 50
  # only run the experiment inside this window
  experimentation:
    start: 2026-04-01T00:00:00.1-05:00
    end: 2026-04-15T00:00:00.1-05:00`;

// Step 2 - exporters: two destinations, one per event type.
const EXPORTER_YAML = `# goff-proxy.yaml
exporters:
  # exposures: which user saw which variation
  - kind: bigquery
    projectID: "my-project"
    datasetID: "goff_experiments"
    tableName: "feature_flag_evaluations"
    eventType: "feature"
  # outcomes: what each user actually did
  - kind: bigquery
    projectID: "my-project"
    datasetID: "goff_experiments"
    tableName: "tracking_events"
    eventType: "tracking"`;

// Step 3 - tracking: record the outcome against the same targeting key.
const TRACKING_CODE = `// When the user completes the action you care about,
// record it against the SAME context you evaluate flags with.
client.track("checkout-completed", evaluationContext, {
  value: 99.77,
  currencyCode: "USD",
});`;

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const codeWrap = 'mx-auto w-full max-w-3xl';
const docLink =
  'mt-4 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const CAPABILITY_CARDS = [
  {
    icon: <FaRandom />,
    title: 'A stable, deterministic split',
    description:
      'Users are bucketed by hashing the targeting key, so the same person stays in the same group for the whole test - no flicker between A and B.',
  },
  {
    icon: <FaLock />,
    title: 'Your data, your warehouse',
    description:
      'Exposures and outcomes are exported to a destination you own - BigQuery, S3, Kafka, a webhook - so you analyse results with the tools you already trust.',
  },
  {
    icon: <FaClock />,
    title: 'Time-boxed and reversible',
    description:
      'An experimentation rollout runs the split only inside a window you set, then falls back to the default automatically - no cleanup redeploy.',
  },
];

const DESTINATION_CARDS = [
  {
    icon: <FaDatabase />,
    title: 'Data warehouses',
    description:
      'Stream feature and tracking events straight into BigQuery for SQL analysis per variation.',
  },
  {
    icon: <FaCloud />,
    title: 'Object storage',
    description:
      'Batch events to AWS S3, Google Cloud Storage, Azure Blob Storage, or the local file system as JSON, CSV, or Parquet.',
  },
  {
    icon: <FaStream />,
    title: 'Streaming & queues',
    description:
      'Push events in near real time to Kafka, AWS Kinesis, Google Cloud Pub/Sub, or AWS SQS.',
  },
  {
    icon: <FaBolt />,
    title: 'Webhook & OpenTelemetry',
    description:
      'Send events to any HTTP endpoint, or emit them as OpenTelemetry signals for your observability stack.',
  },
];

const FAQ_ITEMS = [
  {
    question: 'Do I need a separate A/B testing tool with GO Feature Flag?',
    answer:
      'No. GO Feature Flag already gives you the two halves of an experiment: the evaluation engine assigns each user to a variation, and exporters ship the data out. You run the analysis in whatever analytics tool or warehouse you already use - there is no extra experimentation SaaS to buy.',
  },
  {
    question: 'How does GO Feature Flag decide which users are in A vs B?',
    answer:
      'It hashes the evaluation context’s targeting key deterministically. The same user always lands in the same variation as long as the percentages do not change, so their experience stays consistent for the whole test - and it is consistent across servers and restarts.',
  },
  {
    question: 'Where does my experiment data go?',
    answer:
      'Wherever you point your exporters. Every evaluation emits a feature event (who saw which variation) and the Tracking API emits tracking events (what they did). Both flow through exporters to destinations you control: BigQuery, S3, GCS, Azure Blob, Kafka, Kinesis, Pub/Sub, SQS, a webhook, or OpenTelemetry.',
  },
  {
    question: 'How is an A/B test different from a progressive rollout?',
    answer:
      'A progressive rollout is about shipping a change gradually - you are not measuring, you are ramping. An A/B test holds a fixed split for a fixed window so you can compare outcomes cleanly. Use the experimentation rollout (not a percentage you keep re-tuning) when you need a stable measurement group.',
  },
  {
    question: 'How long should I run an A/B test?',
    answer:
      'Long enough to reach statistical significance for the metric you care about - stopping early (peeking) leads to false winners. A duration calculator such as vwo’s helps you size the window before you start, which you then set as the experimentation start and end dates.',
  },
  {
    question: 'Does this work with the relay proxy and any SDK?',
    answer:
      'Yes. A/B testing is built on the standard evaluation and export paths, so it works the same across every OpenFeature SDK talking to the relay proxy. Exposures are captured automatically on evaluation, and outcomes are recorded through the OpenFeature Tracking API in your SDK.',
  },
];

export default function AbTestingPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  return (
    <Layout
      title={PAGE_TITLE}
      description={PAGE_DESCRIPTION}
      keywords={[
        'A/B testing',
        'feature flag A/B testing',
        'experimentation',
        'split testing',
        'feature experimentation',
        'export evaluation data',
        'conversion tracking',
        'OpenFeature tracking',
      ]}>
      <SeoHead
        title={PAGE_TITLE}
        description={PAGE_DESCRIPTION}
        path="/product/ab-testing"
        image="/img/logo/x-card.png"
      />

      <Title
        title={PAGE_TITLE}
        description="Ship two versions, measure the winner. GO Feature Flag splits your users into A and B, then exports the data so your own analytics can tell you which one actually performed better."
        actions={[
          {
            label: 'Get started',
            href: '/docs/getting-started',
            icon: <FaRocket />,
          },
          {
            label: 'Experimentation docs',
            href: EXPERIMENTATION_DOCS,
            icon: <FaBook />,
            variant: 'secondary',
          },
        ]}
      />

      <FeatureRow
        eyebrow="The idea"
        title="Test two versions, let the data pick the winner"
        imageSrc={useBaseUrl('/img/landing/ab-testing/schema-overview.svg')}
        imageAlt="Users are split into variation A and variation B; every evaluation and outcome is exported to a database, where a comparison shows which variation won."
        imageWidth={800}
        imageHeight={600}
        placeholderLabel="A/B testing loop - split, export, measure the winner">
        <p className={prose}>
          A/B test is the shorthand for a simple controlled experiment: two
          versions, <strong>A</strong> and <strong>B</strong>, are shown to
          comparable groups of users, and you measure which one moves the metric
          you care about.
        </p>
        <p className="mb-0">
          You do not need a separate experimentation platform for this. GO
          Feature Flag already gives you the two building blocks:{' '}
          <strong>evaluation</strong> to split users into A and B, and{' '}
          <strong>exporters</strong> to capture what happened - so you can
          decide the winner with the analytics you already own.
        </p>
      </FeatureRow>

      <Cards title="What you get" cards={CAPABILITY_CARDS} />

      {/* How it works - the recipe */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          A/B testing in three steps
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          The recipe is the same one the docs recommend: an{' '}
          <Link to={EXPERIMENTATION_DOCS}>experimentation rollout</Link>{' '}
          combined with the <Link to={EXPORTER_DOCS}>export of your data</Link>.
          Split your users, capture who saw what, record what they did - then
          read the result. Every step lives in configuration or a single SDK
          call, so there is no redeploy to start or stop a test.
        </p>
      </section>

      {/* Step 1 - evaluation */}
      <FeatureRow
        eyebrow="Step 1 · Evaluation"
        title="Split your users into A and B"
        imageSrc={useBaseUrl('/img/landing/ab-testing/schema-split.svg')}
        imageAlt="A crowd of users is deterministically routed through a hashing node into two equal groups, A and B, inside a bounded time window."
        imageWidth={800}
        imageHeight={600}
        placeholderLabel="Deterministic 50/50 split inside a start-end window">
        <p className="mb-4">
          A{' '}
          <Link to="/docs/configure_flag/rollout-strategies/percentage">
            percentage rollout
          </Link>{' '}
          sends a share of traffic to each variation - say 50/50. The split is
          deterministic on the{' '}
          <Link to="/docs/configure_flag/custom-bucketing">targeting key</Link>,
          so a user keeps the same variation for the whole test instead of
          flipping between requests.
        </p>
        <p className="mb-0">
          Wrap it in an <strong>experimentation rollout</strong> to bound the
          test to a start and end date. Inside the window users get the split;
          outside it, everyone falls back to the default - a clean measurement
          period with an automatic end.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={SPLIT_YAML}
            callout="A deterministic 50/50 split that only runs for two weeks, then returns to the control."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={EXPERIMENTATION_DOCS}>
            Experimentation rollout docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Step 2 - exporters */}
      <FeatureRow
        eyebrow="Step 2 · Exporters"
        title="Capture who saw which variation"
        reverse
        imageSrc={useBaseUrl('/img/landing/ab-testing/schema-collect.svg')}
        imageAlt="Evaluation events stream through an exporter and fan out to destinations: a database, object storage, and a message queue."
        imageWidth={800}
        imageHeight={600}
        placeholderLabel="Events exported to a warehouse, storage, and a queue">
        <p className="mb-4">
          Every flag evaluation emits a <strong>feature event</strong> - the
          targeting key, the flag, and the variation that user received. An{' '}
          <Link to={EXPORTER_DOCS}>exporter</Link> ships those events to a
          destination you own, so you have a record of exactly who was exposed
          to A and who was exposed to B.
        </p>
        <p className="mb-0">
          You wire one exporter per event type. Point them at the same warehouse
          - one table for exposures, one for outcomes - and your experiment data
          lands where your analysts can query it.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="goff-proxy.yaml"
            language="yaml"
            code={EXPORTER_YAML}
            callout="Two BigQuery exporters on the relay proxy: exposures (feature) and outcomes (tracking) into the same dataset."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={EXPORTER_DOCS}>
            Exporter docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Step 3 - tracking */}
      <FeatureRow
        eyebrow="Step 3 · Tracking"
        title="Record what your users did"
        imageSrc={useBaseUrl('/img/landing/ab-testing/schema-measure.svg')}
        imageAlt="Each variation's exposures are joined with its outcomes; a comparison chart highlights the winning variation."
        imageWidth={800}
        imageHeight={600}
        placeholderLabel="Join exposure and outcome per variation to pick a winner">
        <p className="mb-4">
          Exposure alone does not tell you who won - you also need the{' '}
          <em>outcome</em>. The{' '}
          <Link to={TRACKING_DOCS}>OpenFeature Tracking API</Link> lets you
          record a conversion, a purchase amount, or any action against the{' '}
          <strong>same targeting key</strong> you evaluate with. Those tracking
          events flow through the tracking exporter you configured in step 2.
        </p>
        <p className="mb-0">
          Now you can join the two in your analytics: for each variation, how
          many users converted and how much they were worth. That comparison is
          your A/B test result.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="checkout.js"
            language="javascript"
            code={TRACKING_CODE}
            callout="Recorded against the same context as the flag evaluation, so exposure and outcome join on the targeting key."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={TRACKING_DOCS}>
            Tracking API docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      <Cards
        title="Where your experiment data can go"
        cards={DESTINATION_CARDS}
        columns={4}
      />
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to="/docs/tracking/flag-usage-tracking">
            See the full event format <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Why GOFF */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Why run your A/B tests on GO Feature Flag
        </h2>
        <ul className="mx-auto mb-0 max-w-3xl list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>You own the data.</strong> Exposures and outcomes go to your
            warehouse, not a vendor’s - no per-seat experimentation bill, no
            data leaving your stack.
          </li>
          <li>
            <strong>OpenFeature-native.</strong> Assignment and tracking use the
            standard API, so any OpenFeature SDK and the relay proxy work the
            same way.
          </li>
          <li>
            <strong>It composes with your rollouts.</strong> An experiment is
            just one more thing a{' '}
            <Link to="/product/rollouts">feature flag</Link> can do - target a
            segment first, then split within it.
          </li>
          <li>
            <strong>The same recipe compares AI models or prompts.</strong> Swap
            the variations and you are A/B testing{' '}
            <Link to="/product/ai">features for AI</Link> instead of UI.
          </li>
        </ul>
      </section>

      {/* Pitfalls */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          A/B testing pitfalls to avoid
        </h2>
        <ul className="mx-auto mb-0 max-w-3xl list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Re-tuning the percentage mid-test.</strong> Changing the
            split reshuffles some users between A and B. Use an experimentation
            rollout when you need a stable group for the whole window.
          </li>
          <li>
            <strong>Bucketing on an unstable key.</strong> Consistency rides on
            the targeting key; a value that changes per request makes users flip
            variations and pollutes your results.
          </li>
          <li>
            <strong>Exporting exposures but not outcomes.</strong> Knowing who
            saw A or B is only half of it - without tracking events you cannot
            say which variation actually won.
          </li>
          <li>
            <strong>Stopping the moment it looks good.</strong> Peeking at an
            experiment before it reaches significance produces false winners.
            Size the window up front and let it run.
          </li>
        </ul>
      </section>

      <CtaBand
        title="Run your next experiment on your own data"
        description="Self-hosted, OpenFeature-native, MIT-licensed. Split with a YAML file, export to your warehouse, measure the winner - no experimentation SaaS required."
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
