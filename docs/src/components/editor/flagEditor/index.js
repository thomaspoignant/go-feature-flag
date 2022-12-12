import styles from "./styles.module.css";
import {Input} from "../Input";
import {Switch} from "../switch";
import clsx from "clsx";
import {Select} from "../Select";
import {Variations} from "../variations";
import React from "react";
import {Targeting} from "../targeting";

export function  FlagConfiguration({register, watch, control, parentName, setValue}){
  const typeSelectorContent = [
    {value: "boolean", displayName: "☑️ boolean"},
    {value: "string", displayName: "🔤 string"},
    {value: "number", displayName: "🔢 number"},
    {value: "json", displayName: "🖥 JSON"},
  ];
  const colors = ["🟢","🟠","🔴","🟣","⚪️","🔵","⚫️","🟡","🟤"];

  return(
    <div className={styles.flagContainer}>
      <div className="grid-pad grid">
        <div className="col-1-2 mobile-col-1-2" >
          <Input displayText="Flag Name"
                 label={`${parentName}.flagName`}
                 required={true}
                 register={register} />
        </div>
        <div className="col-1-2 mobile-col-1-2">
          <Switch id="disable"
                  label={`${parentName}.disable`}
                  displayText="Disable"
                  register={register} />
        </div>
      </div>
      <div className="grid-pad grid">
        <div className={clsx("col-3-12 mobile-col-1-2", )}>
          <Select title='Flag type'
                  content={typeSelectorContent}
                  register={register}
                  label={`${parentName}.type`}
                  required={true}/>
        </div>
        <div className={clsx("col-1-12", "mobile-col-1-1")}>
        </div>
        <div className="col-2-12 mobile-col-1-2">
          <Input id="version"
                 label={`${parentName}.version`}
                 displayText="Version"
                 register={register}/>
        </div>
      </div>
      <Variations type={watch(`${parentName}.type`)}
                  label={`${parentName}.variations`}
                  register={register}
                  control={control}
                  colors={colors}/>
      <Targeting label={`${parentName}.targeting`}
                 register={register}
                 control={control}
                 variationsLabel={`${parentName}.variations`}
                 watch={watch}
                 colors={colors}
                 setValue={setValue}/>
    </div>
  );
}
