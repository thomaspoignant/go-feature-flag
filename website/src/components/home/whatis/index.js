import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './styles.module.css';
import Link from '@docusaurus/Link';
import clsx from 'clsx';

export function Whatis() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <section className={styles.whatis}>
      <div className="grid grid-pad">
        <div className={clsx('col-1-1')}>
          <span className="goffMainTitle">What is GO Feature Flag?</span>
          <div className={clsx(styles.description, "grid grid-pad")}>
            <div className={"col-1-3 mobile-col-1-1"}>
              <h2><i className="fa-solid fa-laptop-code"></i> Quick and Easy Setup</h2>
              <p>GO Feature Flag believes in simplicity and offers a simple and lightweight solution to use feature flags.</p>
              <p>Our focus is to avoid any complex infrastructure work to use the solution.</p>
            </div>
            <div className={"col-1-3  mobile-col-1-1"}>
              <h2><i className="fa-solid fa-rectangle-list"></i> Complete Feature Flag Solution</h2>
              <p>Target individual segments, users, and development environments, use advanced rollout
              functionality.</p>
              <p>You can also collect usage data of your flags and be notified of configuration changes.</p>
            </div>
            <div className={"col-1-3  mobile-col-1-1"}>
              <h2><i className="fa-solid fa-terminal"></i> Developer Optimized</h2>
              <p>100% Opensource, no vendor locking, supports your favorite languages and is pushing for standardisation with the support of <Link to={siteConfig.customFields.openfeature}>OpenFeature</Link>.</p>
              <p>File based configuration, integrated with the tools that you already use.</p>
            </div>
          </div>
          <Link to="/docs/">
            <button className="pushy__btn pushy__btn--md pushy__btn--red">
              <i className="fa-solid fa-screwdriver-wrench"></i> Dive into GO
              Feature Flag
            </button>
          </Link>
        </div>
      </div>
    </section>
  );
}
