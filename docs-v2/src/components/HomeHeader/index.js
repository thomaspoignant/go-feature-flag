import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './styles.module.css';
import  Link from '@docusaurus/Link';
import clsx from "clsx";

export function HomeHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <img src={siteConfig.customFields.logo} alt="site logo"/>
        <h1 className={styles.title}>{siteConfig.title}</h1>
        <p className={styles.subtitle}>{siteConfig.tagline}</p>
        <div className={styles.stars}>
          <a href={siteConfig.customFields.github} target="_blank">
            <img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/thomaspoignant/go-feature-flag?style=social"/>
          </a>
        </div>
        <div>
          <Link to={siteConfig.customFields.github} className={clsx(styles.github, styles.button, 'button button--lg')}>
            GITHUB
          </Link>
          <Link to="/docs/intro" className={clsx(styles.getstarted, styles.button, 'button button--lg')}>
            GET STARTED 🚀
          </Link>
          <Link to={siteConfig.customFields.sponsor} className={clsx(styles.sponsor, styles.button, 'button button--lg')}>
            ❤️ SPONSOR
          </Link>
        </div>
      </div>
    </header>
  );
}
