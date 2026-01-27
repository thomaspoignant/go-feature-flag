import React from 'react';
import FeatureTable from './index';

// Example usage of the FeatureTable component
const ExampleUsage = () => {
  // Example SDK data structure
  const exampleSdk = {
    name: 'Kotlin / Android',
    featureList: [
      {
        name: 'Flag evaluation',
        status: 'done',
        description: 'It is possible to evaluate all the type of flags',
      },
      {
        name: 'Cache invalidation',
        status: 'done',
        description:
          'A polling mechanism is in place to refresh the cache in case of configuration change',
      },
      {
        name: 'Logging',
        status: 'not implemented',
        description: 'Not supported by the SDK',
      },
      {
        name: 'Flag Metadata',
        status: 'done',
        description: 'You have access to your flag metadata',
      },
      {
        name: 'Event Streaming',
        status: 'done',
        description:
          'You can register to receive some internal event from the provider',
      },
      {
        name: 'Unit test',
        status: 'done',
        description:
          'The test are running one by one, but we still have an issue open to enable fully the tests',
      },
    ],
  };

  return (
    <div>
      <h2>Feature Table Example</h2>
      <FeatureTable sdk={exampleSdk} />
    </div>
  );
};

export default ExampleUsage;
