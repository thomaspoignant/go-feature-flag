import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';
import LinkItem from '@theme/Footer/LinkItem';

function ColumnLinkItem({item}) {
  return item.html ? (
    <li
      className={clsx('footer__item', item.className)}
      // Developer provided the HTML, so assume it's safe.
      // eslint-disable-next-line react/no-danger
      dangerouslySetInnerHTML={{__html: item.html}}
    />
  ) : (
    <li key={item.href ?? item.to} className="footer__item">
      <LinkItem item={item} />
    </li>
  );
}

ColumnLinkItem.propTypes = {
  item: PropTypes.shape({
    html: PropTypes.string,
    className: PropTypes.string,
    href: PropTypes.string,
    to: PropTypes.string,
  }).isRequired,
};

function Column({column}) {
  return (
    <div className={clsx('col footer__col', column.className)}>
      <div className="footer__title">{column.title}</div>
      <ul className="footer__items clean-list">
        {column.items.map((item, i) => (
          <ColumnLinkItem key={i} item={item} />
        ))}
      </ul>
    </div>
  );
}

Column.propTypes = {
  column: PropTypes.shape({
    className: PropTypes.string,
    title: PropTypes.string,
    items: PropTypes.array.isRequired,
  }).isRequired,
};

export default function FooterLinksMultiColumn({columns}) {
  return (
    <div className="row footer__links">
      {columns.map((column, i) => (
        <Column key={i} column={column} />
      ))}
    </div>
  );
}

FooterLinksMultiColumn.propTypes = {
  columns: PropTypes.array.isRequired,
};
