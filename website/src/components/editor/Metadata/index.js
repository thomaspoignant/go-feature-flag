import styles from './styles.module.css';
import {Input} from '../Input';
import clsx from 'clsx';
import React from 'react';
import {useFieldArray, useFormContext} from 'react-hook-form';
import PropTypes from 'prop-types';

Metadata.propTypes = {
  label: PropTypes.string.isRequired,
};
export function Metadata({label}) {
  const {control} = useFormContext();
  const {fields, append, remove} = useFieldArray({
    control,
    name: label,
    rules: {minLength: 1},
  });

  const removeMetadataItem = index => remove(index);
  const handleOnClick = () => append({name: '', value: ''});

  return (
    <div>
      <h2>Metadata</h2>
      {fields.length <= 0 && <div>Add new metadata to your flag</div>}
      <div className="grid grid-pad">
        {fields.map((field, index) => (
          <MetadataItem
            key={field.id}
            label={`${label}.${index}`}
            index={index}
            remove={removeMetadataItem}
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

function MetadataItem({type, label, remove, index, icon}) {
  const {register} = useFormContext();
  const handleOnClick = event => {
    event.preventDefault();
    remove(index);
  };
  return (
    <div className={styles.variation}>
      <div className={clsx('col-4-12 mobile-col-5-12', styles.icons)}>
        <div className={styles.icon}>{icon}</div>
        <Input displayText="Key" label={`${label}.name`} register={register} />
      </div>
      <div className={clsx('col-6-12 mobile-col-7-12')}>
        <Input
          displayText="Value"
          label={`${label}.value`}
          register={register}
          type={type}
        />
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

MetadataItem.propTypes = {
  label: PropTypes.string.isRequired,
  remove: PropTypes.func.isRequired,
  index: PropTypes.number.isRequired,
  type: PropTypes.string,
  icon: PropTypes.node,
};
