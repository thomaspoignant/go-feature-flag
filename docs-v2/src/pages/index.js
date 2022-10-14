import React from 'react';
import clsx from 'clsx';
import  Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import {HomeHeader} from "@site/src/components/homeHeader";

import styles from './index.module.css';

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Description will go into a meta tag in <head />">
      <HomeHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
