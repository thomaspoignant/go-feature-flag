import {Controller, useFormContext} from 'react-hook-form';
import _ from 'lodash';
import styles from './styles.module.css';
import clsx from 'clsx';
import DatePicker from 'react-datepicker';
import {Select} from '../Select';
import {Colors} from '../Colors';
import {Input} from '../Input';
import React from 'react';
import PropTypes from 'prop-types';
import Link from '@docusaurus/Link';

ProgressiveStep.propTypes = {
  name: PropTypes.string,
  initialValue: PropTypes.number,
  label: PropTypes.string.isRequired,
  variations: PropTypes.object,
  defaultDate: PropTypes.any,
};
function ProgressiveStep({name, initialValue, label, variations, defaultDate}) {
  const {
    register,
    control,
    formState: {errors},
  } = useFormContext();

  function DisplayErrors() {
    const stepErrors = _.get(errors, label);
    if (_.isNil(stepErrors)) {
      return null;
    }

    return (
      <ul className={styles.formError}>
        {Object.keys(stepErrors).map(key => (
          <li key={key}>{stepErrors[key].message}</li>
        ))}
      </ul>
    );
  }

  return (
    <div>
      <div className={clsx('grid', styles.progressiveRollout)}>
        <div>{name}</div>
        <div>
          <Controller
            control={control}
            name={`${label}.date`}
            defaultValue={defaultDate}
            rules={{required: {value: true, message: 'Date field is required'}}}
            render={({field}) => (
              <DatePicker
                className={styles.dateInput}
                placeholderText="Select date"
                showTimeSelect
                onChange={date => field.onChange(date)}
                selected={field.value}
                dateFormat="Pp"
              />
            )}
          />
        </div>
        <div>and serve</div>
        <div>
          <Select
            title="Variation"
            content={
              variations
                .map((item, index) => {
                  return {
                    value: item.name,
                    displayName: `${Colors[index % Colors.length]} ${
                      item.name
                    }`,
                  };
                })
                .filter(
                  item => item.value !== undefined && item.value !== ''
                ) || []
            }
            register={register}
            label={`${label}.selectedVar`}
            required={true}
          />
        </div>
        <div className={styles.progressiveRolloutPercentage}>
          to&nbsp;
          <Input
            label={`${label}.percentage`}
            required={true}
            defaultValue={initialValue}
            type="number"
            displayText="%"
            className={styles.percentageInput}
            disablePlaceholder={true}
            disableInlineErr={true}
            validation={{
              valueAsNumber: true,
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
        </div>
        <div>
          <Link
            to={'/docs/configure_flag/rollout/progressive'}
            target={'_blank'}>
            <i className="fa-regular fa-circle-question"></i>
          </Link>
        </div>
      </div>
      <div>
        <DisplayErrors />
      </div>
    </div>
  );
}

ProgressiveRollout.propTypes = {
  variations: PropTypes.object,
  label: PropTypes.string.isRequired,
  selectedVar: PropTypes.string,
};
export function ProgressiveRollout({variations, label, selectedVar}) {
  if (selectedVar !== 'progressive') {
    return null;
  }

  let endDateDefault = new Date();
  endDateDefault = endDateDefault.setDate(endDateDefault.getDate() + 10);
  return (
    <div className={'grid grid-pad'}>
      <div className={clsx('col-1-1', styles.rolloutDesc)}>
        A progressive rollout allows you to increase the percentage of your flag
        over time.
        <br />
        You can select a release ramp where the percentage of your flag will
        increase progressively between the start date and the end date.
      </div>
      <ProgressiveStep
        name={'Start on the'}
        label={`${label}.initial`}
        variations={variations}
        initialValue={0}
        defaultDate={new Date()}
      />

      <ProgressiveStep
        name={'Stop on the'}
        label={`${label}.end`}
        variations={variations}
        initialValue={100}
        defaultDate={new Date(endDateDefault)}
      />
    </div>
  );
}
