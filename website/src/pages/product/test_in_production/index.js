import React from 'react';
import Layout from '@theme/Layout';
import Head from '@docusaurus/Head';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaRocket,
  FaBook,
  FaGithub,
  FaBug,
  FaShieldAlt,
  FaChartLine,
} from 'react-icons/fa';
import {CodeCard} from '@site/src/components/home/HomepageQuickStart/CodeCard';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';
import rollout from '@site/static/img/landing/rollouts/rollout.png';
import deploy from '@site/static/img/landing/feature-flag/deploy.png';
import targeting from '@site/static/img/landing/rollouts/targeting.png';
import progressive from '@site/static/img/landing/rollouts/progressive.png';
import canary from '@site/static/img/landing/rollouts/canary.png';
import killSwitch from '@site/static/img/landing/feature-flag/kill-switch.png';
import experimentation from '@site/static/img/landing/rollouts/experimentation.png';

const PAGE_TITLE = 'Test in production';
const PAGE_DESCRIPTION =
  'Why testing in production is good practice - and how GO Feature Flag makes it safe: ship dark, dogfood internally, canary, an instant kill switch, and measure with real traffic.';

const ROLLOUT_DOCS = '/docs/configure_flag/rollout-strategies/';
const TARGETING_DOCS = '/docs/configure_flag/target-with-flags';

const SHIP_DARK_YAML = `new-checkout:
  variations:
    on: true
    off: false
  defaultRule:
    # Shipped dark: the code is in production, off for everyone
    # until you decide to release it. No redeploy to flip it.
    variation: off`;

const DOGFOOD_YAML = `new-dashboard:
  variations:
    on: true
    off: false
  targeting:
    # Your team sees it in production first - nobody else does.
    - query: email ew "@yourcompany.com"
      variation: on
  defaultRule:
    variation: off`;

const CANARY_YAML = `new-search:
  variations:
    control: "v1"
    candidate: "v2"
  defaultRule:
    percentage:
      candidate: 1 # start at 1% of real traffic
      control: 99  # widen by hand as the dashboards stay green`;

const KILL_SWITCH_YAML = `flaky-feature:
  variations:
    on: true
    off: false
  defaultRule:
    variation: on
  # The kill switch. Set it and every user falls back to the
  # SDK default on the next poll - no redeploy, no rollback build.
  disable: true`;

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const codeWrap = 'mx-auto w-full max-w-3xl';
const docLink =
  'mt-4 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const CAPABILITY_CARDS = [
  {
    icon: <FaBug />,
    title: 'Catch what staging cannot',
    description:
      'Real data, real scale, real third-party calls. The bugs that only show up under production conditions surface where you can see them.',
  },
  {
    icon: <FaShieldAlt />,
    title: 'Limit the blast radius',
    description:
      'A feature flag in front means only the users you choose - your team, a beta segment, 1% of traffic - ever reach the new path.',
  },
  {
    icon: <FaChartLine />,
    title: 'Measure with real users',
    description:
      'Watch the new code against actual traffic, export the events to your own stack, and decide on evidence instead of a hunch.',
  },
];

const WHEN_ROWS = [
  {
    name: 'Ship dark',
    href: '/product/rollouts',
    when: 'The code is merged and deployed, but you are not ready to release it to anyone yet.',
    instead:
      'You want a specific group to start using it now → Dogfood / Beta.',
  },
  {
    name: 'Dogfood internally',
    href: TARGETING_DOCS,
    when: 'Your own team should hit the new path in production before anyone outside does.',
    instead: 'You need a sample of real, external users → Canary.',
  },
  {
    name: 'Beta / ring',
    href: TARGETING_DOCS,
    when: 'A known segment - opted-in beta users, one region, one plan - should get it next.',
    instead: 'You care about how many users, not which ones → Canary.',
  },
  {
    name: 'Canary',
    href: `${ROLLOUT_DOCS}percentage`,
    when: 'You want a small, random slice of real traffic first and will widen it as it proves out.',
    instead: 'You need to test against specific people → Dogfood / Beta.',
  },
  {
    name: 'Kill switch',
    href: '/docs/configure_flag/create-flags',
    when: 'Something looks wrong and you need every user off the new path now.',
    instead: 'Nothing is broken - you are just ramping up → Canary.',
  },
  {
    name: 'Measure',
    href: '/docs/integrations/export-evaluation-data/',
    when: 'You need to compare the new behavior against the old on real outcomes.',
    instead: 'You only need to ship gradually, not measure → Canary.',
  },
];

const FAQ_ITEMS = [
  {
    question: 'Is testing in production actually safe?',
    answer:
      'It is, with guardrails. The risk is not "production" - it is releasing untested code to everyone at once. Put a feature flag in front, target who reaches the new path, keep a kill switch ready, and watch the results. You expose the change to a slice you control, not the whole user base.',
  },
  {
    question: "Isn't that what a staging environment is for?",
    answer:
      'Staging catches a lot, but it never matches production: the data is smaller and cleaner, the scale is lower, third-party integrations are mocked, and real users do things no test script does. Testing in production complements staging - it is where you find the issues staging structurally cannot reproduce.',
  },
  {
    question: 'How do I control who sees a feature in production?',
    answer:
      'With targeting rules and rollout percentages. A rule can match your own team by email, a beta segment, a region, or a plan; a percentage exposes a random slice of traffic. Everyone else keeps the safe default until you widen the audience.',
  },
  {
    question: 'How fast can I roll back if something breaks?',
    answer:
      'As fast as one config change. Set disable: true (or point the default rule back at the safe variation) and the relay proxy picks it up on its next poll - every user falls back to the SDK default. No rollback build, no redeploy.',
  },
  {
    question: 'Do I need to redeploy to test in production?',
    answer:
      'No. The code ships once, behind a flag. After that you change who sees it - on, off, a percentage, a segment - by editing the flag configuration. The deploy and the release are decoupled.',
  },
  {
    question: 'How do I keep test traffic out of my analytics?',
    answer:
      'Set trackEvents: false on the flag while you are testing, so its evaluations are not exported, then turn it back on when you are ready to measure. GO Feature Flag exports evaluation data to your own stack, so the data never leaves your infrastructure either way.',
  },
];

export default function TestInProductionPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  const siteUrl = siteConfig.url;
  const pageUrl = `${siteUrl}/product/test_in_production`;
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
        'testing in production',
        'test in production',
        'test in production safely',
        'feature flags testing in production',
        'canary testing',
        'progressive delivery',
        'kill switch',
        'dogfooding',
      ]}>
      <Head>
        <link rel="canonical" href={pageUrl} />
        <script type="application/ld+json">
          {JSON.stringify(structuredData)}
        </script>
      </Head>

      <Title
        title={PAGE_TITLE}
        description="The only place that behaves exactly like production is production. Test there on purpose - with real traffic, a flag in front, and a kill switch within reach."
        actions={[
          {
            label: 'Get started',
            href: '/docs/getting-started',
            icon: <FaRocket />,
          },
          {
            label: 'Read the docs',
            href: TARGETING_DOCS,
            icon: <FaBook />,
            variant: 'secondary',
          },
        ]}
      />

      <FeatureRow
        eyebrow="The case for it"
        title="Why test in production?"
        imageSrc={rollout}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="Production has real users, real data and real scale; a staging copy never fully reproduces them, so some issues only appear in production.">
        <p className={prose}>
          Staging is useful, but it is a copy - and a copy is never the
          original. The data is smaller and cleaner, the scale is lower,
          third-party services are mocked, and real users behave in ways no test
          script does. A whole class of bugs only appears under production
          conditions.
        </p>
        <p className="mb-0">
          So the most honest place to validate a change is production itself.
          The catch has always been risk - and that is exactly what a{' '}
          <Link to="/product/what-are-feature-flags">feature flag</Link>{' '}
          removes: you ship the code to production but decide separately who
          actually runs it. Testing in production stops being reckless and
          starts being a discipline.
        </p>
      </FeatureRow>

      <Cards title="Testing in production lets you:" cards={CAPABILITY_CARDS} />

      {/* Safely, not recklessly */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Testing in production safely, not recklessly
        </h2>
        <p className="mx-auto max-w-3xl text-center text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          &ldquo;Test in production&rdquo; is not an excuse to skip the basics.
          It means moving the final validation to the one environment that tells
          the truth - behind four guardrails. A <strong>flag in front</strong>{' '}
          so the change ships dark. <strong>Targeting</strong> so only the users
          you pick reach it. A <strong>kill switch</strong> so any user is one
          config change away from the safe path. And{' '}
          <strong>observability</strong> so you can see what the change is
          doing.
        </p>
        <p className="mx-auto mt-4 max-w-3xl text-center text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          The honest part: you own those guardrails. GO Feature Flag gives you
          all four - self-hosted, OpenFeature-native, configured in a YAML file
          you control - but the discipline of using them is yours. For teams
          that want that control, that is the whole point.
        </p>
      </section>

      {/* a - Ship dark */}
      <FeatureRow
        eyebrow="Ship dark"
        title="Decouple deploy from release"
        imageSrc={deploy}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="Deploying ships the code to production; releasing decides who sees it. A flag sits between the two so code can live in production while staying off.">
        <p className="mb-4">
          Merge it, deploy it, and leave it off. With a flag wrapping the new
          path, the code lives in production - exercised by your CI, your
          startup, your health checks - while no user reaches it. You release it
          later, on your terms, without another deploy.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> always, as the foundation. Every
          technique below starts from a feature that is already in production
          but not yet released.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={SHIP_DARK_YAML}
            callout="The feature is in production but off for everyone. Flip the default rule when you are ready - no redeploy."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to="/product/rollouts">
            See rollout strategies <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* b - Dogfood internally */}
      <FeatureRow
        eyebrow="Dogfood internally"
        title="Let your own team hit it first"
        reverse
        imageSrc={targeting}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="A targeting rule matches your own team by email so they get the new feature in production while everyone else stays on the old path.">
        <p className="mb-4">
          The first real users of a feature should be the people who built it. A
          targeting rule matches your team - by email domain, a staff attribute,
          or an internal segment - so you all run the new path in production
          while every customer stays on the old one.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> the first step after shipping dark -
          shake out the obvious problems against real production before anyone
          outside sees the change.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={DOGFOOD_YAML}
            callout="Anyone with a company email gets the new dashboard; everyone else falls through to the default."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={TARGETING_DOCS}>
            Targeting docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* c - Beta / ring */}
      <FeatureRow
        eyebrow="Beta / ring"
        title="Widen to a trusted segment"
        imageSrc={progressive}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="After internal users, the feature widens to outer rings - opted-in beta users, then one region or plan - before reaching everyone.">
        <p className="mb-4">
          Once your team is happy, widen the circle: opted-in beta users, then
          one region, then one plan. Each ring is just another targeting rule,
          so you grow the audience in deliberate steps and keep the people who
          hit new code people who signed up for it.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> when a known group should get the
          feature next - and you want their feedback before a general release.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={TARGETING_DOCS}>
            Targeting docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* d - Canary */}
      <FeatureRow
        eyebrow="Canary"
        title="Expose a small slice of real traffic"
        reverse
        imageSrc={canary}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="A canary sends a small percentage of real production traffic to the new variation while the rest stays on the control.">
        <p className="mb-4">
          A canary points a small, random percentage of real traffic at the new
          variation - 1%, then 5%, then 25% - while everyone else stays on the
          control. The split is deterministic, so the same users stay in the
          same group until you move the numbers. If the canary is healthy, widen
          it; if not, shrink it back instantly.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> when you need a sample of real,
          external users - not a specific segment - and you will widen by hand
          as your dashboards stay green.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={CANARY_YAML}
            callout="Start at 1% of traffic. Bump candidate to 5, 25, then 100 as it proves out."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={`${ROLLOUT_DOCS}percentage`}>
            Percentage rollout docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* e - Kill switch */}
      <FeatureRow
        eyebrow="Kill switch"
        title="Get everyone off the new path in seconds"
        imageSrc={killSwitch}
        imageWidth={1200}
        imageHeight={747}
        imageAlt="A kill switch disables the flag so every user instantly falls back to the safe default, with no redeploy.">
        <p className="mb-4">
          The safety net that makes all of this safe. If a test in production
          goes wrong, you do not roll back a deploy - you flip the flag. Set{' '}
          <code>disable: true</code> (or point the default rule back at the safe
          variation) and every user falls back to the SDK default on the relay
          proxy&rsquo;s next poll.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> the moment something looks wrong.
          Reach for it first, investigate second - it costs you one config
          change and a few seconds.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={KILL_SWITCH_YAML}
            callout="One line. Every user is back on the safe default within a poll interval - no rollback build."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to="/docs/configure_flag/create-flags">
            Flag configuration docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* f - Measure */}
      <FeatureRow
        eyebrow="Measure"
        title="Watch the change against real traffic"
        reverse
        imageSrc={experimentation}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="Evaluation events from the new variation are exported to your own data stack so you can compare it against the old behavior on real outcomes.">
        <p className="mb-4">
          Testing in production is only worth it if you look at the results. GO
          Feature Flag emits an event for every evaluation and exports them to
          your own stack - S3, Kafka, BigQuery, a file, and more - so you can
          compare the new variation against the old on real outcomes, not a
          hunch. Pair it with an experimentation rollout for a clean measurement
          window.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> whenever the point of the test is to
          decide - keep it, change it, or kill it - based on evidence.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <p className="mx-auto max-w-3xl text-center">
          <Link
            className={docLink}
            to="/docs/integrations/export-evaluation-data/">
            Export evaluation data <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Which technique, when */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Which technique, when?
        </h2>
        <p className={`${prose} text-center`}>
          They build on each other - ship dark first, then widen the audience
          the way that fits the change.
        </p>
        <div className="mx-auto w-full max-w-4xl overflow-x-auto">
          <table className="w-full border-collapse text-left align-top">
            <thead>
              <tr className="border-0 border-b-2 border-solid border-gray-200 dark:border-gray-700">
                <th className="py-3 pr-4 font-bold text-gray-800 dark:text-gray-50">
                  Technique
                </th>
                <th className="py-3 pr-4 font-bold text-gray-800 dark:text-gray-50">
                  Reach for it when
                </th>
                <th className="py-3 font-bold text-gray-800 dark:text-gray-50">
                  Look elsewhere when
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
        <p className="mx-auto mt-6 max-w-4xl text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          These compose. A targeting rule can carry its own percentage - dogfood
          your team at 100% while a beta segment gets a 10% canary - and the
          kill switch sits over all of it.
        </p>
      </section>

      {/* Pitfalls */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Testing-in-production pitfalls to avoid
        </h2>
        <ul className="mx-auto mb-0 max-w-3xl list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>No kill switch.</strong> Testing in production without a
            fast way out is the reckless version. Wire the flag so any user is
            one change away from the safe path before you expose the feature.
          </li>
          <li>
            <strong>Leaking test data into analytics.</strong> Evaluations from
            a half-baked feature can pollute your metrics. Set{' '}
            <code>trackEvents: false</code> while you test, then turn it on when
            you mean to measure.
          </li>
          <li>
            <strong>No safe default.</strong> Always set a{' '}
            <code>defaultRule</code>; it is what users get when no rule matches,
            so make it the safe, known-good value.
          </li>
          <li>
            <strong>Untargeted &ldquo;test&rdquo; flags.</strong> A flag meant
            for your team that has no targeting is just a release to everyone.
            Scope who reaches it before you ship dark.
          </li>
          <li>
            <strong>Mistaking a ramp for a measurement.</strong> Widening a
            canary ships gradually; it does not, on its own, tell you the new
            variation is better. Export the events and compare when that is the
            question.
          </li>
        </ul>
      </section>

      <CtaBand
        title="Test in production with confidence"
        description="Self-hosted, OpenFeature-native, MIT-licensed. Ship dark, target who sees it, and keep a kill switch one YAML change away."
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
