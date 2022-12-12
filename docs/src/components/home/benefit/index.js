import React from 'react';
import styles from './styles.module.css';
import clsx from "clsx";


function BenefitCard(props) {
  return (
    <div className="col-1-5 mobile-col-1-2">
      <article className={styles.benefitsPostItem}>
        <div className={styles.benefitsPostItemRow}>
          <img src={props.img} alt="post-thumb"/>
        </div>
        <div className={styles.benefitsPostItemRow}>
          <h2>{props.title}</h2>
          <p>{props.description}</p>
        </div>
      </article>
    </div>
  );
}

export function Benefit() {
  return (
    <section className={styles.benefits}>
      <div className="container">
        <div className="row">
          <div className={clsx("col-1-1", styles.title)}>
            <div>
              <span className="goffMainTitle">Why use feature flags?</span><br/>
              <p>
                Feature flags is a modern software engineering technique that configure select functionality during runtime, without deploying new code.
              </p>
            </div>
          </div>
        </div>
      </div>
      <div className="grid grid-pad">
        <BenefitCard
          img="img/benefits/rocket.jpg"
          title="Test in production"
          description="Test directly in production with your real data by enabling the features to your QA. Decrease incident by disabling the feature as soon as a bug arise." />

        <BenefitCard
          img="img/benefits/pm.jpg"
          title="Give autonomy to stakeholders"
          description="You don't need a software engineer to release a new feature, empower business stakeholders, no development skills are needed." />

        <BenefitCard
          img="img/benefits/inovate.jpg"
          title="Innovate faster"
          description="Deploy code when it is convenient (several times a day). Release when it is ready and it brings value. Deliver software to target audiences progressively." />

        <BenefitCard
          img="img/benefits/data.jpg"
          title="Experiment and learn"
          description="Try new features and measure their success while running A/B test. Export who was using which variation and learn what is successful for your business." />

        <BenefitCard
          img="img/benefits/devteam.jpg"
          title="Make engineers happy and productive"
          description="Have a better developer experience with simplifying how to release, test and deploy your software." />
      </div>
    </section>
  );
}
