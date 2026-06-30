import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  FaGithub,
  FaHeart,
  FaUnlock,
  FaToggleOn,
  FaUsers,
  FaBalanceScale,
  FaHandshake,
  FaShieldAlt,
  FaComments,
} from 'react-icons/fa';
import {RiOpenSourceFill} from 'react-icons/ri';
import SeoHead from '../../../components/landingPage/seo';
import Title from '../../../components/landingPage/title';
import Cards from '../../../components/landingPage/cards';
import FeatureRow from '../../../components/landingPage/featureRow';
import CtaBand from '../../../components/landingPage/ctaBand';
import Faq from '../../../components/landingPage/faq';

const PAGE_TITLE = 'Open source';
const PAGE_DESCRIPTION =
  'GO Feature Flag is 100% open source under the MIT license, self-hosted, and built on the OpenFeature standard - no vendor lock-in. We do not sell the project, so you can sponsor it or get paid support to help keep it going.';

const GITHUB_SPONSORS_URL = 'https://github.com/sponsors/thomaspoignant';
const ADOPTERS_URL =
  'https://github.com/thomaspoignant/go-feature-flag/blob/main/ADOPTERS.md';
const CONTRIBUTING_DOCS = '/docs/contributing';
const PRICING_URL = '/pricing';
const ENTERPRISE_MAILTO =
  'mailto:contact@gofeatureflag.org?subject=Enterprise support';
const BOOK_MEETING_URL = 'https://zcal.co/gofeatureflag/30min';
const SLACK_URL = '/slack';

const sectionHeading =
  'mb-4 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl';
const prose = 'mb-4 text-lg leading-relaxed text-gray-700 dark:text-gray-300';
const docLink =
  'mt-4 inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80';

const COMMITMENT_CARDS = [
  {
    icon: <RiOpenSourceFill />,
    title: 'MIT-licensed, 100% open source',
    description:
      'The whole project lives in the open under the permissive MIT license. Read it, fork it, run it, and build on it - no asterisks, no open-core bait and switch.',
  },
  {
    icon: <FaUnlock />,
    title: 'Self-hosted, no vendor lock-in',
    description:
      'GO Feature Flag runs on the infrastructure you already have. No database to operate, no per-seat bill, and no vendor that can pull the rug out from under you.',
  },
  {
    icon: <FaToggleOn />,
    title: 'Built on an open standard',
    description:
      'Your app evaluates flags through the vendor-neutral OpenFeature API, not a proprietary client - so you stay portable across the whole ecosystem.',
    link: '/product/open-feature',
    linkLabel: 'OpenFeature support',
  },
];

const SUPPORT_BULLETS = [
  {
    icon: <FaShieldAlt />,
    text: 'Premium support and an SLA on CVE fixes',
  },
  {
    icon: <FaComments />,
    text: 'A direct line to the maintainers - they can join your Slack or Teams',
  },
  {
    icon: <FaHandshake />,
    text: 'Help during your integration and a preview of the roadmap',
  },
];

const FAQ_ITEMS = [
  {
    question: 'Is GO Feature Flag really free?',
    answer:
      'Yes. GO Feature Flag is 100% open source and free forever. You get access to every feature, run it in your own infrastructure, and create unlimited feature flags - at no cost. The only thing the free tier does not include is dedicated, paid support.',
  },
  {
    question: 'What license is GO Feature Flag under?',
    answer:
      'The MIT license - one of the most permissive open-source licenses. You are free to use, modify, and distribute it, including in commercial products, with no obligation to open source your own code.',
  },
  {
    question: 'How is the project funded?',
    answer:
      'GO Feature Flag does not make money directly - there is no paid SaaS, no proprietary edition, and no per-seat pricing. It is sustained by the community: people who sponsor the project on GitHub and organizations that pay for enterprise support to keep the maintainers working on it.',
  },
  {
    question: 'How can I support GO Feature Flag?',
    answer: (
      <>
        Two easy ways, both valued. You can{' '}
        <Link to={GITHUB_SPONSORS_URL}>sponsor the project on GitHub</Link> to
        directly fund the work, and you can{' '}
        <Link to={ADOPTERS_URL}>add your company to the adopters list</Link> to
        show the project is trusted in production - a free way to give back.
        Starring the repo and contributing code or docs help too.
      </>
    ),
  },
  {
    question: 'Do you offer paid or enterprise support?',
    answer: (
      <>
        Yes. If you need more than community support, we offer an enterprise
        plan with premium support, an SLA on CVE fixes, a direct line to the
        maintainers, help during integration, and a roadmap preview. See the{' '}
        <Link to={PRICING_URL}>pricing page</Link> for the full breakdown or{' '}
        <Link to={ENTERPRISE_MAILTO}>contact us</Link>.
      </>
    ),
  },
  {
    question: 'How do I contribute?',
    answer: (
      <>
        Contributions are very welcome - code, docs, bug reports, or ideas. Read
        the <Link to={CONTRIBUTING_DOCS}>contributing guide</Link> to get
        started, and join the community on <Link to={SLACK_URL}>Slack</Link>.
      </>
    ),
  },
];

export default function OpenSourcePage() {
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
        'MIT license feature flags',
        'self-hosted feature flags',
        'no vendor lock-in',
        'sponsor GO Feature Flag',
        'GitHub sponsors',
        'open source support',
        'OpenFeature',
      ]}>
      <SeoHead
        title={PAGE_TITLE}
        description={PAGE_DESCRIPTION}
        path="/product/open_source"
        image="/img/logo/x-card.png"
      />

      <Title
        title="Open source, and committed to staying that way"
        description="GO Feature Flag is 100% open source under the MIT license - self-hosted, OpenFeature-native, and free forever. We do not sell it, so the project runs on sponsorships and the teams who choose paid support."
        actions={[
          {
            label: 'View on GitHub',
            href: githubUrl,
            icon: <FaGithub />,
          },
          {
            label: 'Sponsor us ❤️',
            href: GITHUB_SPONSORS_URL,
            icon: <FaHeart />,
            variant: 'secondary',
          },
        ]}
      />

      <Cards
        title="What our open-source commitment means for you"
        cards={COMMITMENT_CARDS}
      />

      {/* Our commitment to open source */}
      <FeatureRow
        eyebrow="Our commitment"
        title="Everything, in the open, under MIT"
        media={
          <div className="mx-auto flex w-full max-w-sm flex-col gap-4">
            <div className="flex items-center gap-4 rounded-2xl border border-solid border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-[#1f2024]">
              <FaBalanceScale className="h-9 w-9 shrink-0 text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]" />
              <div>
                <p className="m-0 text-lg font-bold text-gray-800 dark:text-gray-50">
                  MIT license
                </p>
                <p className="m-0 text-sm text-[color:var(--goff-main-ff-description)]">
                  Permissive. Use it anywhere, even commercially.
                </p>
              </div>
            </div>
            <div className="flex items-center gap-4 rounded-2xl border border-solid border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-[#1f2024]">
              <RiOpenSourceFill className="h-9 w-9 shrink-0 text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]" />
              <div>
                <p className="m-0 text-lg font-bold text-gray-800 dark:text-gray-50">
                  100% open source
                </p>
                <p className="m-0 text-sm text-[color:var(--goff-main-ff-description)]">
                  No open-core, no paywalled features.
                </p>
              </div>
            </div>
          </div>
        }>
        <p className={prose}>
          GO Feature Flag has been open source from day one, and it is not a
          teaser for a paid edition. There is no &ldquo;enterprise
          version&rdquo; hiding the features you actually need behind a license
          key - the whole project is right there on GitHub under the{' '}
          <strong>MIT license</strong>.
        </p>
        <p className="mb-0">
          Development happens in the open: issues, pull requests, and the
          roadmap are all public, and contributions are genuinely welcome -
          whether that is code, documentation, a bug report, or an idea.
        </p>
        <div className="mt-6 flex flex-wrap gap-4">
          <Link className={`${docLink} mt-0`} to={githubUrl}>
            Browse the source <span aria-hidden="true">→</span>
          </Link>
          <Link className={`${docLink} mt-0`} to={CONTRIBUTING_DOCS}>
            How to contribute <span aria-hidden="true">→</span>
          </Link>
        </div>
      </FeatureRow>

      {/* A real community */}
      <FeatureRow
        eyebrow="A real community"
        title="Built with a community, not behind closed doors"
        reverse
        media={
          <div className="mx-auto grid w-full max-w-sm grid-cols-1 gap-4">
            <div className="flex items-center gap-4 rounded-2xl border border-solid border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-[#1f2024]">
              <FaComments className="h-9 w-9 shrink-0 text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]" />
              <p className="m-0 font-semibold text-gray-800 dark:text-gray-50">
                An active Slack channel for questions and help
              </p>
            </div>
            <div className="flex items-center gap-4 rounded-2xl border border-solid border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-[#1f2024]">
              <FaUsers className="h-9 w-9 shrink-0 text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]" />
              <p className="m-0 font-semibold text-gray-800 dark:text-gray-50">
                A growing list of companies running it in production
              </p>
            </div>
          </div>
        }>
        <p className={prose}>
          The project is shaped by the people who use it. Anyone can ask
          questions on Slack, report a bug, or propose a feature - the direction
          of GO Feature Flag is a public conversation.
        </p>
        <p className="mb-0">
          Already using it? Adding your organization to the adopters list is a
          quick, free way to support the project and help others trust it.
        </p>
        <div className="mt-6 flex flex-wrap gap-4">
          <Link className={`${docLink} mt-0`} to={SLACK_URL}>
            Join us on Slack <span aria-hidden="true">→</span>
          </Link>
          <Link className={`${docLink} mt-0`} to={ADOPTERS_URL}>
            See the adopters <span aria-hidden="true">→</span>
          </Link>
        </div>
      </FeatureRow>

      {/* Sponsor the project */}
      <section className="px-6 py-12 max-w-[1400px] mx-auto text-center">
        <span className="mb-2 block text-sm font-semibold uppercase tracking-wide text-titles-500">
          Help keep it going
        </span>
        <h2 className={`${sectionHeading} mx-auto`}>
          We don&rsquo;t sell GO Feature Flag
        </h2>
        <p className="mx-auto max-w-3xl text-lg leading-relaxed text-gray-700 dark:text-gray-300">
          There is no paid SaaS and no proprietary edition - the project does
          not make money directly. It keeps moving thanks to people who choose
          to give back. If GO Feature Flag is useful to you or your team, here
          are two ways to help, both of them genuinely appreciated.
        </p>
        <div className="mx-auto mt-10 grid max-w-4xl grid-cols-1 gap-6 sm:grid-cols-2">
          <div className="flex h-full flex-col rounded-2xl border border-solid border-gray-200 bg-white p-8 text-left shadow-sm dark:border-gray-700 dark:bg-[#1f2024]">
            <FaHeart className="mb-4 h-9 w-9 text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]" />
            <h3 className="m-0 text-2xl font-bold text-gray-800 dark:text-gray-50">
              Sponsor on GitHub
            </h3>
            <p className="mb-6 mt-3 leading-relaxed text-[color:var(--goff-main-ff-description)]">
              Fund the work directly. Sponsorships pay for maintenance, new
              features, and keeping the project healthy and independent.
            </p>
            <Link
              to={GITHUB_SPONSORS_URL}
              className="mt-auto inline-flex items-center justify-center gap-2 rounded-lg bg-[color:var(--ifm-color-primary)] px-6 py-3 text-base font-semibold text-gray-600 no-underline hover:text-gray-800 hover:no-underline hover:brightness-110">
              <FaHeart /> Become a sponsor
            </Link>
          </div>
          <div className="flex h-full flex-col rounded-2xl border border-solid border-gray-200 bg-white p-8 text-left shadow-sm dark:border-gray-700 dark:bg-[#1f2024]">
            <FaUsers className="mb-4 h-9 w-9 text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]" />
            <h3 className="m-0 text-2xl font-bold text-gray-800 dark:text-gray-50">
              Add yourself to the adopters
            </h3>
            <p className="mb-6 mt-3 leading-relaxed text-[color:var(--goff-main-ff-description)]">
              No budget? No problem. Listing your company in the adopters file
              is free, takes a pull request, and helps the project grow.
            </p>
            <Link
              to={ADOPTERS_URL}
              className="mt-auto inline-flex items-center justify-center gap-2 rounded-lg border border-solid border-gray-300 bg-transparent px-6 py-3 text-base font-semibold text-gray-800 no-underline hover:border-gray-400 hover:no-underline dark:border-gray-600 dark:text-gray-100">
              <FaGithub /> Join the adopters
            </Link>
          </div>
        </div>
      </section>

      {/* Paid support */}
      <FeatureRow
        eyebrow="Need a hand?"
        title="Paid support, when community help isn't enough"
        media={
          <div className="mx-auto w-full max-w-sm rounded-2xl border border-solid border-gray-200 bg-white p-8 shadow-sm dark:border-gray-700 dark:bg-[#1f2024]">
            <p className="m-0 mb-5 text-sm font-semibold uppercase tracking-wide text-titles-500">
              Enterprise support includes
            </p>
            <ul className="m-0 list-none space-y-4 p-0">
              {SUPPORT_BULLETS.map(item => (
                <li key={item.text} className="flex items-start gap-3">
                  <span className="mt-1 shrink-0 text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]">
                    {item.icon}
                  </span>
                  <span className="text-gray-700 dark:text-gray-300">
                    {item.text}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        }>
        <p className={prose}>
          Some teams want more than community support - a guaranteed response, a
          maintainer on call, or a hand getting to production. For them we offer
          a paid <strong>enterprise support</strong> plan, and choosing it is
          one of the ways you keep the open-source project funded.
        </p>
        <p className="mb-0">
          You never have to pay to use GO Feature Flag. Support is there if and
          when you need it - everything else stays free and open.
        </p>
        <div className="mt-6 flex flex-wrap gap-4">
          <Link className={`${docLink} mt-0`} to={PRICING_URL}>
            See plans <span aria-hidden="true">→</span>
          </Link>
          <Link className={`${docLink} mt-0`} to={ENTERPRISE_MAILTO}>
            Contact us <span aria-hidden="true">→</span>
          </Link>
          <Link className={`${docLink} mt-0`} to={BOOK_MEETING_URL}>
            Book a meeting <span aria-hidden="true">→</span>
          </Link>
        </div>
      </FeatureRow>

      <CtaBand
        title="Open source today, open source tomorrow"
        description="MIT-licensed, self-hosted, and OpenFeature-native. Use it for free, sponsor it to keep it going, or reach out if your team needs paid support."
        actions={[
          {
            label: 'View on GitHub',
            href: githubUrl,
            icon: <FaGithub />,
          },
          {
            label: 'Sponsor us ❤️',
            href: GITHUB_SPONSORS_URL,
            icon: <FaHeart />,
            variant: 'secondary',
          },
        ]}
      />

      <Faq title="Frequently asked questions" items={FAQ_ITEMS} withJsonLd />
    </Layout>
  );
}
