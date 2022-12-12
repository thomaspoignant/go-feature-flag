import React from "react";
import {useFieldArray} from "react-hook-form";
import {Rule} from "../rule";

export function Targeting({register, watch, label, control, variationsLabel, colors, setValue}){
  const { fields } = useFieldArray({
    control,
    name: label
  });

  return(
    <div>
      <h2>Target specific users</h2>
      {fields.map((field, index) => (
        <Rule variationsLabel={variationsLabel}
              register={register}
              watch={watch}
              label={`${label}.${index}`}
              colors={colors}
              key={field.id}
              isDefaultRule={false}
              control={control}
              setValue={setValue}
        />
      ))}
    </div>
  );
}






