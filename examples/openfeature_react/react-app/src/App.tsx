import "./App.css";
import { EvaluationContext, OpenFeature, OpenFeatureProvider, useFlag } from "@openfeature/react-sdk";
import { GoFeatureFlagWebProvider } from "@openfeature/go-feature-flag-web-provider";

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
      <Page />
    </OpenFeatureProvider>
  );
}

function Page() {
  const { value: hideLogo } = useFlag("hide-logo", false);
  const { value: titleFlag } = useFlag("title-flag", "GO Feature Flag");
  const { value: badgeClass } = useFlag("badge-class", "");

  const setRandomEvaluationContext = () => {
    const availableContexts: Array<EvaluationContext> = [
      { targetingKey: "user-2", userType: "dev", email: "contact@gofeatureflag.org" },
      { targetingKey: "user-3", userType: "admin", company: "GO Feature Flag" },
      { targetingKey: "user-4", userType: "customer", location: "Paris" },
      { targetingKey: "user-5" }

    ];
    const randomIndex = Math.floor(Math.random() * availableContexts.length);
    OpenFeature.setContext(availableContexts[randomIndex]);
  };

  const userType =OpenFeature.getContext().userType ?? "anonymous" as string;

  return (
    <>
      <div>
        <a href="https://gofeatureflag.org" target="_blank">
          {!hideLogo && <img src="/public/logo.png" className="logo" alt="GO Feature Flag logo" />}
        </a>
      </div>
      <h2>React example app</h2>
      <h1>{titleFlag}</h1>

      {badgeClass && <span className={badgeClass}>{userType.toString()}</span>}
      <pre>
        <code className="language-json">{JSON.stringify(OpenFeature.getContext(), null, 2)}</code>
      </pre>
      <button className=""
              onClick={setRandomEvaluationContext}>
        Change evaluation context
      </button>
    </>);
}

export default App;
