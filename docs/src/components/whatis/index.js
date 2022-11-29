import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './styles.module.css';
import  Link from '@docusaurus/Link';
import clsx from "clsx";

export function Whatis() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <section className={styles.whatis}>
      <div className="grid grid-pad">
        <div className={clsx("col-1-1")}>
          <span className="goffMainTitle">What is GO Feature Flag?</span>
          <div className={styles.description}>
            <p>GO Feature Flag believes in simplicity and offers a simple and lightweight solution to use feature flags. Our focus is to avoid any complex infrastructure work to use GO Feature Flag.</p>
            <p>This is a complete feature flagging solution with the possibility to target only a group of users, use any types of flags, store your configuration in various location and advanced rollout functionality. You can also collect usage data of your flags and be notified of configuration changes.</p>
            <p>The solution can be used in 2 different ways, directly using the GO module in your code (you will have no backend to install) or via the <Link to={siteConfig.customFields.openfeature}>OpenFeature</Link> standard which allows to use vendor agnostic SDKs.</p>
          </div>
          <Link to="/docs/">

            <button className="pushy__btn pushy__btn--md pushy__btn--red">
              <i className="fa-solid fa-screwdriver-wrench"></i> Dive into GO Feature Flag
            </button>
          </Link>
        </div>
      </div>
    </section>
  );
}
