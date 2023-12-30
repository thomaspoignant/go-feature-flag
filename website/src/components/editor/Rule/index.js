import clsx from 'clsx';
import styles from './styles.module.css';
import inputStyles from '../Input/styles.module.css';
import * as ReactDnD from 'react-dnd';
import * as ReactDndHtml5Backend from 'react-dnd-html5-backend';
import {Input} from '../Input';
import {Select} from '../Select';
import React, {useCallback, useState, useEffect} from 'react';
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
import './query-builder.css';

Rule.propTypes = {
  variations: PropTypes.array,
  label: PropTypes.string.isRequired,
  isDefaultRule: PropTypes.bool,
};
export function Rule({variations, label, isDefaultRule}) {
  const [query, setQuery] = useState({
    combinator: 'and',
    rules: [{field: '', operator: '==', value: ''}],
  });

  const {register, watch, setValue} = useFormContext();
  const otherOptions = [
    {value: 'percentage', displayName: '️↗️ a percentage rollout'},
    {value: 'progressive', displayName: '↗️ a progressive rollout'},
  ];

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
              controlElements={{
                fieldSelector: FieldSelector,
                removeGroupAction: ({handleOnClick}) => (
                  <RemoveAction handleOnClick={handleOnClick} variant="group" />
                ),
                removeRuleAction: ({handleOnClick}) => (
                  <RemoveAction handleOnClick={handleOnClick} variant="rule" />
                ),
              }}
              resetOnFieldChange={false}
              resetOnOperatorChange={false}
              addRuleToNewGroups={true}
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

FieldSelector.propTypes = {
  className: PropTypes.string,
  handleOnChange: PropTypes.func,
  title: PropTypes.string,
  value: PropTypes.string,
  disabled: PropTypes.bool,
  testID: PropTypes.string,
};
function FieldSelector({
  className,
  handleOnChange,
  title,
  value,
  disabled,
  testID,
}) {
  useEffect(() => handleOnChange(''), []);

  return (
    <input
      data-testid={testID}
      type="text"
      className={className}
      value={value}
      title={title}
      disabled={disabled}
      onChange={e => handleOnChange(e.target.value)}
    />
  );
}

RemoveAction.propTypes = {
  handleOnClick: PropTypes.func,
  variant: PropTypes.string,
};
function RemoveAction({handleOnClick, variant}) {
  const getIcon = useCallback(() => {
    switch (variant) {
      case 'group':
        return 'fa-xmark';
      case 'rule':
        return 'fa-minus';
    }
  }, [variant]);

  return (
    <button className={styles.removeButton} onClick={handleOnClick}>
      <span className="fa-stack fa-1x">
        <i className={clsx('fa-solid fa-circle fa-stack-2x', styles.bg)}></i>
        <i className={`fa-solid ${getIcon()} fa-stack-1x fa-inverse`}></i>
      </span>
    </button>
  );
}

/**
 * Parses a JSON object into a custom query language based on the nikunjy/rules library.
 *
 * @param {Object} json - The JSON object representing the query.
 *   @property {string} combinator - The combinator used to combine multiple rules (e.g., "AND", "OR").
 *   @property {Array} rules - An array of rule objects.
 *     @property {string} field - The field to apply the rule on.
 *     @property {string} operator - The operator to use for the comparison.
 *     @property {string} [value] - The value to compare when the operator requires it.
 *     @property {string} [combinator] - The combinator used to combine multiple sub-rules within this rule (e.g., "AND", "OR").
 *     @property {Array} [rules] - An array of sub-rule objects.
 *
 * @returns {string} - The custom query string.
 */
function parseJsonToCustomQuery(json) {
  /**
   * Recursive helper function to process rules.
   *
   * @param {Object} rule - The rule object to process.
   *   @property {string} field - The field to apply the rule on.
   *   @property {string} operator - The operator to use for the comparison.
   *   @property {string} [value] - The value to compare when the operator requires it.
   *   @property {string} [combinator] - The combinator used to combine multiple sub-rules within this rule (e.g., "AND", "OR").
   *   @property {Array} [rules] - An array of sub-rule objects.
   *
   * @returns {string} - The custom query string for the rule.
   */
  function processRule(rule) {
    let query = '';

    if (rule.field && rule.operator) {
      query += `${rule.field} ${rule.operator}`;

      if (rule.value) {
        if (rule.operator == 'in') query += ` [${rule.value}]`;
        else query += ` ${rule.value}`;
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
