import styles from "./styles.module.css";
import {Input} from "../Input";
import React from "react";
import {Switch} from "../Switch";
import clsx from "clsx";
import {Select} from "../Select";
import {Variations} from "../variations";

export function FlagForm({label}){
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
                 label={`${label}.flagName`}
                 required={true}
                 />
        </div>
        <div className="col-1-2 mobile-col-1-2">
          <Switch id="disable"
                  label={`${label}.disable`}
                  displayText="Disable" />
        </div>
      </div>
      <div className="grid-pad grid">
        <div className={clsx("col-3-12 mobile-col-1-2", )}>
          <Select title='Flag type'
                  content={typeSelectorContent}
                  label={`${label}.type`}
                  required={true}/>
        </div>
        <div className={clsx("col-1-12", "mobile-col-1-1")}>
        </div>
        <div className="col-2-12 mobile-col-1-2">
          <Input id="version"
                 label={`${label}.version`}
                 displayText="Version"/>
        </div>
      </div>
      <Variations label={`${label}.variations`}
                  colors={colors}/>
    </div>
  );
}
