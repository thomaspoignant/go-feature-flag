import React from 'react';
import clsx from 'clsx';
import  Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import {Whatis} from "@site/src//components/whatis";
import HomepageFeatures from "../components/HomepageFeatures";
import {HomeHeader} from "../components/HomeHeader"
import {Benefit} from "../components/benefit";


export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Description will go into a meta tag in <head />">
      <HomeHeader />
      <main>
        <Whatis />
        <Benefit />
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
