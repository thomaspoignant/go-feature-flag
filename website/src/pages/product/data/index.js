import React from 'react';
import PropTypes from 'prop-types';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaRocket,
  FaBook,
  FaGithub,
  FaChartLine,
  FaBullseye,
  FaShareAlt,
  FaExchangeAlt,
} from 'react-icons/fa';
import {CodeCard} from '@site/src/components/home/HomepageQuickStart/CodeCard';
import SeoHead from '../../../components/landingPage/seo';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';
import {
  CoreIdeaIllustration,
  ProviderEventsIllustration,
  TrackJoinIllustration,
} from './_illustrations';

const PAGE_TITLE = 'Route your feature flag data anywhere';
const PAGE_DESCRIPTION =
  'GO Feature Flag does not compute or analyze your data - it captures every flag evaluation and custom outcome and delivers it to your own data stack through 10+ export integrations.';

const EXPORT_DOCS = '/docs/integrations/export-evaluation-data';
const EXPORTER_CONCEPT_DOCS = '/docs/concepts/exporter';
const FLAG_USAGE_DOCS = '/docs/tracking/flag-usage-tracking';
const TRACK_API_DOCS = '/docs/tracking/tracking-api';

const FLAG_USAGE_YAML = `new-checkout:
  variations:
    on: true
    off: false
  defaultRule:
    percentage:
      on: 50
      off: 50
  # trackEvents is true by default - every evaluation
  # becomes a "feature" event sent to your exporters.
  trackEvents: true`;

const TRACK_API_JS = `// Record a business outcome from your app, using the
// same evaluation context you use for flag evaluation.
client.track('checkout-completed', evaluationContext, {
  value: 99.99,
  currency: 'USD',
});
// GO Feature Flag forwards it to your tracking exporters -
// join it with the flag exposure to measure the variation.`;

const TRACKING_EXPORTER_YAML = `# Send tracking events to their own destination by
# setting eventType on a second exporter entry.
exporters:
  - kind: webhook                 # feature events (flag usage)
    endpointUrl: https://my-collector.example.com/feature
  - kind: webhook
    eventType: tracking           # custom Track() outcomes
    endpointUrl: https://my-collector.example.com/tracking`;

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const codeWrap = 'mx-auto w-full max-w-3xl';
const docLink =
  'mt-4 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const PATTERN_CARDS = [
  {
    icon: <FaChartLine />,
    title: 'Evaluation tracking, built in',
    description:
      'Every flag evaluation is captured automatically by the providers and the relay proxy. No code to write - just keep trackEvents on and point it at an exporter.',
    link: FLAG_USAGE_DOCS,
    linkLabel: 'Flag usage tracking',
  },
  {
    icon: <FaBullseye />,
    title: 'Track API for outcomes',
    description:
      'Call client.track() to record what users actually did - a checkout, a click, revenue. The outcome half of experimentation, joined to the flag they saw.',
    link: TRACK_API_DOCS,
    linkLabel: 'Tracking API',
  },
  {
    icon: <FaShareAlt />,
    title: 'Exported everywhere',
    description:
      'Both event streams flow through the exporters you configure to your warehouse, queue, lake, or observability stack - in batches or in real time.',
    link: EXPORTER_CONCEPT_DOCS,
    linkLabel: 'Exporter concept',
  },
];

// Async exporters batch events and write them to storage; sync exporters
// stream each event straight to a queue or endpoint.
const STREAMING_EXPORTERS = [
  {name: 'Apache Kafka', slug: 'kafka'},
  {name: 'Webhook', slug: 'webhook'},
  {name: 'AWS SQS', slug: 'aws-sqs'},
  {name: 'AWS Kinesis', slug: 'aws-kinesis'},
  {name: 'Google Cloud Pub/Sub', slug: 'google-cloud-pubsub'},
  {name: 'OpenTelemetry', slug: 'opentelemetry'},
  {name: 'Log', slug: 'log'},
];

const STORAGE_EXPORTERS = [
  {name: 'AWS S3', slug: 'aws-s3'},
  {name: 'Google Cloud Storage', slug: 'google-cloud-storage'},
  {name: 'Azure Blob Storage', slug: 'azure-blob-storage'},
  {name: 'File', slug: 'file'},
];

const FAQ_ITEMS = [
  {
    question: 'Does GO Feature Flag store or analyze my data?',
    answer:
      'No. GO Feature Flag does not compute analytics or hold a dashboard of its own. It records a lightweight event for each evaluation and each tracked outcome, then delivers those events to the destinations you configure. The analysis happens in your stack, on your terms.',
  },
  {
    question:
      'What is the difference between flag-usage events and tracking events?',
    answer: (
      <>
        A <strong>flag-usage</strong> (or &ldquo;feature&rdquo;) event records
        that a user received a given variation - it is the <em>exposure</em>. A{' '}
        <strong>tracking</strong> event records a custom action or outcome you
        send with the <Link to={TRACK_API_DOCS}>OpenFeature Track API</Link> -
        it is the <em>result</em>. Join the two and you can measure which
        variation drove a conversion.
      </>
    ),
  },
  {
    question: 'Do I need to change my code to collect evaluation data?',
    answer: (
      <>
        No. Flag-usage tracking is built into the providers and the relay proxy.
        Keep <code>trackEvents: true</code> (the default) on your flags and
        configure at least one exporter - evaluations start flowing with no
        application code. The Track API is the only part that needs a call, and
        only because it records outcomes your app knows about.
      </>
    ),
  },
  {
    question: 'Can I send data to more than one destination?',
    answer:
      'Yes. Configure multiple exporters and each event is delivered to all of them. You can also route feature events and tracking events to different backends by setting eventType on separate exporter entries.',
  },
  {
    question: 'How do I correlate flag exposures with conversions?',
    answer: (
      <>
        Use the same evaluation context (the targeting key and attributes) for
        both the flag evaluation and your <code>client.track()</code> calls.
        Both event types carry that key, so in your warehouse you can join
        exposures to outcomes and measure each variation. See the{' '}
        <Link to={TRACK_API_DOCS}>Tracking API docs</Link>.
      </>
    ),
  },
];

function ExporterList({title, exporters}) {
  return (
    <div>
      <h3 className="mb-4 text-xl font-bold text-gray-800 dark:text-gray-50">
        {title}
      </h3>
      <ul className="m-0 grid list-none grid-cols-1 gap-3 p-0 sm:grid-cols-2">
        {exporters.map(exporter => (
          <li key={exporter.slug}>
            <Link
              className="flex items-center justify-between rounded-lg border border-solid border-gray-200 px-4 py-3 font-semibold text-gray-800 no-underline transition-shadow hover:no-underline hover:shadow-md dark:border-gray-700 dark:text-gray-50"
              to={`${EXPORT_DOCS}/${exporter.slug}`}>
              <span>{exporter.name}</span>
              <span aria-hidden="true">→</span>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

ExporterList.propTypes = {
  title: PropTypes.node.isRequired,
  exporters: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      slug: PropTypes.string.isRequired,
    })
  ).isRequired,
};

export default function DataPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  return (
    <Layout
      title={PAGE_TITLE}
      description={PAGE_DESCRIPTION}
      keywords={[
        'feature flag data export',
        'flag evaluation data',
        'feature flag analytics',
        'OpenFeature tracking',
        'flag usage tracking',
        'experimentation data',
        'export evaluation data',
        'feature flag exporter',
      ]}>
      <SeoHead
        title={PAGE_TITLE}
        description={PAGE_DESCRIPTION}
        path="/product/data"
        image="/img/logo/x-card.png"
        imageMeta
      />

      <Title
        title={PAGE_TITLE}
        description="GO Feature Flag doesn't compute your data - it delivers it. Every flag evaluation and every outcome you record is routed straight to your own data stack, so analytics, experimentation, and auditing stay where you already run them."
        actions={[
          {
            label: 'Export your data',
            href: EXPORT_DOCS,
            icon: <FaRocket />,
          },
          {
            label: 'Read the docs',
            href: EXPORTER_CONCEPT_DOCS,
            icon: <FaBook />,
            variant: 'secondary',
          },
        ]}
      />

      <FeatureRow
        eyebrow="The core idea"
        title="A router for your flag data, not an analytics engine"
        media={<CoreIdeaIllustration />}>
        <p className={prose}>
          GO Feature Flag deliberately doesn&rsquo;t compute, aggregate, or
          analyze anything. It captures a lightweight event for every flag
          evaluation - and any outcome you choose to record - and forwards it,
          untouched, to the destinations you configure.
        </p>
        <p className="mb-0">
          That keeps your data <strong>yours</strong>: it lands in your
          warehouse, your queue, or your observability stack, where you already
          do the computing. Two patterns feed that pipeline -{' '}
          <strong>evaluation tracking</strong> built into the providers, and the{' '}
          <strong>OpenFeature Track API</strong> for custom outcomes.
        </p>
      </FeatureRow>

      <Cards
        title="Two ways data is collected, one way out"
        cards={PATTERN_CARDS}
      />

      {/* Pattern 1 - evaluation tracking */}
      <FeatureRow
        eyebrow="Pattern 1 - built into the providers"
        title="Evaluation tracking, with no extra code"
        media={<ProviderEventsIllustration />}>
        <p className="mb-4">
          Every time a flag is evaluated, a <strong>feature event</strong> - who
          saw which variation, and when - is captured automatically. The
          OpenFeature provider&rsquo;s data collector buffers these events and
          the relay proxy ships them to your exporters.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> keep <code>trackEvents: true</code>{' '}
          (the default) on the flags you care about and configure an exporter.
          No application code changes - the exposure data just starts flowing.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={FLAG_USAGE_YAML}
            callout="trackEvents is on by default. Add an exporter and every evaluation is delivered as a feature event."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={FLAG_USAGE_DOCS}>
            Flag usage tracking docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Pattern 2 - track API */}
      <FeatureRow
        eyebrow="Pattern 2 - the OpenFeature Track API"
        title="Record outcomes to close the experimentation loop"
        reverse
        media={<TrackJoinIllustration />}>
        <p className="mb-4">
          Exposures tell you <em>who saw what</em>; outcomes tell you{' '}
          <em>what happened next</em>. The standard{' '}
          <Link to="/product/open-feature">OpenFeature</Link> Track API lets you
          record a named action - a checkout, a click, revenue - against the
          same context you evaluate flags with.
        </p>
        <p className="mb-0">
          <strong>How to use it:</strong> call <code>client.track(...)</code>{' '}
          where the action happens, add a tracking exporter, and join exposures
          to outcomes in your stack to measure which variation actually won.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="track.js"
            language="javascript"
            code={TRACK_API_JS}
            callout="The same evaluation context links the outcome back to the variation the user was exposed to."
          />
          <div className="mt-6">
            <CodeCard
              filename="goff-proxy.yaml"
              language="yaml"
              code={TRACKING_EXPORTER_YAML}
              callout="Route tracking events to their own destination with eventType: tracking."
            />
          </div>
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={TRACK_API_DOCS}>
            Tracking API docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Destinations */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Deliver it wherever you need it
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          Both event streams flow through the same exporters. Stream each event
          straight to a queue or endpoint, or batch them and write to storage -
          mix and match as many destinations as you like.
        </p>
        <div className="mx-auto grid w-full max-w-4xl grid-cols-1 gap-10 md:grid-cols-2">
          <ExporterList
            title="Stream in real time"
            exporters={STREAMING_EXPORTERS}
          />
          <ExporterList
            title="Batch to storage"
            exporters={STORAGE_EXPORTERS}
          />
        </div>
        <p className="mx-auto mt-8 max-w-3xl text-center">
          <Link className={docLink} to={EXPORT_DOCS}>
            All export integrations <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      <CtaBand
        title="Own your feature flag data"
        description="Self-hosted, OpenFeature-native, MIT-licensed. Capture every evaluation and outcome, then route it straight to the stack you already trust - no black box in between."
        actions={[
          {
            label: 'Get started',
            href: '/docs/getting-started',
            icon: <FaExchangeAlt />,
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
