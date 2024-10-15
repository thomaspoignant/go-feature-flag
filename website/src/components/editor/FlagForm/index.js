import styles from './styles.module.css';
import {Input} from '../Input';
import React from 'react';
import {Switch} from '../Switch';
import {Select} from '../Select';
import {Variations} from '../Variation';
import {useFormContext} from 'react-hook-form';
import {Targeting} from '../Targeting';
import {Rule} from '../Rule';
import PropTypes from 'prop-types';
import {Metadata} from '../Metadata';
import {FlagTest} from '../FlagTest';

FlagForm.propTypes = {
  label: PropTypes.string.isRequired,
};

export function FlagForm({label}) {
  const {watch} = useFormContext();
  const typeSelectorContent = [
    {value: 'boolean', displayName: '☑️ boolean'},
    {value: 'string', displayName: '🔤 string'},
    {value: 'number', displayName: '🔢 number'},
    {value: 'json', displayName: '🖥 JSON'},
  ];

  return (
    <div className={styles.flagContainer}>
      <div className="grid grid-cols-12">
        <div className="xl:col-span-6 col-span-2 mb-4">
          <Input
            displayText="Flag Name"
            label={`${label}.flagName`}
            required={true}
          />
        </div>
        <div className="col-span-3">
          <Switch
            id="disable"
            label={`${label}.disable`}
            displayText="Disable"
          />
        </div>
        <div className="col-span-3">
          <Switch
            id="disable"
            label={`${label}.trackEvents`}
            displayText="Track event"
            defaultChecked={true}
          />
        </div>
        <div className={'col-span-3'}>
          <Select
            title="Flag type"
            content={typeSelectorContent}
            label={`${label}.type`}
            required={true}
          />
        </div>
        <div className="col-start-5 col-span-3">
          <Input
            id="version"
            label={`${label}.version`}
            displayText="Version"
          />
        </div>
      </div>
      <Variations label={`${label}.variations`} type={watch(`${label}.type`)} />
      <Targeting
        label={`${label}.targeting`}
        variations={watch(`${label}.variations`)}
      />
      <div>
        <h2>Default</h2>
        <Rule
          label={`${label}.defaultRule`}
          variations={watch(`${label}.variations`)}
          isDefaultRule={true}
        />
      </div>
      <Metadata label={`${label}.metadata`} />
      <FlagTest flagInfo={label} />
    </div>
  );
}
