import React, {useCallback} from 'react';
import {useFieldArray, useFormContext} from 'react-hook-form';
import {Rule} from '../Rule';
import styles from './styles.module.css';
import clsx from 'clsx';
import PropTypes from 'prop-types';

Targeting.propTypes = {
  label: PropTypes.string.isRequired,
  variations: PropTypes.array,
};
export function Targeting({label, variations}) {
  const {control} = useFormContext();
  const {fields, append, remove} = useFieldArray({
    control,
    name: label,
  });

  const addNewRule = useCallback(
    () => append({name: `Rule ${fields.length + 1}`}),
    []
  );
  const removeCurrent = useCallback(index => remove(index), []);

  return (
    <div>
      <h2>Target specific users</h2>
      {fields.length <= 0 && <div>Add Rule</div>}
      {fields.map((field, index) => (
        <div key={field.id} className={clsx(styles.targeting)}>
          <Rule
            variations={variations}
            label={`${label}.${index}`}
            isDefaultRule={false}
          />
          <button className={styles.button} onMouseDown={removeCurrent}>
            <span className="fa-stack fa-1x">
              <i
                className={clsx(
                  'fa-solid fa-circle fa-stack-2x',
                  styles.bg
                )}></i>
              <i className="fa-solid fa-xmark fa-stack-1x fa-inverse" />
            </span>
          </button>
        </div>
      ))}
      <button className={styles.button} onMouseDown={addNewRule}>
        <span className="fa-stack fa-1x">
          <i className={clsx('fa-solid fa-circle fa-stack-2x', styles.bg)}></i>
          <i className="fa-solid fa-plus fa-stack-1x fa-inverse"></i>
        </span>
      </button>
    </div>
  );
}
