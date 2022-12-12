import clsx from "clsx";
import React from "react";
import styles from "./styles.module.css"

export function Input({ label, id, register, required, displayText, className}) {
  return(
    <div className={clsx(styles.editorInputContainer, className)}>
      <input id={id} className={styles.editorInput} type="text"  placeholder=" " {...register(label, { required })}/>
      <div className={styles.editorCut}></div>
      <label htmlFor={id} className={clsx(styles.editorPlaceholder, className)}>{displayText}</label>
    </div>
  );
}
