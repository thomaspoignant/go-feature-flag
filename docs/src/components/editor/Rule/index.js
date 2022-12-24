import clsx from "clsx";
import styles from "./styles.module.css";
import inputStyles from '../Input/styles.module.css'
import {Input} from "../Input";
import {Select} from "../Select";
import React from "react";
import Link from "@docusaurus/Link";
import { useFormContext, Controller } from "react-hook-form";
import {isArray} from "redoc";
import { Progress } from 'react-sweet-progress';
import "react-sweet-progress/lib/style.css";
import {Colors} from "../Colors";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import _ from 'lodash';

export function Rule({ variations, label, isDefaultRule}){
  const { register, watch } = useFormContext();
  const otherOptions = [
    { value: "percentage", "displayName": "️↗️ a percentage rollout"},
    { value: "progressive", "displayName": "↗️ a progressive rollout"},
  ];
  function getVariationList(variations){
    const availableVariations = variations
      .map((item, index) => {
        return {"value": item.name, "displayName": `${Colors[index%Colors.length]} ${item.name}`}
      }).filter(item => item.value !== undefined && item.value !== '') || [];
    return availableVariations;
  }

  function getSelectorList(variations){
    const filteredVariations = getVariationList(variations);
    if (filteredVariations.length >=2){
      return [...filteredVariations, ...otherOptions];
    }
    return filteredVariations;
  }

  return(
    <div className={clsx("grid-pad grid", styles.ruleContainer)}>
      {!isDefaultRule && <div className={"col-1-1"}>
        <div className={"content"}>
          <Input
            label={`${label}.name`}
            displayText={"Rule name"}
            className={clsx(inputStyles.editorInputContainer, styles.ruleName)}
            required={true}
          />
        </div>
      </div> }
      {!isDefaultRule &&
      <div className={clsx("grid")}>
        <div className={"col-9-12"}>
          <div className={clsx("content", styles.inputQuery)}>
            <div className={styles.ifContainer}>
              <div className={clsx(styles.circle)}>IF</div>
            </div>
            <Input
              label={`${label}.query`}
              displayText={"Query"}
              required={true}
            />
            <Link to={'/docs/configure_flag/rule_format'} target={"_blank"}>
              <i className="fa-regular fa-circle-question"></i>
            </Link>
          </div>
        </div>
      </div>
      }
      <div className={"col-5-12"}>
        <div className={clsx("content",styles.serve)}>
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

function PercentagesForm({variations, label, selectedVar}){
  const {register, watch} = useFormContext()
  if(selectedVar !== 'percentage') {
    return null;
  }

  function computePercentages(percentages){
    const sum = percentages.filter(item => item && !isNaN(item.value)).reduce(
      (accumulator, currentValue) => accumulator + currentValue.value, 0);
    console.log(sum)
    return sum;
  }

  function ProgressBar({percentages}){
    if (!percentages || !isArray(percentages) || percentages.length <= 0) {
      return null;
    }
    const sum = computePercentages(percentages);
    if (sum > 100) {
      return (<div className={styles.error}>The total percentage cannot be more than 100%</div>);
    }

    return(<Progress percent={sum} />);
  }

  return (
    <div className={"grid-pad grid"}>
      <div className={clsx("col-1-1", styles.rolloutDesc)}>
        A percentage rollout means that your users are divided in different buckets and you serve different variations
        to them. Note that a user will always have the same variation.
      </div>
      <div className={"col-1-2"}>
        <ul className={styles.percentageContainer}>
        {variations.map((field, index)=>(
          <li key={`${label}.${index}`} >
            <Input label={`${label}.${index}.value`}
                   required={true}
                   defaultValue={0}
                   type="number"
                   displayText={`%  ${Colors[index % Colors.length]} ${field.name}`}
                   className={styles.percentageInput}
                   disablePlaceholder={true}
                   disableInlineErr={true}
                   validation={{
                     valueAsNumber: true,
                     required: { value: true, message: "Percentage field is required"},
                     min: {value:0, message:"Percentage should be between 0 and 100"},
                     max:{value:100, message:"Percentage should be between 0 and 100"}
                   }}/>
            <input type="hidden" {...register(`${label}.${index}.name`)} value={field.name} />
          </li>
          ))
        }
        </ul>
        <ProgressBar percentages={watch(label)}/>
      </div>
    </div>
  );
}

function ProgressiveStep({name, initialValue, label, variations, defaultDate}){
  const {register, control, formState: {errors}, setError, resetField } = useFormContext();

  function DisplayErrors(){
    const stepErrors = _.get(errors, label);
    if (_.isNil(stepErrors)) {
      return null;
    }

    return(<ul className={styles.formError}>
      {Object.keys(stepErrors)
        .map(key => ( <li>{stepErrors[key].message}</li>))}
    </ul>);
  }

  return(
    <div>
      <div className={clsx("grid", styles.progressiveRollout)}>
        <div>{name}</div>
        <div>
          <Controller
            control={control}
            name={`${label}.date`}
            defaultValue={defaultDate}
            rules={{required: { value: true, message: "Date field is required"},}}
            render={({ field }) => (
              <DatePicker
                className={styles.dateInput}
                placeholderText='Select date'
                showTimeSelect
                onChange={(date) => field.onChange(date)}
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
            content={variations
              .map((item, index) => {
                return {"value": item.name, "displayName": `${Colors[index%Colors.length]} ${item.name}`}
              }).filter(item => item.value !== undefined && item.value !== '') || []}
            register={register}
            label={`${label}.selectedVar`}
            required={true}
          />
        </div>
        <div className={styles.progressiveRolloutPercentage}>
          to&nbsp;<Input label={`${label}.percentage`}
                         required={true}
                         defaultValue={initialValue}
                         type="number"
                         displayText="%"
                         className={styles.percentageInput}
                         disablePlaceholder={true}
                         disableInlineErr={true}
                         validation={{
                           valueAsNumber: true,
                           min: {value:0, message:"Percentage should be between 0 and 100"},
                           max:{value:100, message:"Percentage should be between 0 and 100"}
                         }}/>
        </div>
        <div>
          <Link to={'/docs/configure_flag/rollout/progressive'} target={"_blank"}>
            <i className="fa-regular fa-circle-question"></i>
          </Link>
        </div>
      </div>
      <div>
        <DisplayErrors />
      </div>
    </div>
  )
}

function ProgressiveRollout({variations, label, selectedVar}){
  if(selectedVar !== 'progressive') {
    return null;
  }



  let endDateDefault = new Date();
  endDateDefault = endDateDefault.setDate(endDateDefault.getDate() + 10);
  return (
  <div className={"grid grid-pad"}>
    <div className={clsx("col-1-1", styles.rolloutDesc)}>A progressive rollout allows you to increase the percentage of your flag over time.<br/>
      You can select a release ramp where the percentage of your flag will increase progressively between the start date and the end date.</div>
    <ProgressiveStep
      name={"Start on the"}
      label={`${label}.initial`}
      variations={variations}
      initialValue={0}
      defaultDate={new Date()}
    />

    <ProgressiveStep
      name={"Stop on the"}
      label={`${label}.end`}
      variations={variations}
      initialValue={100}
      defaultDate={new Date(endDateDefault)}
    />
  </div>);
}
