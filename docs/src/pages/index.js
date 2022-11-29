import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import {Whatis} from "../components/whatis";
import {HomeHeader} from "../components/HomeHeader"
import {Benefit} from "../components/benefit";
import {Features} from "../components/features";

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Description will go into a meta tag in <head />">
      <HomeHeader />
      <Whatis />
      <Features />
      <Benefit />
    </Layout>
  );
}
