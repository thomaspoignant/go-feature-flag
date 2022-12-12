import styles from "./styles.module.css";
import React from "react";

export function Select({title, content, register, label, required, onChange}){
  return(
    <div className={styles.selector}>
      <select name="typeSelector" defaultValue="0" {...register(label, required)}>
        <option disabled={true} defaultChecked={true} value={null}>{title}</option>
        {content.map((item, index) => (
          <option value={item.value} key={`variation_type_${index}`}>{item.displayName}</option>
        ))}
      </select>
    </div>
  );
}
