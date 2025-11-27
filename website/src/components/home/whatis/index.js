import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Link from '@docusaurus/Link';
import clsx from 'clsx';

export function Whatis() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <section className="py-[2rem] text-center bg-gray-100 dark:bg-[#363636]">
      <span className="text-gray-800 dark:text-gray-50 text-5xl font-poppins font-bold tracking-[-0.18rem]">
        What is GO Feature Flag?
      </span>
      <div
        className={clsx(
          'flex items-center justify-center m-auto py-[2rem] font-poppins text-left text-md max-w-6xl 2xl:max-w-full'
        )}>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-x-5 max-w-full lg:gap-x-10 md:max-w-7xl">
          <div>
            <h2>
              <i className="fa-solid fa-laptop-code text-titles-500"></i> Quick
              and Easy Setup
            </h2>
            <p className="mb-2">
              GO Feature Flag believes in simplicity and offers a simple and
              lightweight solution to use feature flags.
            </p>
            <p className="mb-2">
              Our focus is to avoid any complex infrastructure work to use the
              solution.
            </p>
          </div>
          <div>
            <h2>
              <i className="fa-solid fa-rectangle-list text-titles-500"></i>{' '}
              Complete Feature Flag Solution
            </h2>
            <p className="mb-2">
              Target individual segments, users, and development environments,
              use advanced rollout functionality.
            </p>
            <p className="mb-2">
              You can also collect usage data of your flags and be notified of
              configuration changes.
            </p>
          </div>
          <div>
            <h2>
              <i className="fa-solid fa-terminal text-titles-500"></i> Developer
              Optimized
            </h2>
            <p className="mb-2">
              100% Opensource, no vendor locking, supports your favorite
              languages and is pushing for standardisation with the support of{' '}
              <Link to={siteConfig.customFields.openfeature}>OpenFeature</Link>.
            </p>
            <p className="mb-2">
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
