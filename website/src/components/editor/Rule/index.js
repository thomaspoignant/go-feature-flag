import clsx from 'clsx';
import styles from './styles.module.css';
import inputStyles from '../Input/styles.module.css';
import * as ReactDnD from 'react-dnd';
import * as ReactDndHtml5Backend from 'react-dnd-html5-backend';
import {Input} from '../Input';
import {Select} from '../Select';
import React, {useCallback, useState, useMemo} from 'react';
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
    rules: [{field: '', operator: 'eq', value: ''}],
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
              fields={[]}
              controlElements={{
                fieldSelector: FieldSelector,
                valueEditor: FieldSelector,
                operatorSelector: OperatorSelector,
                combinatorSelector: CombinatorSelector,
                addGroupAction: AddGroupAction,
                addRuleAction: AddRuleAction,
                removeGroupAction: RemoveGroupAction,
                removeRuleAction: RemoveRuleAction,
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
function FieldSelector({handleOnChange, title, value, disabled, testID}) {
  return (
    <div>
      <Input
        data-testid={testID}
        value={value}
        label={title}
        displayText={title}
        disabled={disabled}
        controlled={true}
        required={true}
        onChange={e => handleOnChange(e.target.value)}
      />
    </div>
  );
}

OperatorSelector.propTypes = {
  handleOnChange: PropTypes.func,
  options: PropTypes.array,
  title: PropTypes.string,
};
function OperatorSelector({options, handleOnChange, title}) {
  const content = useMemo(
    () =>
      options.map(({name: value, label: displayName}) => ({
        value,
        displayName,
      })),
    [options]
  );

  return (
    <Select
      content={content}
      controlled={true}
      label="Operator"
      onChange={e => handleOnChange(e.target.value)}
      required={false}
      title={title}
    />
  );
}

CombinatorSelector.propTypes = {
  handleOnChange: PropTypes.func,
  options: PropTypes.array,
  title: PropTypes.string,
};
function CombinatorSelector({options, handleOnChange, title}) {
  const content = useMemo(
    () =>
      options.map(({name: value, label: displayName}) => ({
        value,
        displayName,
      })),
    [options]
  );

  return (
    <div style={{maxWidth: '8rem'}}>
      <Select
        content={content}
        controlled={true}
        label="Operator"
        onChange={e => handleOnChange(e.target.value)}
        required={false}
        title={title}
      />
    </div>
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

AddAction.propTypes = {
  handleOnClick: PropTypes.func,
  variant: PropTypes.string,
};
function AddAction({handleOnClick, variant}) {
  const label = useMemo(() => {
    switch (variant) {
      case 'group':
        return '+Group';
      case 'rule':
        return '+Rule';
    }
  }, [variant]);

  return (
    <button
      className="pushy__btn pushy__btn--md pushy__btn--black"
      onClick={handleOnClick}>
      {label}
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
        if (rule.operator === 'in') {
          query += ` ${convertToFormattedArray(rule.value)}`;
        } else {
          query += isNumeric(rule.value)
            ? ` ${rule.value}`
            : ` "${rule.value}"`;
        }
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

/**
 * Checks if a given value is numeric.
 * @param {string} value - The value to check.
 * @returns {boolean} Returns true if the value is numeric, false otherwise.
 */
function isNumeric(value) {
  return /^-?\d+$/.test(value);
}

/**
 * Converts a comma-separated string of numbers and strings into a formatted array string.
 * @param {string} input - The input string containing comma-separated values.
 * @returns {string} Returns the formatted array string.
 *
 * Example usage:
 * const input1 = '1,2,3,4,5';
 * const input2 = '1,"UUID-345-1234",3';
 *
 * const output1 = convertToFormattedArray(input1);
 * const output2 = convertToFormattedArray(input2);
 *
 * console.log(output1); // Output: [1,2,3,4,5]
 * console.log(output2); // Output: [1,"UUID-345-1234",3]
 */
function convertToFormattedArray(input) {
  const elements = input.split(',');

  const formattedArray = elements.map(element => {
    const trimmedElement = element.trim();

    if (isNumeric(trimmedElement)) {
      return parseInt(trimmedElement, 10); // Ensure to specify the radix when parsing integers.
    } else {
      // Remove double quotes around string elements
      return trimmedElement.replace(/^"(.*)"$/, '$1');
    }
  });

  return JSON.stringify(formattedArray);
}
function AddGroupAction({handleOnClick}) {
  return <AddAction handleOnClick={handleOnClick} variant="group" />;
}

function AddRuleAction({handleOnClick}) {
  return <AddAction handleOnClick={handleOnClick} variant="rule" />;
}

function RemoveGroupAction({handleOnClick}) {
  return <RemoveAction handleOnClick={handleOnClick} variant="group" />;
}

function RemoveRuleAction({handleOnClick}) {
  return <RemoveAction handleOnClick={handleOnClick} variant="rule" />;
}

const ruleOperators = [
  {name: 'eq', label: 'Equals To'},
  {name: 'ne', label: 'Not Equals To'},
  {name: 'lt', label: 'Less Than'},
  {name: 'gt', label: 'Greater Than'},
  {name: 'le', label: 'Less Than Equal To'},
  {name: 'ge', label: 'Greater Than Equal To'},
  {name: 'co', label: 'Contains'},
  {name: 'sw', label: 'Starts With'},
  {name: 'ew', label: 'Ends With'},
  {name: 'in', label: 'In a List'},
  {name: 'pr', label: 'Present'},
  {name: 'not', label: 'Not'},
];
