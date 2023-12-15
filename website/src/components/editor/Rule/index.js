import clsx from 'clsx';
import styles from './styles.module.css';
import inputStyles from '../Input/styles.module.css';
import {Input} from '../Input';
import {Select} from '../Select';
import React from 'react';
import Link from '@docusaurus/Link';
import {useFormContext} from 'react-hook-form';
import 'react-sweet-progress/lib/style.css';
import {Colors} from '../Colors';
import 'react-datepicker/dist/react-datepicker.css';
import PropTypes from 'prop-types';
import {PercentagesForm} from './percentageForm';
import {ProgressiveRollout} from './progressiveRolloutForm';

Rule.propTypes = {
  variations: PropTypes.array,
  label: PropTypes.string.isRequired,
  isDefaultRule: PropTypes.bool,
};
export function Rule({variations, label, isDefaultRule}) {
  const {register, watch} = useFormContext();
  const otherOptions = [
    {value: 'percentage', displayName: '️↗️ a percentage rollout'},
    {value: 'progressive', displayName: '↗️ a progressive rollout'},
  ];

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

  const ruleOperators = [
    {
      value: '==',
      displayName: 'Equals To',
      displayText: 'eq',
    },
    {
      value: '!=',
      displayName: 'Not Equals To',
      displayText: 'ne',
    },
    {
      value: '<',
      displayName: 'Less Than',
      displayText: 'lt',
    },
    {
      value: '>',
      displayName: 'Greater Than',
      displayText: 'gt',
    },
    {
      value: '<=',
      displayName: 'Less Than Equal To',
      displayText: 'le',
    },
    {
      value: '>=',
      displayName: 'Greater Than Equal To',
      displayText: 'ge',
    },
    {
      value: 'co',
      displayName: 'Contains',
      displayText: 'co',
    },
    {
      value: 'sw',
      displayName: 'Starts With',
      displayText: 'sw',
    },
    {
      value: 'ew',
      displayName: 'Ends With',
      displayText: 'ew',
    },
    {
      value: 'in',
      displayName: 'In a List',
      displayText: 'in',
    },
    {
      value: 'pr',
      displayName: 'Present',
      displayText: 'pr',
    },
    {
      value: 'not',
      displayName: 'Not',
      displayText: 'not',
    },
  ];

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
        <div className={clsx('grid')}>
          <div className={'col-9-16'}>
            <div className={clsx('content', styles.inputQuery)}>
              <div className={styles.ifContainer}>
                <div className={clsx(styles.circle)}>IF</div>
              </div>
              <Input
                label={`${label}.key`}
                displayText="Key"
                required={true}
              />
              <div className={clsx(styles.operatorContainer)}>
                <Select
                  title="Operator"
                  content={ruleOperators}
                  register={register}
                  label={`${label}.operator`}
                  required={true}
                />
              </div>
              <Input
                label={`${label}.value`}
                displayText="Value"
                required={true}
              />
              <Link to={'/docs/configure_flag/rule_format'} target={'_blank'}>
                <i className="fa-regular fa-circle-question"></i>
              </Link>
            </div>
          </div>
        </div>
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
