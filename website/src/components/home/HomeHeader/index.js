import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './styles.module.css';
import Link from '@docusaurus/Link';
import clsx from 'clsx';

export function HomeHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <section className={styles.hero}>
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
      <div className={clsx('container', styles.container)}>
        <div className="row">
          <div className="col col--6">
            <div className={styles.heroContent}>
              <span className="text-gray-800 dark:text-gray-50 text-5xl font-poppins font-bold tracking-[-0.18rem]">
                GO Feature Flag
              </span>
              <br />
              <span className="text-titles-500 text-[2.8rem] font-poppins font-bold tracking-[-0.18rem] leading-10">
                {`${siteConfig.tagline}`}
              </span>
              <p>
                <span className={styles.descriptionFirstLine}>
                  Ship Faster, Reduce Risk, and Build Scale
                </span>
                <br />
                Feature flags lets you modify system behavior without changing
                code. Deploy every day, release when you are ready. Reduce risk
                by releasing your features progressively.
              </p>
            </div>
            <div className={styles.ghStars}>
              <Link to={siteConfig.customFields.github}>
                <img
                  alt="GitHub Repo stars"
                  src="https://img.shields.io/github/stars/thomaspoignant/go-feature-flag?style=social"
                />
              </Link>
            </div>
            <div className={'text-center items-center gap-30'}>
              <div className="relative inline-flex group">
                <div className="absolute transitiona-all duration-1000 opacity-70 -inset-px bg-gradient-to-r from-[#44BCFF] via-[#FF44EC] to-[#FF675E] rounded-xl blur-lg group-hover:opacity-100 group-hover:-inset-1 group-hover:duration-200 animate-tilt"></div>
                <Link
                  to={siteConfig.customFields.github}
                  title="Available on GitHub"
                  className="hover:no-underline hover:text-gray-500 relative inline-flex items-center justify-center px-8 py-4 text-lg font-bold text-white transition-all duration-200 bg-gray-900 font-pj rounded-xl focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900">
                  <i className="fa-brands fa-github mr-4"></i>
                  Available on GitHub
                </Link>
              </div>
              <div className="relative inline-flex group ml-5 mt-4">
                <div className="border-gray-700 border-4 absolute transitiona-all duration-1000 opacity-70 -inset-px bg-gradient-to-r from-[#44BCFF] via-[#FF44EC] to-[#FF675E] rounded-xl blur-lg group-hover:opacity-100 group-hover:-inset-1 group-hover:duration-200 animate-tilt hover:no-underline"></div>
                <Link
                  to={'/docs/getting-started'}
                  title="Get Started"
                  className="hover:no-underline hover:text-gray-700 relative inline-flex items-center justify-center px-8 py-4 text-lg font-bold text-white transition-all duration-200 bg-[#9fbeb3] font-pj rounded-xl focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900">
                  <i className="fa-solid fa-circle-right mr-4"></i>
                  Get Started
                </Link>
              </div>
            </div>
          </div>
          <div className="max-md:hidden">
            <div className="hero-image">
              <img
                src={siteConfig.customFields.logo}
                alt="go-feature-flag-logo"
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
