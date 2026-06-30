import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaRocket,
  FaBook,
  FaGithub,
  FaChartLine,
  FaShieldAlt,
  FaClock,
} from 'react-icons/fa';
import {CodeCard} from '@site/src/components/home/HomepageQuickStart/CodeCard';
import SeoHead from '../../../components/landingPage/seo';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';
import rollout from '@site/static/img/landing/rollouts/rollout.png';
import canary from '@site/static/img/landing/rollouts/canary.png';
import progressive from '@site/static/img/landing/rollouts/progressive.png';
import scheduled from '@site/static/img/landing/rollouts/scheduled.png';
import experimentation from '@site/static/img/landing/rollouts/experimentation.png';
import targeting from '@site/static/img/landing/rollouts/targeting.png';

const PAGE_TITLE = 'Feature flag rollouts';
const PAGE_DESCRIPTION =
  'How to roll out a feature safely with GO Feature Flag: percentage, progressive, scheduled, experimentation, and targeting rollouts - and when to use each.';

const ROLLOUT_DOCS = '/docs/configure_flag/rollout-strategies/';

const HOW_YAML = `scream-level:
  variations:
    low: "whisper"
    medium: "talk"
    high: "scream"
  targeting:
    - query: targetingKey eq "12345"
      variation: high
  defaultRule:
    variation: low`;

const PERCENTAGE_YAML = `new-checkout-flow:
  variations:
    control: "v1"
    candidate: "v2"
  defaultRule:
    percentage:
      candidate: 10 # 10% of users get the candidate
      control: 90`;

const PROGRESSIVE_YAML = `progressive-flag:
  variations:
    variationA: A
    variationB: B
  defaultRule:
    progressiveRollout:
      initial:
        variation: variationA
        date: 2024-01-01T05:00:00.100Z
      end:
        variation: variationB
        date: 2024-01-05T05:00:00.100Z`;

const SCHEDULED_YAML = `scheduled-flag:
  variations:
    variationA: A
    variationB: B
  defaultRule:
    percentage:
      variationA: 100
      variationB: 0
  scheduledRollout:
    - date: 2024-04-10T00:00:00.1+02:00
      targeting:
        - name: rule1
          query: beta eq "true"
          percentage:
            variationA: 0
            variationB: 100
    - date: 2024-05-12T15:36:00.1+02:00
      targeting:
        - name: rule1
          query: beta eq "false"`;

const EXPERIMENTATION_YAML = `experimentation-flag:
  variations:
    variationA: A
    variationB: B
  defaultRule:
    percentage:
      variationA: 50
      variationB: 50
  experimentation:
    start: 2024-03-20T00:00:00.1-05:00
    end: 2024-03-27T00:00:00.1-05:00`;

const TARGETING_YAML = `new-dashboard:
  variations:
    control: "old"
    beta: "new"
  targeting:
    - query: plan eq "enterprise"
      variation: beta
    - query: region eq "eu"
      percentage:
        beta: 50
        control: 50
  defaultRule:
    variation: control`;

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const codeWrap = 'mx-auto w-full max-w-3xl';
const docLink =
  'mt-4 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const CAPABILITY_CARDS = [
  {
    icon: <FaChartLine />,
    title: 'Ship gradually, not all at once',
    description:
      'Move from a handful of users to everyone on a curve you control.',
  },
  {
    icon: <FaShieldAlt />,
    title: 'Limit the blast radius',
    description:
      'If a variation misbehaves, only a slice of traffic ever saw it - dial it back in seconds.',
  },
  {
    icon: <FaClock />,
    title: 'Automate or schedule the ramp',
    description:
      'Let the percentage climb on its own, or apply changes on the dates you set.',
  },
];

const WHEN_ROWS = [
  {
    name: 'Percentage',
    href: `${ROLLOUT_DOCS}percentage`,
    when: 'You want a canary: a small slice first, widened by hand as it proves out.',
    instead: 'You would rather the ramp advance on its own → Progressive.',
  },
  {
    name: 'Progressive',
    href: `${ROLLOUT_DOCS}progressive`,
    when: 'You want a hands-off ramp between two dates, no babysitting.',
    instead: 'You need to pause or adjust on judgement mid-ramp → Percentage.',
  },
  {
    name: 'Scheduled',
    href: `${ROLLOUT_DOCS}scheduled`,
    when: 'The launch is tied to dates, or it happens in planned phases.',
    instead:
      'Timing is “when it looks healthy”, not a date → Percentage / Progressive.',
  },
  {
    name: 'Experimentation',
    href: `${ROLLOUT_DOCS}experimentation`,
    when: 'You are measuring variation A against B for a fixed window.',
    instead: 'You just want to ship gradually, not measure → Progressive.',
  },
  {
    name: 'Targeting rules',
    href: '/docs/configure_flag/target-with-flags',
    when: 'Specific users or segments should get a specific variation.',
    instead:
      'The choice is about how many, not who → Percentage / Progressive.',
  },
];

const FAQ_ITEMS = [
  {
    question: 'What is the difference between a rollout and a release?',
    answer:
      'A release is making a feature available to users. A rollout is how you get there - which users see which variation, in what order, over what time. The flag controls the rollout, so you change it without a redeploy.',
  },
  {
    question: 'Is a percentage rollout the same as a canary release?',
    answer:
      'Yes. A percentage rollout sends a small slice of traffic to a new variation first - that is a canary. You widen the percentage as it proves out, and shrink it back instantly if it does not.',
  },
  {
    question: 'How does GO Feature Flag decide which users are in a rollout?',
    answer:
      'It hashes the evaluation context’s targeting key deterministically. The same user keeps the same variation as long as the percentages do not change, so experiences stay consistent across requests.',
  },
  {
    question: 'Can a rollout advance automatically over time?',
    answer:
      'Yes. A progressive rollout ramps the percentage from an initial value to an end value between two dates, with no manual step in between.',
  },
  {
    question: 'Can I schedule a rollout for a specific date?',
    answer:
      'Yes. A scheduled rollout applies changes - new targeting rules, different percentages - at the dates you set, in as many stages as you need.',
  },
  {
    question: 'Do I need to redeploy to change a rollout?',
    answer:
      'No. Rollouts live in the flag configuration file. Edit it and the relay proxy picks up the change on its next poll - no redeploy, no restart.',
  },
];

export default function RolloutsPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  return (
    <Layout
      title={PAGE_TITLE}
      description={PAGE_DESCRIPTION}
      keywords={[
        'feature flag rollout',
        'feature rollout',
        'progressive rollout',
        'percentage rollout',
        'canary release',
        'scheduled rollout',
        'A/B test rollout',
        'rollout strategies',
      ]}>
      <SeoHead
        title={PAGE_TITLE}
        description={PAGE_DESCRIPTION}
        path="/product/rollouts"
        image="/img/logo/x-card.png"
      />

      <Title
        title={PAGE_TITLE}
        description="Move a feature from nobody to everybody on your terms - a slice of users, a steady ramp, a scheduled date, or a specific segment."
        actions={[
          {
            label: 'Get started',
            href: '/docs/getting-started',
            icon: <FaRocket />,
          },
          {
            label: 'Rollout docs',
            href: ROLLOUT_DOCS,
            icon: <FaBook />,
            variant: 'secondary',
          },
        ]}
      />

      <FeatureRow
        eyebrow="The core idea"
        title="What is a rollout?"
        imageSrc={rollout}
        imageAlt="A rollout is how you take a feature from nobody to everybody - deciding which users get it and when, separately from when you deploy the code."
        placeholderLabel="Prompt A - rollout ramp (0% → 100% across a growing crowd)">
        <p className={prose}>
          A rollout is how you take a feature from nobody to everybody -
          deciding <strong>which users</strong> get it and <strong>when</strong>
          , separately from when you deploy the code.
        </p>
        <p className="mb-0">
          With a <Link to="/product/what-are-feature-flags">feature flag</Link>{' '}
          in place, the new code ships dark. A rollout then steers traffic
          toward it on a curve you control - a few users, then a percentage,
          then everyone - and lets you reverse course the moment something looks
          wrong.
        </p>
      </FeatureRow>

      <Cards title="Rollouts let you:" cards={CAPABILITY_CARDS} />

      {/* How rollouts work */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          How rollouts work in GO Feature Flag
        </h2>
        <p className={prose}>
          Every flag has named <strong>variations</strong> - and they are not
          limited to on/off. A variation can be a boolean, a string, a number,
          or a JSON object, and a flag can have more than two of them. A rollout
          decides which variation each user gets, and how that split changes
          over time or by audience.
        </p>
        <p className={prose}>
          You configure it in the flag&rsquo;s YAML - inside the{' '}
          <code>defaultRule</code> or a <code>targeting</code> rule. Evaluation
          is deterministic on the targeting key, so a user stays in the same
          bucket between requests. Change the file and the relay proxy picks it
          up on its next poll - no redeploy.
        </p>
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={HOW_YAML}
            callout="Three string variations, not a boolean. User 12345 always gets high; everyone else falls through to the default."
          />
        </div>
      </section>

      {/* Rollout vs deploy vs progressive delivery */}
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Rollout, deploy, and progressive delivery
        </h2>
        <p className="mx-auto max-w-3xl text-center text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          These get used interchangeably, so to be precise: a{' '}
          <strong>deploy</strong> ships code to your servers. A{' '}
          <strong>rollout</strong> decides who actually sees the feature
          afterward — that&rsquo;s the flag&rsquo;s job, and it needs no
          redeploy. <strong>Progressive delivery</strong> is the umbrella
          practice: deploy continuously, then release gradually behind flags.
          The rollout is the lever; progressive delivery is the habit of pulling
          it carefully.
        </p>
      </section>

      {/* Percentage */}
      <FeatureRow
        eyebrow="Percentage rollout"
        title="Send a slice of traffic to a variation"
        imageSrc={canary}
        imageAlt="A percentage rollout splits traffic across variations by share - say, 10% to the candidate and 90% to the control."
        reverse
        placeholderLabel="Prompt B - percentage / canary (a small highlighted slice of a user crowd)">
        <p className="mb-4">
          A percentage rollout splits traffic across variations by share - say,
          10% to the candidate and 90% to the control. The split is
          deterministic, so the same users stay in the same group until you move
          the numbers.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> a canary - you want a small group
          first and you will widen it by hand as your dashboards stay green.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={PERCENTAGE_YAML}
            callout="Bump candidate to 25, then 100, as it proves out. No redeploy."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={`${ROLLOUT_DOCS}percentage`}>
            Percentage rollout docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Progressive */}
      <FeatureRow
        eyebrow="Progressive rollout"
        title="Ramp the percentage automatically"
        imageSrc={progressive}
        imageAlt="A progressive rollout moves the share from an initial value to an end value between two dates, on its own. You set the start and finish; GO Feature Flag interpolates the rest."
        placeholderLabel="Prompt C - progressive (a line ramping up along a time axis between two date markers)">
        <p className="mb-4">
          A progressive rollout moves the share from an initial value to an end
          value between two dates, on its own. You set the start and finish; GO
          Feature Flag interpolates the rest.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> a hands-off ramp over hours or days
          when you are confident enough to let it advance without a person
          nudging it.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={PROGRESSIVE_YAML}
            callout="From variationA to variationB over four days - add initial/end percentages to ramp partially."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={`${ROLLOUT_DOCS}progressive`}>
            Progressive rollout docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Scheduled */}
      <FeatureRow
        eyebrow="Scheduled rollout"
        title="Change the flag on set dates"
        reverse
        imageSrc={scheduled}
        imageAlt="A scheduled rollout applies changes at specific dates, in as many stages as you need - flip on a targeting rule today, change the split next month. Each step can edit any part of the flag."
        placeholderLabel="Prompt D - scheduled (a horizontal timeline with milestone markers on dates)">
        <p className="mb-4">
          A scheduled rollout applies changes at specific dates, in as many
          stages as you need - flip on a targeting rule today, change the split
          next month. Each step can edit any part of the flag.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> a launch tied to a calendar - a
          marketing date, a maintenance window, or a phased plan you want locked
          in ahead of time.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={SCHEDULED_YAML}
            callout="Two stages: open to beta users on the first date, then to everyone else on the second."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={`${ROLLOUT_DOCS}scheduled`}>
            Scheduled rollout docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Experimentation */}
      <FeatureRow
        eyebrow="Experimentation rollout"
        title="Run a variation for a fixed window"
        imageSrc={experimentation}
        imageAlt="An experimentation rollout makes the flag active only between a start and end time. Inside the window it serves your split; outside it, everyone gets the default. Pair it with the data export to measure which variation won."
        placeholderLabel="Prompt E - experimentation (an A vs B split inside a bracketed start–end window)">
        <p className="mb-4">
          An experimentation rollout makes the flag active only between a start
          and end time. Inside the window it serves your split; outside it,
          everyone gets the default. Pair it with the data export to measure
          which variation won.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> a time-boxed A/B test - you want a
          clean measurement period, then an automatic return to the default. The
          same approach lets you compare two models or prompts with{' '}
          <Link to="/product/ai">feature flags for AI</Link>.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={EXPERIMENTATION_YAML}
            callout="A 50/50 split that only runs for one week, then falls back to the default."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to={`${ROLLOUT_DOCS}experimentation`}>
            Experimentation docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Targeting */}
      <FeatureRow
        eyebrow="Targeting rules"
        title="Give specific people a specific variation"
        imageSrc={targeting}
        imageAlt="Targeting rules answer who, not just how many. Each rule matches users with a query - by plan, region, role, or any attribute - and serves them a variation. Rules run in order; the first match wins, and a rule can itself carry a percentage."
        reverse
        placeholderLabel="Prompt F - targeting (distinct user segments routed to different variations)">
        <p className="mb-4">
          Targeting rules answer <em>who</em>, not just how many. Each rule
          matches users with a query - by plan, region, role, or any attribute -
          and serves them a variation. Rules run in order; the first match wins,
          and a rule can itself carry a percentage.
        </p>
        <p className="mb-0">
          <strong>When to use it:</strong> a segment should get a particular
          variation - enterprise plans, internal staff, one region - or you want
          to combine <em>who</em> with <em>how many</em>.
        </p>
      </FeatureRow>
      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <div className={codeWrap}>
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={TARGETING_YAML}
            callout="Enterprise plans get the beta; the EU gets a 50/50 split; everyone else stays on control."
          />
        </div>
        <p className="mx-auto max-w-3xl text-center">
          <Link className={docLink} to="/docs/configure_flag/target-with-flags">
            Targeting docs <span aria-hidden="true">→</span>
          </Link>
        </p>
      </section>

      {/* Which to use */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Which rollout should you use?
        </h2>
        <p className={`${prose} text-center`}>
          Pick by what you are optimizing for - speed of feedback, automation,
          timing, measurement, or audience.
        </p>
        <div className="mx-auto w-full max-w-4xl overflow-x-auto">
          <table className="w-full border-collapse text-left align-top">
            <thead>
              <tr className="border-0 border-b-2 border-solid border-gray-200 dark:border-gray-700">
                <th className="py-3 pr-4 font-bold text-gray-800 dark:text-gray-50">
                  Strategy
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
          These compose. A targeting rule can itself carry a percentage or a
          progressive rollout - target the segment first, then ramp within it.
        </p>
      </section>

      {/* Pitfalls */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Rollout pitfalls to avoid
        </h2>
        <ul className="mx-auto mb-0 max-w-3xl list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Leaving finished rollouts in the code.</strong> Once a flag
            sits at 100% for everyone, it&rsquo;s flag debt — clean it up.
          </li>
          <li>
            <strong>No safe default.</strong> Always set a{' '}
            <code>defaultRule</code>; it&rsquo;s the value users get when no
            rule matches, so make it the safe one.
          </li>
          <li>
            <strong>Re-tuning a percentage reshuffles some users.</strong>{' '}
            Changing the split can move users between variations — fine for a
            ramp, but use an experimentation rollout when you need a stable A/B
            group.
          </li>
          <li>
            <strong>Bucketing on an unstable key.</strong> Consistency rides on
            the targeting key; a value that changes per request makes users flip
            between variations.
          </li>
          <li>
            <strong>Mistaking a ramp for a measurement.</strong> Percentage and
            progressive rollouts ship gradually; reach for experimentation when
            you actually need to compare outcomes.
          </li>
        </ul>
      </section>

      <CtaBand
        title="Roll out your next feature safely"
        description="Self-hosted, OpenFeature-native, MIT-licensed. Change a YAML file and the rollout updates - no redeploy."
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
