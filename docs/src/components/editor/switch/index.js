import clsx from "clsx";
import styles from "./styles.module.css"
import React from "react";

export function Switch({register, label, id, displayText}){
  return(
    <div className="grid grid-pad">
      <div className={clsx("col-3-12", styles.displayText)}>{displayText}</div>
      <div className="col-2-12">
        <div className={clsx(styles.switch, styles.r)}>
          <input type="checkbox" className={styles.checkbox} id={id} {...register(label)}/>
          <div className={styles.knobs}></div>
          <div className={styles.layer}></div>
        </div>
      </div>
    </div>
  );
}
