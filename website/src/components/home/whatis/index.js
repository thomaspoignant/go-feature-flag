import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './styles.module.css';
import Link from '@docusaurus/Link';
import clsx from 'clsx';

export function Whatis() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <section className={clsx(styles.whatis)}>
      <span className="goffMainTitle">What is GO Feature Flag?</span>
      <div
        className={clsx(
          styles.description,
          'flex items-center justify-center'
        )}>
        <div
          className={clsx(
            'grid grid-cols-1 md:grid-cols-3 gap-x-5 max-w-full lg:gap-x-10 md:max-w-7xl'
          )}>
          <div className={''}>
            <h2>
              <i className="fa-solid fa-laptop-code"></i> Quick and Easy Setup
            </h2>
            <p>
              GO Feature Flag believes in simplicity and offers a simple and
              lightweight solution to use feature flags.
            </p>
            <p>
              Our focus is to avoid any complex infrastructure work to use the
              solution.
            </p>
          </div>
          <div className={''}>
            <h2>
              <i className="fa-solid fa-rectangle-list"></i> Complete Feature
              Flag Solution
            </h2>
            <p>
              Target individual segments, users, and development environments,
              use advanced rollout functionality.
            </p>
            <p>
              You can also collect usage data of your flags and be notified of
              configuration changes.
            </p>
          </div>
          <div className={''}>
            <h2>
              <i className="fa-solid fa-terminal"></i> Developer Optimized
            </h2>
            <p>
              100% Opensource, no vendor locking, supports your favorite
              languages and is pushing for standardisation with the support of{' '}
              <Link to={siteConfig.customFields.openfeature}>OpenFeature</Link>.
            </p>
            <p>
              File based configuration, integrated with the tools that you
              already use.
            </p>
          </div>
        </div>
      </div>
      <Link to="/docs/">
        <button
          type="button"
          className="cursor-pointer text-white bg-gradient-to-br from-purple-600 to-blue-500 hover:bg-gradient-to-bl focus:ring-4 focus:outline-none focus:ring-blue-300 dark:focus:ring-blue-800 font-medium rounded-lg text-sm px-5 py-2.5 text-center me-2 mb-2">
          <i className="fa-solid fa-screwdriver-wrench"></i> Dive into GO
          Feature Flag
        </button>
      </Link>
    </section>
  );
}
