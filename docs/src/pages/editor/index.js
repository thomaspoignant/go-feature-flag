import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import React from "react";
import {FlagConfiguration} from "../../components/editor/flagEditor";
import {useFieldArray, useForm, useWatch} from "react-hook-form";

const EDITOR_NAME = 'GOFeatureFlagEditor';
function Editor(){
  const { register, handleSubmit, watch, formState: { errors }, control, setValue} = useForm({
    defaultValues: {
      GOFeatureFlagEditor: [{
        flagName:"test flag",
        variations: [{name:"v1", value:true},{name:"v2", value:false}],
        targeting: [ {name: 'Rule 1'}]
      },
      ]
    }
  });

  const { fields, append, remove} = useFieldArray({
    control,
    name: EDITOR_NAME,
    rules: { minLength: 1 }
  });

  const onSubmit = data => console.log(data);

  return (
    <div className="grid-pad grid">
      <div className="col-8-12">
        <form onSubmit={handleSubmit(onSubmit)}>
          {fields.map((field, index) => (
            <FlagConfiguration
              register={register}
              watch={watch}
              control={control}
              parentName={`${EDITOR_NAME}.${index}`}
              key={field.id}
              setValue={setValue}/>
          ))}
          <input type="submit" />
        </form>
      </div>
      <div className="col-4-12">right</div>
    </div>
  );
}

export default function Page() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Description will go into a meta tag in <head />">
      <Editor />
    </Layout>
  );
}
