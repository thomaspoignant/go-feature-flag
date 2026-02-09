import React from 'react';
import PropTypes from 'prop-types';

export const Vignette = ({color, link, title, description, icon}) => {
  const colors = {
    cyan: {
      href: 'block bg-gradient-to-br from-cyan-50 to-sky-100 dark:from-cyan-900/20 dark:to-sky-900/20 border border-cyan-200 dark:border-cyan-700 rounded-lg p-6 hover:shadow-lg transition-all duration-200 hover:scale-105 hover:border-cyan-300 dark:hover:border-cyan-600',
      icon: 'bg-cyan-500 text-white p-3 rounded-lg mr-4',
      link: 'inline-flex items-center mt-4 text-cyan-600 hover:text-cyan-800 dark:text-cyan-400 dark:hover:text-cyan-300 font-medium no-underline',
    },
    blue: {
      href: 'block bg-gradient-to-br from-blue-50 to-slate-100 dark:from-blue-900/20 dark:to-slate-800/20 border border-blue-200 dark:border-blue-700 rounded-lg p-6 hover:shadow-lg transition-all duration-200 hover:scale-105 hover:border-blue-300 dark:hover:border-blue-600',
      icon: 'bg-blue-600 text-white p-3 rounded-lg mr-4',
      link: 'inline-flex items-center mt-4 text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 font-medium no-underline',
    },
  };

  return (
    <a href={link} className={colors[color].href}>
      <div className="flex items-center mb-4">
        <div className={colors[color].icon}>
          <i className={`fa-solid ${icon} text-2xl`}></i>
        </div>
        <h3 className="text-lg font-semibold no-underline text-gray-900 dark:text-gray-100">
          {title}
        </h3>
      </div>
      <p className="text-gray-700 dark:text-gray-300 leading-relaxed">
        {description}
      </p>
      <div className={colors[color].link}>
        Get started <i className="fa-solid fa-circle-arrow-right ml-2"></i>
      </div>
    </a>
  );
};

Vignette.propTypes = {
  color: PropTypes.oneOf(['cyan', 'blue']),
  link: PropTypes.string.isRequired,
  title: PropTypes.string.isRequired,
  description: PropTypes.string.isRequired,
  icon: PropTypes.string.isRequired,
};

Vignette.defaultProps = {
  color: 'blue',
};
