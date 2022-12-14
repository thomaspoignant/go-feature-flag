import clsx from "clsx";
import React from "react";
import styles from "./styles.module.css"
import {useFormContext} from "react-hook-form";

export function Input({ label, displayText, className, required}) {
  const { register } = useFormContext();
  return(
    <div className={clsx(styles.editorInputContainer, className)}>
      <input id={`${label}.input`} className={styles.editorInput} type="text"  placeholder=" " {...register(label, { required })}/>
      <div className={styles.editorCut}></div>
      <label htmlFor={`${label}.input`} className={clsx(styles.editorPlaceholder, className)}>{displayText}</label>
    </div>
  );
}
