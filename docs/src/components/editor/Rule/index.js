import clsx from "clsx";
import styles from "./styles.module.css";
import {Input} from "../Input";
import {Select} from "../Select";
import React from "react";
import Link from "@docusaurus/Link";
import { useFormContext } from "react-hook-form";
import {isArray} from "redoc";
import { Progress } from 'react-sweet-progress';
import "react-sweet-progress/lib/style.css";
import {Colors} from "../Colors";

export function Rule({ variations, label, isDefaultRule}){
  const { register, watch } = useFormContext();
  const otherOptions = [
    { value: "percentage", "displayName": "️↗️ a percentage rollout"},
    { value: "progressive", "displayName": "↗️ a progressive rollout"},
  ];
  function getVariationList(variations, colors){
    const availableVariations = variations
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
            content={[...getVariationList(variations, Colors), ...otherOptions]}
            register={register}
            label={`${label}.selectedVar`}
            required={true}
            />
        </div>
      </div>
      <div className={"col-1-1"}>
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

function PercentagesForm({variations, label, selectedVar, colors}){
  const {register, watch} = useFormContext()
  if(selectedVar !== 'percentage') {
    return null;
  }

  function ProgressBar({percentages}){
    if (!percentages || !isArray(percentages) || percentages.length <= 0) {
      return null;
    }

    const sum = percentages.filter(item => item && !isNaN(item.value)).reduce(
      (accumulator, currentValue) => accumulator + currentValue.value, 0);

    if (sum > 100) {
      return (<div className={styles.error}>The total percentage cannot be more than 100%</div>);
    }

    return(<Progress percent={sum} />);
  }

  return (<div className={"col-1-2"}>
    <ul className={styles.percentageContainer}>
    {variations.map((field, index)=>(
      <li key={`${label}.${index}`} >
        <input className={styles.percentageInput}
               type="number" {...register(`${label}.${index}.value`,{required: true, valueAsNumber:true, min: 0, max: 100})}
               defaultValue={0} /> {Colors[index % Colors.length]} {field.name}
        <input type="hidden" {...register(`${label}.${index}.name`)} value={field.name} />
      </li>
      ))
    }
    </ul>
    <ProgressBar percentages={watch(label)}/>

  </div>);
}

function ProgressiveRollout({variations, label, selectedVar, colors}){
  const {register} = useFormContext()
  if(selectedVar !== 'progressive') {
    return null;
  }

  console.log(variations);

  // initial:
  //   variation: variationB
  // percentage: 0
  // date: 2021-03-20T00:00:00.1-05:00
  // end:
  //   variation: variationB
  // percentage: 100
  // date: 2021-03-21T00:00:00.1-05:00

  function ProgressiveStep({name, initialValue, label, variations}){
    return(
      <div  className={"grid"}>
        <div className={"col-2-12"}>{name}</div>
        <div className={"col-3-12"}>
          <Select
            title="Variation"
            content={variations
              .map((item, index) => {
                return {"value": item.name, "displayName": `${Colors[index%Colors.length]} ${item.name}`}
              }).filter(item => item.value !== undefined && item.value !== '') || []}
            register={register}
            label={`${label}.selectedVar`}
            required={true}
          />
        </div>
      </div>
    )
  }


  return (<div>
    <ProgressiveStep name={"initial"} label={`${label}.initial`} variations={variations} register={register}/>
    <ProgressiveStep name={"end"} label={`${label}.end`} variations={variations} register={register}/>
  </div>);
}
