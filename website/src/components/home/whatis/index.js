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
      <div className="relative inline-flex group ml-5">
        <div className="border-gray-700 border-4 absolute transitiona-all duration-1000 opacity-70 -inset-px bg-gradient-to-r from-[#44BCFF] via-[#FF44EC] to-[#FF675E] rounded-xl blur-lg group-hover:opacity-100 group-hover:-inset-1 group-hover:duration-200 animate-tilt hover:no-underline"></div>
        <Link
          to={'/docs/'}
          title="Dive into GO Feature Flag"
          className="hover:no-underline hover:text-gray-700 relative inline-flex items-center justify-center px-6 py-3 text-white transition-all duration-200 bg-[#9fbeb3] font-pj rounded-xl focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900">
          <i className="fa-solid fa-screwdriver-wrench mr-1"></i>
          Dive into GO Feature Flag
        </Link>
      </div>
    </section>
  );
}
