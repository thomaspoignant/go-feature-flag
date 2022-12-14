import React from "react";
import {useForm, FormProvider, handleSummit, useFieldArray} from "react-hook-form";
import Layout from "@theme/Layout";
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {FlagForm} from "../../components/editor/FlagForm";

function App() {
  const EDITOR_NAME = 'GOFeatureFlagEditor';
  const methods = useForm({
    defaultValues: {
      GOFeatureFlagEditor: [{
        flagName:"test flag",
        variations: [{name:"v1", value:true},{name:"v2", value:false}],
        targetings: [ {name: 'Rule 1'}]
      },
      ]
    }
  });

  const { fields, append, remove} = useFieldArray({
    control: methods.control,
    name: EDITOR_NAME,
    rules: { minLength: 1 }
  });

  const onSubmit = data => console.log(data);
  return (
    <div className="grid-pad grid">
      <div className="col-8-12">
        <FormProvider {...methods} >
          <form onSubmit={methods.handleSubmit(onSubmit)}>
            {fields.map((field, index) => (
              <FlagForm
                label={`${EDITOR_NAME}.${index}`}
                key={field.id}
                />
            ))}
            <input type="submit" />
          </form>
        </FormProvider>
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
      <App />
    </Layout>
  );
}
