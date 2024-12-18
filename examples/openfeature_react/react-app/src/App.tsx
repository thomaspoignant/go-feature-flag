import "./App.css";
import {OpenFeature, OpenFeatureProvider, useFlag} from "@openfeature/react-sdk";
import {GoFeatureFlagWebProvider} from "@openfeature/go-feature-flag-web-provider";
import {AccountSelector} from "./components/account-selector.tsx";
import {login} from "./service/login.tsx";
import {Badges} from "./components/badges.tsx";

const goFeatureFlagWebProvider = new GoFeatureFlagWebProvider({
  endpoint: "http://localhost:1031"
});

OpenFeature.setContext({
  targetingKey: "user-1",
  admin: false
});
OpenFeature.setProvider(goFeatureFlagWebProvider);

function App() {
  return (
    <OpenFeatureProvider>
      <Page/>
    </OpenFeatureProvider>
  );
}

function Page() {
  const {value: hideLogo} = useFlag("hide-logo", false);
  const {value: titleFlag} = useFlag("title-flag", "GO Feature Flag");


  const onLoginChange = (name: string) => {
    const ctx = login(name);
    OpenFeature.setContext(ctx);
  }

  return (
    <>
      <div className="flex justify-center items-center pt-10">
        <a href="https://gofeatureflag.org" target="_blank">
          {!hideLogo && <img src="/public/logo.png" className="max-w-72" alt="GO Feature Flag logo"/>}
        </a>
      </div>
      <div className="flex justify-center items-center">
        <h2 className={"text-xl text-gray-200 pb-5"}>Openfeature React example app</h2>
      </div>
      <AccountSelector onChange={onLoginChange}/>
      <div className="flex justify-center items-center mt-12 gap-4">
        <span className={"text-sm text-gray-200"}>Title comming from the feature flag:</span>
        <h1 className={"text-2xl text-gray-200"}>{titleFlag}</h1>
      </div>
      <Badges/>

      <h1 className={"flex text-md justify-center items-center place-items-center mt-28 text-gray-200"}>
        Evaluation context used:</h1>
      <div className={"flex justify-center items-center place-items-center "}>
        <pre className={"text-gray-50 bg-zinc-700 py-3 px-10 rounded min-w-xl max-w-2xl"}>
          <code className="language-json">{JSON.stringify(OpenFeature.getContext(), null, 2)}</code>
        </pre>
      </div>
    </>
  );
}

export default App;
