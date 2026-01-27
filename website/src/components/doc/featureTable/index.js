import React from 'react';
import PropTypes from 'prop-types';

/**
 * FeatureTable component that displays a table of features for a given SDK
 * @param {Object} sdk - The SDK object containing featureList array
 * @param {string} sdk.name - The name of the SDK
 * @param {Array} sdk.featureList - Array of feature objects with name and status
 * @param {string} sdk.featureList[].name - The name of the feature
 * @param {string} sdk.featureList[].status - The status of the feature (done, not implemented, in progress, etc.)
 * @param {string} sdk.featureList[].description - Optional description of the feature
 * @returns {JSX.Element} A table showing the features and their status
 */
export const FeatureTable = ({sdk}) => {
  if (!sdk || !sdk.featureList || !Array.isArray(sdk.featureList)) {
    return null;
  }

  const getStatusIcon = status => {
    switch (status.toLowerCase()) {
      case 'done':
        return <i className="fa-solid fa-circle-check text-green-500"></i>;
      case 'WIP':
        return <i className="fa-solid fa-clock text-yellow-500 mr-1"></i>;
      default:
        return <i className="fa-solid fa-circle-xmark text-red-500"></i>;
    }
  };

  return (
    <div className="feature-table my-6">
      <table className="w-full border-collapse border border-gray-300 dark:border-gray-600">
        <thead>
          <tr className="bg-gray-100 dark:bg-gray-700">
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-2 text-left font-semibold">
              Status
            </th>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-2 text-left font-semibold">
              Feature
            </th>
            <th className="border border-gray-300 dark:border-gray-600 px-4 py-2 text-left font-semibold">
              Description
            </th>
          </tr>
        </thead>
        <tbody>
          {sdk.featureList
            .sort((a, b) => a.status.localeCompare(b.status))
            .map(feature => (
              <tr
                key={feature.name}
                className="hover:bg-gray-50 dark:hover:bg-gray-800">
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-2 text-center">
                  {getStatusIcon(feature.status)}
                </td>
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-2 font-medium">
                  {feature.name}
                </td>
                <td className="border border-gray-300 dark:border-gray-600 px-4 py-2">
                  {feature.description ?? 'N/A'}
                </td>
              </tr>
            ))}
        </tbody>
      </table>

      <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
        <span className="mr-4">
          <i className="fa-solid fa-circle-check text-green-500 mr-1" />
          Implemented
        </span>
        <span className="mr-4">
          <i className="fa-solid fa-clock text-yellow-500 mr-1" />
          In-progress
        </span>
        <span>
          <i className="fa-solid fa-circle-xmark text-red-500 mr-1" />
          Not implemented yet
        </span>
      </div>
    </div>
  );
};

FeatureTable.propTypes = {
  sdk: PropTypes.shape({
    name: PropTypes.string,
    featureList: PropTypes.arrayOf(
      PropTypes.shape({
        name: PropTypes.string.isRequired,
        status: PropTypes.string.isRequired,
        description: PropTypes.string,
      })
    ),
  }).isRequired,
};

export default FeatureTable;
