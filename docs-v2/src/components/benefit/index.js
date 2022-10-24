import React from 'react';
import styles from './styles.module.css';
import clsx from "clsx";

export function Benefit() {
  return (
    <section className={styles.benefits}>
      <div className="container">
        <div className="row">
          <div className={clsx("col-lg-8 mx-auto text-center", styles.title)}>
            <div>
              <span className="goffMainTitle">Why use feature flags?</span><br/>
              <p>
                Feature flags is a modern software engineering technique that configure select functionality during runtime, without deploying new code.
              </p>
            </div>
          </div>
        </div>
      </div>
        <div className="mx-auto row align-self-center">
          <div className="offset-xl-1 col-xl-2 offset-lg-0 col-lg-4 col-md-6 text-center">
            <article className={styles.benefitsPostItem}>
              <div className="row">
                <img src="img/benefits/rocket.jpg" alt="post-thumb"/>
              </div>
              <div className="row">
                <h2>Test in production</h2>
                <p>Test directly in production with your real data by enabling the features to your QA. Decrease incident by disabling the feature as soon as a bug arise.</p>
              </div>
            </article>
          </div>
          <div className="col-xl-2 col-lg-4 col-md-6 text-center">
            <article className={styles.benefitsPostItem}>
              <div className="row">
                <img src="img/benefits/pm.jpg" alt="post-thumb"/>
              </div>
              <div className="row">
                <h2>Give autonomy to stakeholders</h2>
                <p>You don't need a software engineer to release a new feature, empower business stakeholders, no development skills are needed.</p>
              </div>
            </article>
          </div>
          <div className="col-xl-2 col-lg-4 col-md-6 text-center">
            <article className={styles.benefitsPostItem}>
              <div className="row">
                <img src="img/benefits/inovate.jpg" alt="post-thumb"/>
              </div>
              <div className="row">
                <h2>Innovate faster</h2>
                <p>Deploy code when it is convenient (several times a day). Release when it is ready and it brings value. Deliver software to target audiences progressively.</p>
              </div>
            </article>
          </div>
          <div className="col-xl-2 col-lg-4 col-md-6 text-center">
            <article className={styles.benefitsPostItem}>
              <div className="row">
                <img src="img/benefits/data.jpg" alt="post-thumb"/>
              </div>
              <div className="row">
                <h2>Experiment and learn</h2>
                <p>Try new features and measure their success while running A/B test. Export who was using which variation and learn what is successful for your business.</p>
              </div>
            </article>
          </div>
          <div className="col-xl-2 col-lg-4 col-md-6 text-center">
            <article className={styles.benefitsPostItem}>
              <div className="row">
                <img src="img/benefits/devteam.jpg" alt="post-thumb"/>
              </div>
              <div className="row">
                <h2>Make engineers happy and productive</h2>
                <p>Have a better developer experience with simplifying how to release, test and deploy your software.</p>
              </div>
            </article>
          </div>
        </div>
    </section>
  );
}
