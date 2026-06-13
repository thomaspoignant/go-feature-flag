import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import {Whatis} from '../components/home/whatis';
import {QuickStart} from '../components/home/HomepageQuickStart';
import {HomeHeader} from '../components/home/HomeHeader';
import {Benefit} from '../components/home/benefit';
import {WhyGoff} from '../components/home/why-goff';
import {
  Integration,
  OpenFeatureEcosystem,
  Rollout,
  Sdk,
} from '../components/home/features';
import {UsingIt} from '../components/home/using-it';
import {Headline} from '../components/home/headline';
import {HowItWorks} from '../components/home/how-it-works';

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.tagline}`}
      description={`${siteConfig.customFields.description}`}>
      <HomeHeader />
      <UsingIt />
      <Whatis />

      <QuickStart />
      <div className={'my-10'}></div>
      <OpenFeatureEcosystem />
      <Sdk />
      <Integration />
      <WhyGoff />
      <Headline />
      <Rollout />
      <HowItWorks />
      <Benefit />
    </Layout>
  );
}
