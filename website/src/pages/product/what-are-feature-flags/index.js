import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import {
  FaToggleOn,
  FaEye,
  FaSlidersH,
  FaRocket,
  FaBook,
  FaGithub,
  FaFlask,
  FaUserShield,
  FaPowerOff,
} from 'react-icons/fa';
import {CodeCard} from '@site/src/components/home/HomepageQuickStart/CodeCard';
import SeoHead from '../../../components/landingPage/seo';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';
import evaluationFlow from '@site/static/img/landing/feature-flag/multi-ff.png';
import deployVsRelease from '@site/static/img/landing/feature-flag/deploy.png';
import killswitch from '@site/static/img/landing/feature-flag/kill-switch.png';
import gofflogo from '@site/static/img/logo/logo.png';

const PAGE_TITLE = 'What are feature flags?';
const PAGE_DESCRIPTION =
  "A developer's guide to feature flags (feature toggles): what they are, how they work, what they're for, and how to self-host them for free.";

const TARGETING_YAML = `show-email-contact:
  variations:
    enabled: true
    disabled: false
  targeting:
    - query: targetingKey eq "1"
      variation: enabled
  defaultRule:
    variation: disabled`;

const PERCENTAGE_YAML = `new-checkout-flow:
  variations:
    new: true
    old: false
  defaultRule:
    percentage:
      new: 5 # start at 5%, widen as the metrics hold
      old: 95`;

const GO_SNIPPET = `// Connect to your relay proxy through the OpenFeature SDK.
provider, _ := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
  Endpoint: "http://localhost:1031",
})
_ = of.SetProviderAndWait(provider)
client := of.NewClient("my-app")

// Evaluate the flag for one user. You always get a value back.
evalCtx := of.NewEvaluationContext("1", map[string]any{})
if client.Boolean(ctx, "show-email-contact", false, evalCtx) {
  // the feature is on for this user
}`;

const JS_SNIPPET = `const {OpenFeature} = require("@openfeature/server-sdk");
const {GoFeatureFlagProvider} = require("@openfeature/go-feature-flag-provider");

OpenFeature.setProvider(new GoFeatureFlagProvider({
  endpoint: "http://localhost:1031/",
}));
const client = OpenFeature.getClient("my-app");

// targetingKey is the unique identifier we bucket on.
const ctx = {targetingKey: "1", admin: true};
const show = await client.getBooleanValue("show-email-contact", false, ctx);`;

const DOCKER_SNIPPET = `docker run \\
  -p 1031:1031 \\
  -v $(pwd)/flags.goff.yaml:/goff/flags.goff.yaml \\
  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml \\
  gofeatureflag/go-feature-flag:latest`;

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';

const CAPABILITY_CARDS = [
  {
    icon: <FaToggleOn />,
    title: 'Activate or deactivate specific application functionality',
  },
  {icon: <FaEye />, title: 'Control the visibility of certain features'},
  {
    icon: <FaSlidersH />,
    title: 'Conditionally change feature behavior in real time',
  },
];

const FLAG_TYPE_CARDS = [
  {
    icon: <FaRocket />,
    title: 'Release flags',
    description:
      'Roll a new feature out gradually and turn it off instantly if it misbehaves.',
  },
  {
    icon: <FaFlask />,
    title: 'Experiment flags',
    description:
      'Serve different variations to measure which one performs best — A/B tests and beyond.',
  },
  {
    icon: <FaUserShield />,
    title: 'Permission flags',
    description:
      'Show a feature to specific people: beta testers, internal staff, or a paid tier.',
  },
  {
    icon: <FaPowerOff />,
    title: 'Operational flags',
    description:
      'Kill switches and config you can flip in production without shipping a release.',
  },
];

const FAQ_ITEMS = [
  {
    question: "Feature flag vs feature toggle — what's the difference?",
    answer:
      'Nothing. They’re two names for the same thing. "Feature toggle" comes from the continuous-delivery literature; "feature flag" is the more common term today. You’ll also see "feature switch" and "feature management".',
  },
  {
    question: 'Do I need a database to run GO Feature Flag?',
    answer:
      'No. Flags live in a configuration file (YAML, JSON, or TOML) that the relay proxy loads into memory and polls for changes. There’s nothing to provision, migrate, or back up beyond the file itself, which you keep in Git.',
  },
  {
    question: 'Is GO Feature Flag free?',
    answer: (
      <>
        Yes. The project is open source under the MIT license and free to run,
        forever. Paid <Link to="/pricing">support contracts</Link> add an SLA
        and prioritized security fixes on top — the software itself stays free.
      </>
    ),
  },
  {
    question: 'Which languages does it support?',
    answer: (
      <>
        Evaluation goes through OpenFeature SDKs, so any OpenFeature-supported
        language works: Go, Java/Kotlin, JavaScript/TypeScript, Python, .NET,
        Ruby, Swift, PHP, and more. See the{' '}
        <Link to="/docs/sdk">full SDK list</Link>.
      </>
    ),
  },
  {
    question: 'How is it different from LaunchDarkly?',
    answer:
      'Same core idea, different model. LaunchDarkly defined the category and runs the backend for you; its pricing scales with seats and monthly active users — two numbers that only go up. GO Feature Flag is self-hosted and open source, so the bill doesn’t move with your traffic. The tradeoff is that you operate it yourself.',
  },
  {
    question: 'Can I migrate away later?',
    answer:
      'Yes. Because your apps talk to OpenFeature, not a proprietary SDK, you can change the backend without touching evaluation code. That’s the point of building on a standard.',
  },
];

export default function FeatureFlagPage() {
  const {siteConfig} = useDocusaurusContext();
  const githubUrl =
    siteConfig.customFields?.github ??
    'https://github.com/thomaspoignant/go-feature-flag';

  return (
    <Layout
      title={PAGE_TITLE}
      description={PAGE_DESCRIPTION}
      keywords={[
        'feature flag',
        'feature flags',
        'feature toggle',
        'feature toggles',
        'feature management',
        'open source feature flags',
        'self-hosted feature flags',
      ]}>
      <SeoHead
        title={PAGE_TITLE}
        description={PAGE_DESCRIPTION}
        path="/product/what-are-feature-flags"
        image="/img/landing/feature-flag/multi-ff.png"
        imageMeta
        imageWidth={1200}
        imageHeight={896}
      />
      <Title
        title={PAGE_TITLE}
        description="Deploy your code one day and release the feature another — to the users you choose, with an instant off switch."
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
        title="What is a feature flag?"
        imageSrc={deployVsRelease}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="Deploying code and releasing a feature are separate steps: shipped code passes through a feature-flag toggle that rolls out to a growing share of users.">
        <p className={prose}>
          A switch in your code that turns a feature on or off at runtime —{' '}
          <strong>no redeploy required</strong>.
        </p>
        <p className={prose}>
          You wrap a piece of functionality in a named conditional and decide,
          live, whether it runs. That one move decouples deploying code from
          releasing a feature. <br />
          New code can ship to production switched off, sitting there dark, and
          you turn it on when you&rsquo;re ready — not whenever the deploy
          happened to land.
        </p>
        {/* <div className="mx-auto w-full max-w-3xl">
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={SIMPLE_FLAG_YAML}
            callout="One flag, two variations, a default. Change the file and every app reading it picks up the new value — no redeploy."
          />
        </div> */}
        <p className={clsx(prose, 'mt-4')}>
          Once that switch exists, it earns its keep:
        </p>
        <ul className="mb-4 list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Progressive rollout</strong> — release to 1% of users, watch
            the dashboards, widen the cohort only once you trust it.
          </li>
          <li>
            <strong>Kill switch</strong> — flip an expensive or risky feature
            off in seconds when traffic spikes. No redeploy, no incident bridge.
          </li>
          <li>
            <strong>Targeted release</strong> — gate features by plan, region,
            or user, so the right people get the right experience.
          </li>
          <li>
            <strong>Experiments</strong> — send cohorts down different paths and
            settle the debate with data, not opinions.
          </li>
        </ul>
        <p className={clsx(prose, 'mt-4')}>
          Feature flags are the building block of the wider practice of{' '}
          <Link to="/product/what_is_feature_management">
            feature management
          </Link>
          .
        </p>
      </FeatureRow>

      <Cards
        title="Feature flags support teams looking to:"
        cards={CAPABILITY_CARDS}
      />

      {/* How it works + tabbed code */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          How feature flags work
        </h2>
        <p className={prose}>Three pieces do the work:</p>
        <ul className="mb-6 list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Variations</strong> — the values a flag can return. Often{' '}
            <code>on</code> / <code>off</code>, but a variation can also be a
            string, a number, or a JSON object.
          </li>
          <li>
            <strong>Evaluation context</strong> — who you&rsquo;re asking about.
            A <strong>targeting key</strong> (the unique identifier we bucket
            on) plus any attributes you have: plan, country, beta-tester, and so
            on.
          </li>
          <li>
            <strong>Targeting rules</strong> — the logic that maps a context to
            a variation.{' '}
            <em>
              Admins get the new UI. Everyone in the EU gets the old checkout.
              The rest get the default.
            </em>
          </li>
        </ul>
        <p className={prose}>
          Evaluation happens at runtime. Your application asks
          &ldquo;what&rsquo;s the value of this flag for this user?&rdquo; and
          always gets an answer back — including a safe default if anything goes
          wrong.
        </p>
        <div className="mx-auto w-full max-w-3xl">
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            callout="The flag returns enabled for user 1, disabled for everyone else. Same config, evaluated per user, on every request."
            tabs={[
              {
                value: 'yaml',
                label: 'Flag',
                language: 'yaml',
                displayName: 'flags.goff.yaml',
                code: TARGETING_YAML,
              },
              {
                value: 'go',
                label: 'Go',
                language: 'go',
                displayName: 'Go (server)',
                code: GO_SNIPPET,
              },
              {
                value: 'js',
                label: 'JavaScript',
                language: 'javascript',
                displayName: 'JavaScript (server)',
                code: JS_SNIPPET,
              },
            ]}
            moreLink={{to: '/docs/sdk', label: 'All SDKs'}}
          />
        </div>
        <p className="mt-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          Want the details? Read about{' '}
          <Link to="/docs/concepts/flag-evaluation">flag evaluation</Link> and
          the{' '}
          <Link to="/docs/concepts/evaluation-context">evaluation context</Link>
          .
        </p>
      </section>

      <FeatureRow
        eyebrow="The key element"
        title="Decouple deploy from release"
        imageSrc={killswitch}
        imageWidth={1200}
        imageHeight={747}
        reverse
        imageAlt="A kill switch turning a live feature off instantly, separating the moment code is deployed from the moment a feature is released.">
        <p className="mb-4">
          Deploying code and releasing a feature used to be the same event.
          Feature flags split them apart. Your code ships to production turned
          off, and you decide — separately, later — who sees it and when.
        </p>
        <p className="mb-0">
          Start with your own team, expand to 1% of users, watch your
          dashboards, then roll out to everyone. If something looks wrong, flip
          the flag back. No redeploy, no revert, no waiting on a pipeline.
        </p>
      </FeatureRow>

      <FeatureRow
        eyebrow="In practice"
        title="What you can do with feature flags"
        imageSrc={evaluationFlow}
        imageWidth={1200}
        imageHeight={896}
        imageAlt="A honeycomb of feature flag use cases: a toggle switch, A/B testing, a progressive rollout rocket, a kill switch, a safe-rollback shield, user targeting, and branching.">
        <ul className="m-0 list-disc space-y-2 pl-6">
          <li>
            <strong>
              <Link to="/product/rollouts">Progressive rollouts</Link>
            </strong>{' '}
            — ship to 1%, then 10%, then everyone, watching your metrics as you
            go.
          </li>
          <li>
            <strong>A/B testing and experimentation</strong> — serve different
            variations and measure which one wins.
          </li>
          <li>
            <strong>Kill switches</strong> — turn a misbehaving feature off in
            seconds, without a deploy.
          </li>
          <li>
            <strong>Safe rollbacks</strong> — back out a release by flipping a
            flag instead of reverting a commit.
          </li>
          <li>
            <strong>Targeted access</strong> — beta features for opted-in users,
            internal tools for staff, regional features by country.
          </li>
          <li>
            <strong>Trunk-based development</strong> — merge unfinished work
            behind an off flag instead of keeping long-lived branches.
          </li>
          <li>
            <strong>
              <Link to="/product/ai">Feature flags for AI</Link>
            </strong>{' '}
            — ship AI-generated code dark and roll out models and prompts
            safely, with a kill switch one flag away.
          </li>
        </ul>
      </FeatureRow>

      <Cards
        title="Types of feature flags"
        columns={4}
        cards={FLAG_TYPE_CARDS}
      />

      {/* Worked example */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          A feature flag, end to end
        </h2>
        <p className={prose}>
          Say you&rsquo;re rebuilding checkout. Here is the whole life of one
          flag, start to finish:
        </p>
        <ol className="mb-6 list-decimal space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Define it, off by default.</strong> Merge and deploy — the
            old checkout still runs for everyone, so the deploy is a non-event.
          </li>
          <li>
            <strong>Turn it on for your team.</strong> Add a targeting rule for
            internal users and click through the new flow in production.
          </li>
          <li>
            <strong>Open the gate gradually.</strong> Serve it to 5% of traffic,
            watch error rates and conversion, then widen to 25% and 100% as the
            numbers hold.
          </li>
          <li>
            <strong>Clean up.</strong> Once it&rsquo;s live everywhere and
            stable, delete the flag and the old code path.
          </li>
        </ol>
        <p className={prose}>
          A percentage rollout is only a few lines — bump the numbers and the
          relay proxy picks them up on its next poll:
        </p>
        <div className="mx-auto w-full max-w-3xl">
          <CodeCard
            filename="flags.goff.yaml"
            language="yaml"
            code={PERCENTAGE_YAML}
            callout="Change the split in the file; no redeploy, no restart. The next poll rolls it out."
          />
        </div>
      </section>

      {/* GOFF positioning */}
      <FeatureRow
        eyebrow="Self-hosted"
        title="Feature flags with GO Feature Flag"
        imageSrc={gofflogo}
        imageWidth={400}
        imageHeight={378}
        imageAlt="GO Feature Flag logo"
        imageClassName="mx-auto h-auto w-auto max-w-[500px]"
        actions={[
          {label: 'Why GO Feature Flag', href: '/product/why_go_feature_flag'},
        ]}>
        <p className="mb-4">
          You don&rsquo;t need a SaaS account for any of this.{' '}
          <strong>GO Feature Flag</strong> is an open-source, self-hosted
          feature flag system. What that means in practice:
        </p>
        <ul className="m-0 list-disc space-y-2 pl-6">
          <li>
            <strong>No database.</strong> Your flags live in a file you already
            version in Git. The relay proxy reads it into memory and polls for
            changes.
          </li>
          <li>
            <strong>No per-seat bill.</strong> It runs on infrastructure you
            already own; the cost doesn&rsquo;t climb with your team or traffic.
          </li>
          <li>
            <strong>No lock-in.</strong> Apps read flags through standard
            OpenFeature SDKs, so you can swap the backend out anytime.
          </li>
          <li>
            <strong>MIT-licensed.</strong> The whole project is free, with no
            paywalled feature tier.
          </li>
        </ul>
        <p className="mb-0 mt-4">
          The honest tradeoff: you self-host it, so you run it. For teams that
          want to own their data and keep their bill flat, that&rsquo;s usually
          the reason they choose it.
        </p>
      </FeatureRow>

      <section className="px-6 pb-12 max-w-[1400px] mx-auto">
        <p className={`${prose} text-center`}>
          Point the relay proxy at your config file and you have a working flag
          backend:
        </p>
        <div className="mx-auto w-full max-w-3xl">
          <CodeCard
            filename="terminal"
            language="shell"
            code={DOCKER_SNIPPET}
            callout="One container, one YAML file, two lines of SDK code. That's the whole setup."
          />
        </div>
      </section>

      <FeatureRow
        eyebrow="Open standard"
        title="Built on OpenFeature"
        reverse
        imageSrc={useBaseUrl('/img/features/openfeature.svg')}
        imageAlt="OpenFeature ecosystem"
        actions={[
          {
            label: 'Our OpenFeature support',
            href: '/product/open-feature',
          },
        ]}>
        <p className="mb-4">
          GO Feature Flag speaks <strong>OpenFeature</strong>, the
          vendor-neutral standard for feature flagging. Your apps evaluate flags
          through the standard SDKs, not a proprietary client.
        </p>
        <p className="mb-0">
          That&rsquo;s what keeps you free to leave: outgrow us, change your
          mind, or mix tools later, and your evaluation code stays the same.
        </p>
      </FeatureRow>

      {/* Lifecycle / flag debt */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Lifecycle and flag debt
        </h2>
        <p className={prose}>
          Not every flag is temporary, and treating them all the same is how
          teams get into trouble.
        </p>
        <ul className="mb-4 list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Release flags are short-lived.</strong> They exist to ship
            one feature safely, then they should be removed. Left behind, they
            become <em>flag debt</em> — dead conditionals that make code harder
            to read, test, and reason about.
          </li>
          <li>
            <strong>
              Operational and permission flags can live for years.
            </strong>{' '}
            Kill switches and entitlement checks are part of how the system
            runs, not temporary scaffolding.
          </li>
        </ul>
        <p className="m-0 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          The discipline that keeps flags from rotting is simple: give every
          flag an owner and an expected lifespan, review the list on a schedule,
          and delete release flags the moment a feature is fully live. A flag
          you&rsquo;re afraid to remove is a flag you no longer understand.
        </p>
      </section>

      {/* Best practices */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto">
        <h2 className={`${sectionHeading} text-center`}>
          Feature flag best practices
        </h2>
        <p className={prose}>
          A few habits keep feature flags from turning into technical debt:
        </p>
        <ul className="mb-4 list-disc space-y-2 pl-6 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          <li>
            <strong>Keep release flags short-lived.</strong> Delete them once
            the feature is fully rolled out.
          </li>
          <li>
            <strong>Name them so anyone can read them</strong> —{' '}
            <code>release-new-checkout</code> tells you what it does and when it
            can go.
          </li>
          <li>
            <strong>Default to off</strong>, and pick a safe fallback value for
            every evaluation.
          </li>
          <li>
            <strong>Test both states</strong> — the on path <em>and</em> the off
            path.
          </li>
          <li>
            <strong>Know who owns each flag</strong>, and review the list
            regularly.
          </li>
        </ul>
        <p className="m-0 text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          We go deeper in the{' '}
          <Link to="/blog/feature-flag-best-practice">
            feature flag best practices guide
          </Link>
          .
        </p>
      </section>

      <CtaBand
        title="Ship your first flag in minutes"
        description="One Docker container, a YAML file, and two lines of SDK code. Self-hosted, OpenFeature-native, MIT-licensed."
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
