import hljs from 'highlight.js';
import json from 'highlight.js/lib/languages/json';
import { defaultAnonymousCtx, defaultAuthenticatedCtx } from './evaluationCtx';
import { OpenFeature } from '@openfeature/web-sdk';
// import { GoFeatureFlagWebProvider } from '@openfeature/go-feature-flag-web-provider';


let currentEvaluationCtx;
let openfeatureClient;

/**
 * init is the function that initialialize the page.
 * 
 * In this example you can see that we are setting up Open Feature and the connexion to GO Feature Flag
 * We will use a feature flag to display different images based on the context of the user.
 */
function init(){
    hljs.registerLanguage('json', json);
    
    // const goffProvider = GoFeatureFlagWebProvider({endpoint: 'http://localhost:1031'});
    OpenFeature.setContext(currentEvaluationCtx);
    // OpenFeature.setProvider(goffProvider);
    openfeatureClient = OpenFeature.getClient();
    setEvaluationContext(defaultAnonymousCtx);
}

/**
 * setEvaluationContext is a function called when we change the context of the user
 * in this example we change the context when move from an anonymous user to an authenticated user.
 */ 
function setEvaluationContext(ctx){
    currentEvaluationCtx = ctx;
    document.getElementById('evaluation-ctx').innerHTML = JSON.stringify(currentEvaluationCtx, '', 2);
    hljs.highlightAll();

    // We are providing to openfeature the new context for this user.
    OpenFeature.setContext(currentEvaluationCtx);


    /**
     *  In this example we are loading 2 differents feature flag
     * badge-title: it is a string flag that control the text we want to display in the badge on the top of the screen
     * beta: it is a boolean flag that control wether or not we display the beta chip associated to the context
     * 
     * You can play with the targeting rules and your flag configuration file to see how things are changing in this page.
     */

    // We are using the Openfeature API to know what we should display as a title for the badge
    const badgeTitle = openfeatureClient.getStringValue('badge-title', 'flag badge-title not loaded');
    document.getElementById('badge-user').innerHTML = badgeTitle;

    // We are using the Openfeature API to know if the user is enroll in the beta to display the chip
    if (openfeatureClient.getBooleanValue('beta', false)){
        const betaChip = '<span class="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">beta</span>';
        document.getElementById('badge-user').innerHTML += betaChip;
    }
}

init();


// This eventListener is changing the context when we click on the button.
document.getElementById('login-btn').addEventListener('click', (event) => {
    const button = event.target;
    const isLogin = button.innerHTML === 'Login';
    if (isLogin) {
        button.classList.remove('btn-success');
        button.classList.add('btn-danger');
        button.innerHTML = 'Logout';
        document.getElementById('main-div').style.backgroundColor = '#e8f4ea';
        setEvaluationContext(defaultAuthenticatedCtx);
    } else {
        button.classList.remove('btn-danger');
        button.classList.add('btn-success');
        button.innerHTML = 'Login';
        document.getElementById('main-div').style.backgroundColor = '#f3d5d5';
        setEvaluationContext(defaultAnonymousCtx);
    }
});