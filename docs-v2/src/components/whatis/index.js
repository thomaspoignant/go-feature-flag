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
            <p>GO Feature Flag is a lightweight feature flag solution build in GO.</p>
            <p>You don't need a complex infrastructure to install, you just have a single configuration file you can host
              anywhere (<code>HTTP</code>, <code>S3</code>, <code>Kubernetes</code>, <code>file</code>, ...).
              GO Feature Flag can be used directly and without any server in your GO application <i>(using the library)</i>, but
              can also be used with different languages (<code>JAVA</code>, <code>TypeScript</code>, <code>JavaScript</code>, ...)
              with the usage of the relay proxy and the <Link to={siteConfig.customFields.openfeature}>Openfeature</Link> SDKs.
            </p>
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
