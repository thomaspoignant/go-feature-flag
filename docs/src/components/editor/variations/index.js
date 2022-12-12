import styles from "./styles.module.css";
import {Input} from "../Input";
import clsx from "clsx";
import React from "react";
import {JsonEditor} from "../jsonEditor";
import {useFieldArray, useForm} from "react-hook-form"

export function Variations({type, register, control, label, colors}){
  const { fields, append, remove} = useFieldArray({
    control,
    name: label,
    rules: { minLength: 1 }
  });

  const removeVariation = (index) => remove(index)
  const handleOnClick = ()=> append({value : ""})

  return(
    <div>
      <h2>Variations</h2>
      <div className="grid grid-pad">
        {fields.map((field, index) => (
          <Variation type={type}
                     key={field.id}
                     register={register}
                     label={`${label}.${index}`}
                     index={index}
                     remove={removeVariation}
                     icon={colors[index%colors.length]}/>
        ))}
      </div>
      <button className={styles.buttonPlus} onMouseDown={handleOnClick}>
          <span className="fa-stack fa-1x">
            <i className={clsx("fa-solid fa-circle fa-stack-2x", styles.bg)}></i>
            <i className="fa-solid fa-plus fa-stack-1x fa-inverse"></i>
          </span>
      </button>
    </div>);
}

function Variation({type, register, label, remove, index, icon}){
  const valueField = (type, label, register) => {
    const isJson = type && type.toUpperCase() === 'JSON';
    if(isJson){
      return <JsonEditor register={register} required={true} label={label} />
    }
    return <Input displayText="Value"
                  label={label}
                  register={register}
                  required={true}/>
  }

  const handleOnClick = ()=> remove(index)

  return(
    <div className={styles.variation}>
      <div className="col-3-12 mobile-col-3-12">
        <Input displayText="Name"
               label={`${label}.name`}
               register={register}
               required={true}/>
      </div>
      <div className="col-3-12 mobile-col-3-12">
        {valueField(type, label+'.value', register)}
      </div>
      <div className={clsx("col-5-12 mobile-col-5-12", styles.icons)}>
        <Input id="description"
               displayText="Description"
               label={`${label}.description`}
               register={register}
               required={false}/>
        <div className={styles.icon}>{icon}</div>
      </div>
      <div className="col-1-12 mobile-col-1-12">
        <button className={styles.buttonPlus} onMouseDown={handleOnClick}>
          <span className="fa-stack fa-1x">
            <i className={clsx("fa-solid fa-circle fa-stack-2x", styles.bg)}></i>
            <i className="fa-solid fa-minus fa-stack-1x fa-inverse"></i>
          </span>
        </button>
      </div>
    </div>
  );
}
