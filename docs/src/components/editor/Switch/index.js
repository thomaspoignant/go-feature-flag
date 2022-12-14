import clsx from "clsx";
import styles from "./styles.module.css"
import React from "react";
import {useFormContext} from "react-hook-form";

export function Switch({label, displayText}){
  const { register } = useFormContext();
  return(
    <div className="grid grid-pad">
      <div className={clsx("col-3-12", styles.displayText)}>{displayText}</div>
      <div className="col-2-12">
        <div className={clsx(styles.switch, styles.r)}>
          <input type="checkbox" className={styles.checkbox} {...register(label)}/>
          <div className={styles.knobs}></div>
          <div className={styles.layer}></div>
        </div>
      </div>
    </div>
  );
}
