import React from 'react';
import PropTypes from 'prop-types';
import Link from '@docusaurus/Link';

const CARD_CLASS =
  'flex h-full flex-col rounded-2xl border border-solid border-gray-200 bg-white p-8 shadow-sm transition-shadow duration-200 hover:shadow-md dark:border-gray-700 dark:bg-[#1f2024]';

const COLUMN_CLASS = {
  2: 'sm:grid-cols-2',
  3: 'sm:grid-cols-2 lg:grid-cols-3',
  4: 'sm:grid-cols-2 lg:grid-cols-4',
};

function CardBody({icon, title, description, link, linkLabel}) {
  return (
    <>
      {icon && (
        <div className="mb-6 flex h-12 w-12 items-center justify-center text-4xl text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] [&>svg]:h-full [&>svg]:w-full">
          {icon}
        </div>
      )}
      <h3 className="m-0 text-2xl font-bold leading-snug text-gray-800 dark:text-gray-50">
        {title}
      </h3>
      {description && (
        <p className="mb-0 mt-3 leading-relaxed text-[color:var(--goff-main-ff-description)]">
          {description}
        </p>
      )}
      {link && (
        <span className="mt-auto inline-flex items-center gap-1 pt-5 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)]">
          {linkLabel} <span aria-hidden="true">→</span>
        </span>
      )}
    </>
  );
}

CardBody.propTypes = {
  icon: PropTypes.node,
  title: PropTypes.node.isRequired,
  description: PropTypes.node,
  link: PropTypes.string,
  linkLabel: PropTypes.node,
};

function Card({icon, title, description, link, linkLabel}) {
  if (link) {
    return (
      <Link
        to={link}
        className={`${CARD_CLASS} no-underline hover:no-underline`}>
        <CardBody
          icon={icon}
          title={title}
          description={description}
          link={link}
          linkLabel={linkLabel}
        />
      </Link>
    );
  }
  return (
    <div className={CARD_CLASS}>
      <CardBody icon={icon} title={title} description={description} />
    </div>
  );
}

Card.propTypes = CardBody.propTypes;

export default function Cards({title, cards, columns}) {
  const colClass = COLUMN_CLASS[columns] ?? COLUMN_CLASS[3];
  return (
    <section className="px-6 py-12 lg:px-16 xl:px-64 2xl:px-96">
      {title && (
        <h2 className="mb-10 text-center text-3xl font-bold text-gray-800 dark:text-gray-50">
          {title}
        </h2>
      )}
      <div className={`grid grid-cols-1 gap-6 ${colClass}`}>
        {cards.map(card => (
          <Card
            key={typeof card.title === 'string' ? card.title : card.id}
            icon={card.icon}
            title={card.title}
            description={card.description}
            link={card.link}
            linkLabel={card.linkLabel ?? 'Learn more'}
          />
        ))}
      </div>
    </section>
  );
}

Cards.propTypes = {
  // Optional heading rendered above the grid of cards.
  title: PropTypes.node,
  // Number of columns on large screens (2, 3 or 4). Defaults to 3.
  columns: PropTypes.oneOf([2, 3, 4]),
  // The cards to render. Each card: {icon?, title, description?, link?, linkLabel?}.
  cards: PropTypes.arrayOf(
    PropTypes.shape({
      icon: PropTypes.node,
      title: PropTypes.node.isRequired,
      description: PropTypes.node,
      link: PropTypes.string,
      linkLabel: PropTypes.node,
      // Optional explicit key, used when title is not a plain string.
      id: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
    })
  ).isRequired,
};
