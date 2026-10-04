import React from 'react';
import PropTypes from 'prop-types';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaRocket,
  FaBook,
  FaGithub,
  FaDownload,
  FaUpload,
  FaBell,
} from 'react-icons/fa';
import {SocialIcon} from '@site/src/components/home/features';
import {integrations} from '@site/data/integrations';
import SeoHead from '../../../components/landingPage/seo';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';
import useBaseUrl from '@docusaurus/useBaseUrl';

const PAGE_TITLE = 'Integrations';
const PAGE_DESCRIPTION =
  'GO Feature Flag is cloud-native and plugs into your whole stack out of the box - retrieve flag config from S3, GCS, GitHub, Kubernetes, MongoDB, Redis and more; export evaluation data to Kafka, BigQuery, PubSub, OpenTelemetry and more; get notified on Slack, Discord or Microsoft Teams.';

const RETRIEVER_DOCS = '/docs/concepts/retriever';
const EXPORTER_DOCS = '/docs/concepts/exporter';
const NOTIFIER_DOCS = '/docs/concepts/notifier';

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const docLink =
  'mt-6 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const EXTENSION_CARDS = [
  {
    icon: <FaDownload />,
    title: 'Retrievers',
    description:
      'Load your flag configuration from wherever it already lives - object storage, a Git provider, a database, or a Kubernetes ConfigMap.',
    link: RETRIEVER_DOCS,
    linkLabel: 'Retriever concept',
  },
  {
    icon: <FaUpload />,
    title: 'Exporters',
    description:
      'Stream every flag evaluation to your data and analytics stack so you can measure rollouts, run experiments, and audit usage.',
    link: EXPORTER_DOCS,
    linkLabel: 'Exporter concept',
  },
  {
    icon: <FaBell />,
    title: 'Notifiers',
    description:
      'Get told the moment a flag configuration changes, on the channel your team already watches.',
    link: NOTIFIER_DOCS,
    linkLabel: 'Notifier concept',
  },
];

// Per-category logo grid, reusing the home page's circular badge component and
// the canonical integrations data so this page stays in sync automatically.
function LogoGrid({items}) {
  return (
    <div className="grid grid-cols-3 justify-items-center gap-y-2 sm:grid-cols-4 md:grid-cols-6">
      {items.map(item => (
        <SocialIcon
          key={item.name}
          backgroundColor={item.bgColor}
          fontAwesomeIcon={item.faLogo}
          img={item.logo}
          tooltipText={item.name}
          colorClassName=""
        />
      ))}
    </div>
  );
}

LogoGrid.propTypes = {
  items: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      bgColor: PropTypes.string,
      faLogo: PropTypes.string,
      logo: PropTypes.string,
    })
  ).isRequired,
};

const FAQ_ITEMS = [
  {
    question: 'What can GO Feature Flag retrieve flags from?',
    answer: (
      <>
        A lot of places, with no plugin to install. Built-in{' '}
        <Link to={RETRIEVER_DOCS}>retrievers</Link> cover HTTP(S), the local
        file system, Kubernetes ConfigMaps, AWS S3, Google Cloud Storage, Azure
        Blob Storage, GitHub, GitLab, Bitbucket, MongoDB, Redis and PostgreSQL.
        You point GO Feature Flag at wherever your configuration already lives.
      </>
    ),
  },
  {
    question: 'Where can I export evaluation data?',
    answer: (
      <>
        To your existing data stack. Built-in{' '}
        <Link to={EXPORTER_DOCS}>exporters</Link> send evaluation events to AWS
        S3, Azure Blob Storage, the file system, Apache Kafka, AWS Kinesis,
        Google Cloud Storage, Google Cloud PubSub, Google Cloud BigQuery, AWS
        SQS, a webhook, your application logs, or OpenTelemetry.
      </>
    ),
  },
  {
    question: 'How do I get notified when a flag changes?',
    answer: (
      <>
        Configure a <Link to={NOTIFIER_DOCS}>notifier</Link> and GO Feature Flag
        tells you the moment a flag configuration changes - on Slack, Discord,
        Microsoft Teams, a webhook, or in your application logs.
      </>
    ),
  },
  {
    question: 'Do integrations need plugins or extra services?',
    answer:
      'No. Retrievers, exporters and notifiers are built into GO Feature Flag - you turn them on with a few lines of configuration. There is no marketplace to browse and nothing extra to deploy.',
  },
  {
    question: 'Can I add an integration that is not listed?',
    answer: (
      <>
        Yes. Each extension point is a small Go interface -{' '}
        <Link to={RETRIEVER_DOCS}>Retriever</Link>,{' '}
        <Link to={EXPORTER_DOCS}>Exporter</Link> and{' '}
        <Link to={NOTIFIER_DOCS}>Notifier</Link>. Implement the one you need and
        GO Feature Flag uses it like any built-in integration.
      </>
    ),
  },
];

export default function IntegrationsPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  return (
    <Layout
      title={PAGE_TITLE}
      description={PAGE_DESCRIPTION}
      keywords={[
        'feature flag integrations',
        'cloud native feature flags',
        'feature flag retriever',
        'feature flag exporter',
        'feature flag notifier',
        'S3 feature flags',
        'Kafka feature flags',
        'OpenTelemetry feature flags',
        'Slack feature flag notification',
      ]}>
      <SeoHead
        title={PAGE_TITLE}
        description={PAGE_DESCRIPTION}
        path="/product/integrations"
      />

      <Title
        title="Integrates with your whole stack"
        description="GO Feature Flag is cloud-native by design. It reads its configuration from the tools you already run, ships evaluation data to your data stack, and notifies your team on change - all built in, with no plugins to install."
        actions={[
          {
            label: 'Get started',
            href: '/docs/getting-started',
            icon: <FaRocket />,
          },
          {
            label: 'Browse the docs',
            href: '/docs/integrations/store-flags-configuration',
            icon: <FaBook />,
            variant: 'secondary',
          },
        ]}
      />

      <FeatureRow
        eyebrow="The core idea"
        title="Three pluggable extension points"
        imageSrc={useBaseUrl('/docs/openfeature/architecture.svg')}
        imageAlt="GO Feature Flag architecture: OpenFeature SDKs talking to the relay-proxy, which loads flag configuration via retrievers and emits events via notifiers and exporters."
        imageWidth={1200}
        imageHeight={896}
        placeholderLabel="Retrievers → GO Feature Flag → Exporters & Notifiers">
        <p className={prose}>
          GO Feature Flag is built around three small, well-defined extension
          points - and every integration is one of them:
        </p>
        <ul className="mb-0 list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            A <Link to={RETRIEVER_DOCS}>retriever</Link> fetches your flag
            configuration from where it lives and polls it for changes.
          </li>
          <li>
            An <Link to={EXPORTER_DOCS}>exporter</Link> sends every flag
            evaluation to your data and analytics stack.
          </li>
          <li>
            A <Link to={NOTIFIER_DOCS}>notifier</Link> tells your team the
            moment a flag configuration changes.
          </li>
        </ul>
      </FeatureRow>

      <Cards
        title="Pick the integrations you need"
        columns={3}
        cards={EXTENSION_CARDS}
      />

      {/* Retrievers */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Retrieve your configuration from anywhere
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          Keep your flags where they already live - object storage, a Git
          provider, a database, or a Kubernetes ConfigMap. GO Feature Flag reads
          it and polls for changes.
        </p>
        <LogoGrid items={integrations.retrievers} />
        <div className="text-center">
          <Link className={docLink} to={RETRIEVER_DOCS}>
            Learn about retrievers <span aria-hidden="true">→</span>
          </Link>
        </div>
      </section>

      {/* Exporters */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Export evaluation data to your data stack
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          Every evaluation can be streamed out to measure rollouts, run
          experiments, and audit usage - synchronously or in batches, to the
          destination you already use.
        </p>
        <LogoGrid items={integrations.exporters} />
        <div className="text-center">
          <Link className={docLink} to={EXPORTER_DOCS}>
            Learn about exporters <span aria-hidden="true">→</span>
          </Link>
        </div>
      </section>

      {/* Notifiers */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Get notified when a flag changes
        </h2>
        <p className={`${prose} mx-auto max-w-3xl text-center`}>
          The moment a flag configuration changes, GO Feature Flag tells the
          channel your team already watches.
        </p>
        <LogoGrid items={integrations.notifiers} />
        <div className="text-center">
          <Link className={docLink} to={NOTIFIER_DOCS}>
            Learn about notifiers <span aria-hidden="true">→</span>
          </Link>
        </div>
      </section>

      <CtaBand
        title="Don't see your tool?"
        description="Retrievers, exporters and notifiers are small Go interfaces - implement the one you need and it works like any built-in integration. Self-hosted, OpenFeature-native, MIT-licensed."
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
