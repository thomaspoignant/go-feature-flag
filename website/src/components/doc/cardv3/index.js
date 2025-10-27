import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';

export const Card = ({
  logo,
  name,
  link,
  description,
  cssLogo,
  badge,
  featureList,
}) => {
  return (
    <div className="max-w-xs p-6 bg-white border border-gray-200 rounded-lg shadow dark:bg-gray-800 dark:border-gray-700">
      {logo && <img src={logo} className={'min-h-20 max-h-20'} alt={'logo'} />}
      {!logo && cssLogo && <i className={clsx(cssLogo, 'text-6xl')}></i>}
      <a href={link}>
        <h5 className="mb-2 text-2xl font-semibold tracking-tight text-gray-900 dark:text-white">
          {name}
        </h5>
      </a>
      <p className={'font-normal text-gray-700 dark:text-gray-400'}>
        {description}
      </p>
      {badge && (
        <p>
          <img src={badge} alt={'badge'} />
        </p>
      )}

      {featureList && (
        <p>
          <ul className={'list-none pl-0'}>
            {featureList
              .sort((a, b) => {
                // Sort 'done' features first, then others
                if (a.status === 'done' && b.status !== 'done') return -1;
                if (a.status !== 'done' && b.status === 'done') return 1;
                return 0;
              })
              .map((feature, index) => (
                <li
                  key={`${index}.${feature.name}`}
                  className="flex items-center">
                  {feature.status === 'done' ? (
                    <i className="fa-solid fa-circle-check text-green-500 mr-1.5"></i>
                  ) : (
                    <i className="fa-solid fa-circle-xmark text-red-500 mr-1.5"></i>
                  )}

                  <span>{feature.name}</span>
                </li>
              ))}
          </ul>
        </p>
      )}
      <a
        href={link}
        className="inline-flex font-medium items-center text-blue-600 hover:underline">
        More details <i className="ml-2 fa-solid fa-up-right-from-square"></i>
      </a>
    </div>
  );
};

Card.propTypes = {
  logo: PropTypes.string,
  cssLogo: PropTypes.string,
  name: PropTypes.string,
  link: PropTypes.string,
  description: PropTypes.string,
  badge: PropTypes.string,
  featureList: PropTypes.array,
};
