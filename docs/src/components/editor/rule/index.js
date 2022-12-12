import clsx from "clsx";
import styles from "./styles.module.css";
import {Input} from "../Input";
import {Select} from "../Select";
import React, {useEffect, useState} from "react";
import Link from "@docusaurus/Link";
import {useFieldArray, useFormContext} from "react-hook-form";

export function Rule({control, watch, register, variationsLabel, label, colors, isDefaultRule, setValue}){
  const otherOptions = [
    { value: "percentage", "displayName": "️↗️ a percentage rollout"},
    { value: "progressive", "displayName": "↗️ a progressive rollout"},
  ];
  function getVariationList(variationsLabel, colors){
    const availableVariations = watch(variationsLabel)
      .map((item, index) => {
        return {"value": item.name, "displayName": `${colors[index%colors.length]} ${item.name}`}
      }).filter(item => item.value !== undefined && item.value !== '') || [];
    return availableVariations;
  }

  function Query({label, register}) {
    return(
      <div className={clsx("grid")}>
        <div className={"col-9-12"}>
          <div className={clsx("content", styles.inputQuery)}>
            <div className={styles.ifContainer}>
              <div className={clsx(styles.circle)}>IF</div>
            </div>
            <Input
              label={`${label}`}
              register={register}
              displayText={"Query"}
            />
            <Link to={'/docs/configure_flag/rule_format'} target={"_blank"}>
              <i className="fa-regular fa-circle-question"></i>
            </Link>
          </div>
        </div>
      </div>
    );
  }
  // function PercentageRollout({label, register, variationsLabel, variations}) {
  //   console.log(variations);
  //   return(
  //     <div className={clsx("grid")}>
  //       <div className={"col-9-12"}>
  //         <div className={clsx("content")}>
  //           <ul>
  //             { variations.map((field, index) => (
  //                 <li key={`${label}.${index}`} >
  //                   <input type="number" {...register(`${label}.${index}.value`, false)} />
  //                   {field.name}
  //                 </li>
  //             ))}
  //           </ul>
  //         </div>
  //       </div>
  //     </div>);
  // }

  function shouldDisplayPercentage(){
    console.log(watch(`${label}.variation`));
    return watch(`${label}.variation`) === 'percentage';
  }





  return(
    <div className={clsx("grid-pad grid", styles.ruleContainer)}>
      <div className={"col-1-1"}>
        <div className={"content"}>
          <Input
            label={`${label}.name`}
            register={register}
            displayText={"Rule name"}
            className={styles.ruleName}/>
        </div>
      </div>
      <Query
        label={`${label}.query`}
        register={register}/>
      <div className={"col-5-12"}>
        <div className={clsx("content",styles.serve)}>
          <div className={styles.serveTitle}>Serve</div>
          <Select
            title="Variation"
            content={[...getVariationList(variationsLabel, colors), ...otherOptions]}
            register={register}
            label={`${label}.variation`}
            required={true}
            />
        </div>
      </div>
      <div className={"col-1-1"}>
        <PercentagesForm
          variations={watch(`${label}.variation`)}
          variationsLabel={variationsLabel}
          watch={watch}
          register={register}
          label={`${label}.percentages`}
          control={control}
          setValue={setValue}
        />
      </div>
    </div>
  );
}

function PercentagesForm({variations, watch, variationsLabel, label, control, setValue}){
  if(variations !== 'percentage') {
    return null;
  }
  const {register} = useFormContext()
  const { fields, append, remove, replace,} = useFieldArray({
    control,
    name: label,
    rules: { minLength: 1 }
  });


  return (<div>
    <ul>
    {fields.map((field, index)=>(
      <li key={`${label}.${index}`} >
        <input type="number" {...register(`${label}.${index}.value`, false)} /> {field.name}
      </li>
      ))
    }
    </ul>
  </div>);
}
