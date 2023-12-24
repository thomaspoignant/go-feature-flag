import clsx from 'clsx';
import styles from './styles.module.css';
import inputStyles from '../Input/styles.module.css';
import * as ReactDnD from 'react-dnd';
import * as ReactDndHtml5Backend from 'react-dnd-html5-backend';
import {Input} from '../Input';
import {Select} from '../Select';
import React, {useCallback, useState} from 'react';
import {useFormContext} from 'react-hook-form';
import 'react-sweet-progress/lib/style.css';
import {Colors} from '../Colors';
import 'react-datepicker/dist/react-datepicker.css';
import PropTypes from 'prop-types';
import {PercentagesForm} from './percentageForm';
import {ProgressiveRollout} from './progressiveRolloutForm';
import {QueryBuilder} from 'react-querybuilder';
import {QueryBuilderDnD} from '@react-querybuilder/dnd';
import 'react-querybuilder/dist/query-builder.css';

Rule.propTypes = {
  variations: PropTypes.array,
  label: PropTypes.string.isRequired,
  isDefaultRule: PropTypes.bool,
};
export function Rule({variations, label, isDefaultRule}) {
  // TODO: Changes on field key removes the value
  const [query, setQuery] = useState({
    combinator: 'and',
    rules: [{field: '', operator: '==', value: ''}],
  });

  const {register, watch, setValue} = useFormContext();
  const otherOptions = [
    {value: 'percentage', displayName: '️↗️ a percentage rollout'},
    {value: 'progressive', displayName: '↗️ a progressive rollout'},
  ];

  // TODO: Validate query (look for ~)
  const parseQuery = useCallback(
    query => {
      setValue(`${label}.query`, parseJsonToCustomQuery(query));

      setQuery(query);
    },
    [parseJsonToCustomQuery, setValue, setQuery, label]
  );

  function getVariationList(variations) {
    const availableVariations =
      variations
        .map((item, index) => {
          return {
            value: item.name,
            displayName: `${Colors[index % Colors.length]} ${item.name}`,
          };
        })
        .filter(item => item.value !== undefined && item.value !== '') || [];
    return availableVariations;
  }

  function getSelectorList(variations) {
    const filteredVariations = getVariationList(variations);
    if (filteredVariations.length >= 2) {
      return [...filteredVariations, ...otherOptions];
    }
    return filteredVariations;
  }

  return (
    <div className={clsx('grid-pad grid', styles.ruleContainer)}>
      {!isDefaultRule && (
        <div className={'col-1-1'}>
          <div className={'content'}>
            <Input
              label={`${label}.name`}
              displayText={'Rule name'}
              className={clsx(
                inputStyles.editorInputContainer,
                styles.ruleName
              )}
              required={true}
            />
          </div>
        </div>
      )}
      {!isDefaultRule && (
        <>
          <input
            name={`${label}.query`}
            type="hidden"
            {...register(`${label}.query`)}
          />

          <QueryBuilderDnD dnd={{...ReactDnD, ...ReactDndHtml5Backend}}>
            <QueryBuilder
              controlElements={{fieldSelector: FieldSelector}}
              operators={ruleOperators}
              query={query}
              onQueryChange={parseQuery}
            />
          </QueryBuilderDnD>
        </>
      )}
      <div className={'col-5-12'}>
        <div className={clsx('content', styles.serve)}>
          <div className={styles.serveTitle}>Serve</div>
          <Select
            title="Variation"
            content={getSelectorList(variations)}
            register={register}
            label={`${label}.selectedVar`}
            required={true}
          />
        </div>
      </div>
      <div className={'col-1-1'}>
        <PercentagesForm
          selectedVar={watch(`${label}.selectedVar`)}
          variations={variations}
          label={`${label}.percentages`}
        />
        <ProgressiveRollout
          selectedVar={watch(`${label}.selectedVar`)}
          variations={variations}
          label={`${label}.progressive`}
        />
      </div>
    </div>
  );
}

const FieldSelector = ({
  className,
  handleOnChange,
  title,
  value,
  disabled,
  testID,
}) => (
  <input
    data-testid={testID}
    type="text"
    className={className}
    // TODO: remove this hack and fix the issue with the query builder
    value={value === '~' ? '' : value}
    title={title}
    disabled={disabled}
    onChange={e => handleOnChange(e.target.value)}
  />
);

/**
 * Parses a JSON object into a custom query language based on the nikunjy/rules library.
 *
 * @param {Object} json - The JSON object representing the query.
 * @returns {string} - The custom query string.
 */
function parseJsonToCustomQuery(json) {
  /**
   * Recursive helper function to process rules.
   *
   * @param {Object} rule - The rule object to process.
   * @returns {string} - The custom query string for the rule.
   */
  function processRule(rule) {
    let query = '';

    if (rule.field && rule.operator) {
      query += `${rule.field} ${rule.operator}`;

      if (rule.value) {
        query += ` ${rule.value}`;
      }
    }

    if (rule.rules && rule.rules.length > 0) {
      const subRules = rule.rules.map(processRule).join(` ${rule.combinator} `);
      query += ` (${subRules}) `;
    }

    return query.trim();
  }

  if (!json.combinator || !json.rules || !Array.isArray(json.rules)) {
    throw new Error('Invalid JSON format for the query.');
  }

  return json.rules.map(processRule).join(` ${json.combinator} `);
}

const ruleOperators = [
  {name: '==', label: 'Equals To'},
  {name: '!=', label: 'Not Equals To'},
  {name: '<', label: 'Less Than'},
  {name: '>', label: 'Greater Than'},
  {name: '<=', label: 'Less Than Equal To'},
  {name: '>=', label: 'Greater Than Equal To'},
  {name: 'co', label: 'Contains'},
  {name: 'sw', label: 'Starts With'},
  {name: 'ew', label: 'Ends With'},
  {name: 'in', label: 'In a List'},
  {name: 'pr', label: 'Present'},
  {name: 'not', label: 'Not'},
];
