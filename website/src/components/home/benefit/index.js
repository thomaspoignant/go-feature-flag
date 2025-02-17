import React from 'react';
import styles from './styles.module.css';
import PropTypes from 'prop-types';

BenefitCard.propTypes = {
  img: PropTypes.string.isRequired,
  title: PropTypes.string.isRequired,
  description: PropTypes.string.isRequired,
};

function BenefitCard({img, title, description}) {
  return (
    <div className="w-fit bg-white border border-gray-200 rounded-lg shadow-sm dark:bg-gray-800 dark:border-gray-700">
      <img className="rounded-t-lg p-3" src={img} alt={title} />
      <div className="p-5">
        <h5 className="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
          {title}
        </h5>
        <p className="mb-3 font-normal text-gray-700 dark:text-gray-400">
          {description}
        </p>
      </div>
    </div>
  );
}

export function Benefit() {
  return (
    <section className={styles.benefits}>
      <div className={'text-center mb-5'}>
        <span className="goffMainTitle">The Benefits of Feature Flags</span>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-2 px-3">
        <BenefitCard
          img="img/benefits/rocket.jpg"
          title="Test in production"
          description="Test directly in production with your real data by enabling the features to your QA. Decrease incident by disabling the feature as soon as a bug arise."
        />

        <BenefitCard
          img="img/benefits/pm.jpg"
          title="Give autonomy to stakeholders"
          description="You don't need a software engineer to release a new feature, empower business stakeholders, no development skills are needed."
        />

        <BenefitCard
          img="img/benefits/inovate.jpg"
          title="Innovate faster"
          description="Deploy code when it is convenient (several times a day). Release when it is ready and it brings value. Deliver software to target audiences progressively."
        />

        <BenefitCard
          img="img/benefits/data.jpg"
          title="Experiment and learn"
          description="Try new features and measure their success while running A/B test. Export who was using which variation and learn what is successful for your business."
        />

        <BenefitCard
          img="img/benefits/devteam.jpg"
          title="Make engineers happy and productive"
          description="Have a better developer experience with simplifying how to release, test and deploy your software."
        />
      </div>
    </section>
  );
}
