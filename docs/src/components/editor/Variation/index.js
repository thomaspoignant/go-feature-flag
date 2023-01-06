import styles from './styles.module.css';
import {Input} from '../Input';
import clsx from 'clsx';
import React from 'react';
import {JsonEditor} from '../JsonEditor';
import {useFieldArray, useFormContext} from 'react-hook-form';
import {Colors} from '../Colors';
import PropTypes from 'prop-types';

Variations.propTypes = {
  type: PropTypes.string,
  label: PropTypes.string.isRequired,
};
export function Variations({type, label}) {
  const {control} = useFormContext();
  const {fields, append, remove} = useFieldArray({
    control,
    name: label,
    rules: {minLength: 1},
  });

  function removeVariation({ index }) {
    remove(index);
  }
  const handleOnClick = () => append({});

  return (
    <div>
      <h2>Variations</h2>
      <div className="grid grid-pad">
        {fields.map((field, index) => (
          <Variation
            type={type}
            key={field.id}
            label={`${label}.${index}`}
            index={index}
            remove={removeVariation}
            icon={Colors[index % Colors.length]}
          />
        ))}
      </div>
      <button className={styles.buttonPlus} onMouseDown={handleOnClick}>
        <span className="fa-stack fa-1x">
          <i className={clsx('fa-solid fa-circle fa-stack-2x', styles.bg)}></i>
          <i className="fa-solid fa-plus fa-stack-1x fa-inverse"></i>
        </span>
      </button>
    </div>
  );
}

Variation.propTypes = {
  type: PropTypes.string,
  label: PropTypes.string.isRequired,
  remove: PropTypes.func.isRequired,
  index: PropTypes.number.isRequired,
  icon: PropTypes.string,
};
function Variation({type, label, remove, index, icon}) {
  const {register} = useFormContext();
  const valueField = (type, label, register) => {
    const isJson = type && type.toUpperCase() === 'JSON';
    if (isJson) {
      return <JsonEditor register={register} required={true} label={label} />;
    }
    return (
      <Input
        displayText="Flag Value"
        label={label}
        register={register}
        type={type}
        required={true}
      />
    );
  };

  const handleOnClick = event => {
    event.preventDefault();
    remove(index);
  };
  return (
    <div className={styles.variation}>
      <div className={clsx('col-4-12 mobile-col-5-12', styles.icons)}>
        <div className={styles.icon}>{icon}</div>
        <Input
          displayText="Name"
          label={`${label}.name`}
          register={register}
          required={true}
        />
      </div>
      <div className={clsx('col-6-12 mobile-col-7-12')}>
        {valueField(type, label + '.value', register)}
      </div>
      <div className="col-1-12 mobile-col-1-12">
        {
          <button className={styles.buttonPlus} onMouseDown={handleOnClick}>
            <span className="fa-stack fa-1x">
              <i
                className={clsx(
                  'fa-solid fa-circle fa-stack-2x',
                  styles.bg
                )}></i>
              <i className="fa-solid fa-minus fa-stack-1x fa-inverse"></i>
            </span>
          </button>
        }
      </div>
    </div>
  );
}
