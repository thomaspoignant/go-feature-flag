import React, {useEffect, useState} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './styles.module.css';
import Link from '@docusaurus/Link';
import clsx from 'clsx';
import {FaStar, FaArrowAltCircleRight} from 'react-icons/fa';
import {sdk} from '@site/data/sdk';

function GetStartedButton() {
  return (
    <div className="relative inline-flex group">
      <div className="absolute transitiona-all duration-1000 opacity-70 -inset-px bg-gradient-to-r from-[#44BCFF] via-[#FF44EC] to-[#FF675E] rounded-xl blur-lg group-hover:opacity-100 group-hover:-inset-1 group-hover:duration-200 animate-tilt"></div>
      <Link
        to={'/docs/'}
        title="Get Started with GO Feature Flag"
        className="hover:no-underline hover:text-white relative inline-flex items-center justify-center px-8 py-4 text-lg font-bold text-white transition-all duration-200 bg-gray-900 font-pj rounded-xl overflow-hidden border-2 border-solid border-transparent focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900">
        <FaArrowAltCircleRight className="mr-2" />
        Get Started in 60 seconds
      </Link>
    </div>
  );
}

function ViewOnGitHubButton() {
  const {siteConfig} = useDocusaurusContext();
  const [githubStars, setGithubStars] = useState(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    const shieldsUrl = `https://img.shields.io/github/stars/${siteConfig.organizationName}/${siteConfig.projectName}.json`;

    fetch(shieldsUrl, {signal: controller.signal})
      .then(response => {
        if (!response.ok) {
          throw new Error(`shields.io responded with ${response.status}`);
        }
        return response.json();
      })
      .then(data => {
        const isStarCount =
          typeof data?.message === 'string' &&
          /^[\d.,]+[kmb]?$/i.test(data.message.trim());
        if (isStarCount) {
          setGithubStars(data.message);
        } else {
          setFailed(true);
        }
      })
      .catch(error => {
        if (error.name !== 'AbortError') {
          console.error('Failed to fetch GitHub star count:', error);
          setFailed(true);
        }
      });

    return () => controller.abort();
  }, [siteConfig.organizationName, siteConfig.projectName]);

  const githubLinkTitle = githubStars
    ? `View on GitHub (${githubStars} stars)`
    : 'View on GitHub';

  return (
    <div className="inline-flex">
      <Link
        to={siteConfig.customFields.github}
        title={githubLinkTitle}
        className="hover:no-underline inline-flex items-center justify-center px-8 py-3 text-lg font-bold text-gray-800 dark:text-gray-100 hover:text-white bg-transparent hover:bg-[#9fbeb3] border-2 border-solid border-[#9fbeb3] transition-all duration-200 font-pj rounded-xl focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900">
        <i className="fa-brands fa-github mr-4" aria-hidden="true"></i>
        View on GitHub
        {(githubStars !== null || !failed) && (
          <span className="ml-2 font-semibold">
            <div className="flex items-center tabular-nums">
              <FaStar className="w-6 h-6 mr-1" />{' '}
              <span className={clsx('min-w-[3ch]', !githubStars && 'opacity-0')}>
                {githubStars ?? '0.0k'}
              </span>
            </div>
          </span>
        )}
      </Link>
    </div>
  );
}

function HeroBackgroundShapes() {
  return (
    <div className={styles.heroShape}>
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 898.595 671.33">
        <g data-name="Group 1168">
          <path
            data-name="Path 1352"
            d="M77.225.84S65.754 56.25 72.99 108.615c6.519 47.174 14.071 83.313 45.359 132.19s67.663 74.2 113.344 90.467c10.087 2.544 22.468 4.651 35.375 10.446 17.912 8.956 39.851 18.784 63.185 64.959 29.724 58.823 31.289 129.222 102.94 193.364 30.523 27.324 85.56 48.346 152.718 60.609 90.568 16.538 203.528 16.044 311.709-17.053 0-46.584-.016-642.734-.016-642.734z"
            fill="#cdf7e7"
          />
          <path
            data-name="Path 1353"
            d="M4.946.863s-11.953 71.78 4.545 135.746 50.072 127.642 106.223 162.953c30.391 18.524 54.077 22.62 54.077 22.62s35.965 6.587 58.362 28.851 33.95 47.335 40.287 63.6 30.656 87.859 39.048 101.217 22.093 51.037 70.9 84.776 130.668 56.964 257.731 60.438 240.329-44.6 261.458-55.888"
            fill="none"
            stroke="#273437"
            strokeLinecap="round"
            strokeWidth="1.5"
          />
        </g>
      </svg>
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 266.025 234.723">
        <g data-name="Group 1169">
          <path
            data-name="Path 1351"
            d="M246.353.908s32.245 23.839 14.178 60.475-62.607 44.54-85.191 84.439-37.268 86.821-83.942 85.186c-36.9-1.289-90.335-44.54-90.335-44.54L1.188.657z"
            fill="#cdf7e7"
          />
          <path
            data-name="Path 1350"
            d="M178.777.908s4.869 19.639 32.623 49.99 37.492 50.964 33.6 76.932-38.147 51.938-67.684 70.765-58.263 37.33-101.763 35.22c-40.1-1.945-59.9-27.777-71.327-57.935-.862-2.273-1.961-6.034-3.171-6.986"
            fill="none"
            stroke="#273437"
            strokeLinecap="round"
            strokeWidth="1.5"
          />
        </g>
      </svg>
    </div>
  );
}

function HeroCopy() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <div className={styles.heroContent}>
      <span className="block text-base md:text-2xl font-poppins font-semibold tracking-tight text-titles-500 mb-2">
        {`${siteConfig.title}`}
      </span>
      <h1 className="text-[2.25rem] md:text-[3.5rem] leading-[1.1] md:leading-[1.05] font-poppins font-extrabold tracking-[-0.08rem] md:tracking-[-0.18rem] text-gray-800 dark:text-gray-50 m-0 mb-5">
        Open-source feature flags
        <span className="text-titles-500"> —&nbsp;built on OpenFeature.</span>
      </h1>
      <p className="text-xl md:text-lg leading-relaxed font-normal mb-10 text-[color:var(--goff-main-ff-description)]">
        The feature flag tool for the AI era: ship fast, roll out safely, roll
        back instantly. Set up in minutes with no infrastructure to manage — and
        it works with the stack you already use.
      </p>
    </div>
  );
}

function HeroCTAs() {
  return (
    <div className="flex flex-wrap items-center justify-center gap-4">
      <GetStartedButton />
      <ViewOnGitHubButton />
    </div>
  );
}

function HeroProofLine() {
  return (
    <p className="mt-6 text-center text-sm md:text-base text-gray-500 dark:text-gray-400">
      100% OpenSource · MIT Licensed · Support{' '}
      <span className="font-bold text-gray-900 dark:text-gray-100">
        {sdk.length} languages & frameworks
      </span>
    </p>
  );
}

function HeroLogo() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <div className="max-md:hidden">
      <div className="hero-image">
        <img src={siteConfig.customFields.logo} alt="go-feature-flag-logo" />
      </div>
    </div>
  );
}

export function HomeHeader() {
  return (
    <section className={styles.hero}>
      <HeroBackgroundShapes />
      <div className={clsx('container', styles.container)}>
        <div className="row">
          <div className="col col--6">
            <HeroCopy />
            <HeroCTAs />
            <HeroProofLine />
          </div>
          <HeroLogo />
        </div>
      </div>
    </section>
  );
}
