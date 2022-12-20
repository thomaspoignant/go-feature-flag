import React from "react";
import {useFormContext} from "react-hook-form";
import Highlight from 'react-highlight'
import 'highlight.js/styles/a11y-dark.css'



function formToGoFeatureFlag(formData){
  // {
  //   "GOFeatureFlagEditor": [
  //   {
  //     "targetings": [
  //       {
  //         "name": "Rule 1",
  //         "query": "xYOxd",
  //         "selectedVar": "Variation_1"
  //       },
  //       {
  //         "name": "Rule 2",
  //         "query": "dd",
  //         "selectedVar": "progressive",
  //         "progressive": {
  //           "initial": {
  //             "date": "2022-11-28T23:00:00.000Z",
  //             "selectedVar": "Variation_1",
  //             "percentage": 0
  //           },
  //           "end": {
  //             "date": "2022-12-19T18:00:00.261Z",
  //             "selectedVar": "Variation_2",
  //             "percentage": 100
  //           }
  //         }
  //       }
  //     ],
  //     "defaultRule": {
  //       "selectedVar": "percentage",
  //       "percentages": [
  //         {
  //           "name": "Variation_1",
  //           "value": 10
  //         },
  //         {
  //           "name": "Variation_2",
  //           "value": 90
  //         }
  //       ]
  //     }
  //   }
  // ]
  // }

  function convertValueIntoType(value, type){
    switch(type) {
      case 'json':
        try {
          return JSON.parse(value.value);
        } catch (e) {
          return undefined;
        }
      case 'number':
        return Number(value) || undefined;
      case 'boolean':
        if (typeof value == "boolean") return value
        return value !== undefined && (typeof value === 'string' || value instanceof String) && value.toLowerCase() === 'true';
      default:
        return String(value) || undefined;
    }
  }

  function convertRule(ruleForm){
    let variation, percentage, progressiveRollout = undefined;
    console.log(ruleForm);
    const { selectedVar } = ruleForm;
    switch(selectedVar){
      case "percentage":
        percentage= {}
        ruleForm.percentages.forEach(i=> percentage[i.name]=i.value);
        break;
      case "progressive":
        progressiveRollout = {
          initial: {
            variation: ruleForm.progressive.initial.selectedVar,
            percentage: ruleForm.progressive.initial.percentage,
            date: ruleForm.progressive.initial.date,
          },
          end: {
            variation: ruleForm.progressive.end.selectedVar,
            percentage: ruleForm.progressive.end.percentage,
            date: ruleForm.progressive.end.date,
          },
        };
        break;
      default:
        variation=selectedVar;
        break;
    }



    return {
      name: ruleForm.name || undefined,
      query: ruleForm.query,
      variation,
      percentage,
      progressiveRollout,
    }
  }
  function singleFlagFormConvertor(flagFormData) {
    const variationType = flagFormData.type;
    const variations = {};

    flagFormData.variations
      .filter(i => i.name !== undefined && i.name !== '' && i.value !== undefined && i.value !== '')
      .forEach(i =>  variations[i.name] = convertValueIntoType(i.value, variationType));


    const targeting = flagFormData.targeting.map(t => convertRule(t));
    const trackEvents = convertValueIntoType(flagFormData.trackEvents, "boolean");
    const disable = convertValueIntoType(flagFormData.disable, "boolean");
    const defaultRule= convertRule(flagFormData.defaultRule);

    return {
      variations,
      disable: !disable? undefined : disable,
      trackEvents: trackEvents? undefined: trackEvents,
      version: flagFormData.version === ''? undefined : flagFormData.version,
      targeting: targeting.length > 0 ? targeting : undefined,
      defaultRule,
    }
  }

  const goffFlags = {};
  formData.GOFeatureFlagEditor
    .filter(i => i.flagName !== undefined && i.flagName !== '')
    .forEach(i=> goffFlags[i.flagName]=singleFlagFormConvertor(i));

  return goffFlags;
}

export function FlagDisplay(){
  const { watch } = useFormContext();
  const data = watch();
  return(
    <div>
      <Highlight className='JSON'>
        {JSON.stringify(formToGoFeatureFlag(data), null, 2)}
      </Highlight>
    </div>
  );
}
