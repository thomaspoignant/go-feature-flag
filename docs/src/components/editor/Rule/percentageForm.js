import {useFormContext} from 'react-hook-form';
import {isArray} from 'redoc';
import styles from './styles.module.css';
import {Progress} from 'react-sweet-progress';
import clsx from 'clsx';
import {Input} from '../Input';
import {Colors} from '../Colors';
import React from 'react';
import PropTypes from 'prop-types';

PercentagesForm.propTypes = {
  variations: PropTypes.array,
  label: PropTypes.string.isRequired,
  selectedVar: PropTypes.string,
};
export function PercentagesForm({variations, label, selectedVar = ''}) {
  const {register, watch} = useFormContext();
  if (selectedVar !== 'percentage') {
    return null;
  }

  function computePercentages(percentages) {
    const sum = percentages
      .filter(item => item && !isNaN(item.value))
      .reduce(
        (accumulator, currentValue) => accumulator + currentValue.value,
        0
      );
    return sum;
  }

  return (
    <div className={'grid-pad grid'}>
      <div className={clsx('col-1-1', styles.rolloutDesc)}>
        A percentage rollout means that your users are divided in different
        buckets and you serve different variations to them. Note that a user
        will always have the same variation.
      </div>
      <div className={'col-1-2'}>
        <ul className={styles.percentageContainer}>
          {variations.map((field, index) => (
            <li key={`${field}`}>
              <Input
                label={`${label}.${index}.value`}
                required={true}
                defaultValue={0}
                type="number"
                displayText={`%  ${Colors[index % Colors.length]} ${
                  field.name
                }`}
                className={styles.percentageInput}
                disablePlaceholder={true}
                disableInlineErr={true}
                validation={{
                  valueAsNumber: true,
                  required: {
                    value: true,
                    message: 'Percentage field is required',
                  },
                  min: {
                    value: 0,
                    message: 'Percentage should be between 0 and 100',
                  },
                  max: {
                    value: 100,
                    message: 'Percentage should be between 0 and 100',
                  },
                }}
              />
              <input
                type="hidden"
                {...register(`${label}.${index}.name`)}
                value={field.name}
              />
            </li>
          ))}
        </ul>
        <ProgressBar percentages={watch(label)} />
      </div>
    </div>
  );
}

ProgressBar.propTypes = {
  percentages: PropTypes.array,
};
function ProgressBar({percentages}) {
  if (!percentages || !isArray(percentages) || percentages.length <= 0) {
    return null;
  }
  const sum = computePercentages(percentages);
  if (sum > 100) {
    return (
      <div className={styles.error}>
        The total percentage cannot be more than 100%
      </div>
    );
  }

  return <Progress percent={sum} />;
}
