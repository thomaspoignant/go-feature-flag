import { OpenFeature } from '@openfeature/web-sdk';
import { GoFeatureFlagWebProvider } from '@openfeature/go-feature-flag-web-provider';

export async function Test(){
    const evaluationCtx = {
        targetingKey: 'user-key',
        email: 'john.doe@gofeatureflag.org',
        name: 'John Doe',
    };

    const goFeatureFlagWebProvider = new GoFeatureFlagWebProvider({
        endpoint: 'http://localhost:1031',
        // ...
    }, console.log);
    

    await OpenFeature.setContext(evaluationCtx);
    OpenFeature.setProvider(goFeatureFlagWebProvider);
    const client = await OpenFeature.getClient();
    if(client.getBooleanValue('my-new-feature', false)){
        console.log('true');
    } else {
        console.log('false');
    }
}