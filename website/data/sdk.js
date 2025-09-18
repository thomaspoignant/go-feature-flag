const sdkFeatureAvailableList = [
  {
    key: 'localCache',
    sdkType: 'client',
    name: 'Flags Local Cache',
    description:
      'The provider is able to cache the flags, it allow to evaluate the feature flags without waiting for the remote evaluation to be done.',
  },
  // {
  //   key: "inprocess",
  //   sdkType: 'server',
  //   name: 'In process Evaluation',
  //   description: 'The provider is able to evaluate the feature flags in process, it means that the provider does not do any remote call to evaluate the feature flags.',
  // },
  {
    key: 'remote',
    name: 'Remote Evaluation',
    description:
      'The provider is calling the remote server to evaluate the feature flags.',
  },
  {
    key: 'trackingFlag',
    name: 'Tracking Flag Evaluation',
    description:
      'The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.',
  },
  // {
  //   key: "trackingEvents",
  //   name: 'Tracking Custom Events',
  //   description: 'The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.',
  // },
  {
    key: 'configurationChange',
    name: 'Configuration Change Updates',
    description:
      'The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration.',
  },
  {
    key: 'providerEvents',
    name: 'Provider Events Reactions',
    description:
      'You can add an event handler to the provider to react to the provider events.',
  },
];

const features = (keys, sdkType) => {
  return sdkFeatureAvailableList
    .filter(it => it.sdkType === sdkType || it.sdkType === undefined)
    .map(it => ({
      ...it,
      status: keys.includes(it.key) ? 'done' : 'not implemented',
    }));
};

export const sdk = [
  {
    key: 'go',
    name: 'Golang',
    paradigm: ['Server'],
    faLogo: 'devicon-go-original-wordmark colored',
    badgeUrl:
      'https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fproxy.golang.org%2Fgithub.com%2Fopen-feature%2Fgo-sdk-contrib%2Fproviders%2Fgo-feature-flag%2F%40latest&query=%24.Version&label=GO&color=blue&style=flat-square&log=go',
    docLink: 'server_providers/openfeature_go',
    featureList: features(
      [
        'inprocess',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
      ],
      'server'
    ),
  },
  {
    key: 'java',
    name: 'Java',
    paradigm: ['Server'],
    faLogo: 'devicon-java-plain colored',
    badgeUrl:
      'https://img.shields.io/maven-central/v/dev.openfeature.contrib.providers/go-feature-flag?color=blue&style=flat-square&logo=java',
    docLink: 'server_providers/openfeature_java',
    featureList: features(
      [
        'inprocess',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
        'trackingEvents',
      ],
      'server'
    ),
  },
  {
    key: 'kotlin',
    name: 'Kotlin',
    paradigm: ['Server'],
    faLogo: 'devicon-kotlin-plain colored',
    badgeUrl:
      'https://img.shields.io/maven-central/v/dev.openfeature.contrib.providers/go-feature-flag?color=blue&style=flat-square&logo=java',
    docLink: 'server_providers/openfeature_java',
    featureList: features(
      [
        'inprocess',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
        'trackingEvents',
      ],
      'server'
    ),
  },
  {
    key: 'dotnet',
    name: '.NET',
    paradigm: ['Server'],
    faLogo: 'devicon-dot-net-plain-wordmark colored',
    badgeUrl:
      'https://img.shields.io/nuget/v/OpenFeature.Providers.GOFeatureFlag?color=blue&style=flat-square&logo=nuget',
    docLink: 'server_providers/openfeature_dotnet',
    featureList: features(
      [
        'inprocess',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
        'trackingEvents',
      ],
      'server'
    ),
  },
  {
    key: 'python',
    name: 'Python',
    paradigm: ['Server'],
    faLogo: 'devicon-python-plain colored',
    badgeUrl:
      'https://img.shields.io/pypi/v/gofeatureflag-python-provider?color=blue&style=flat-square&logo=pypi',
    docLink: 'server_providers/openfeature_python',
    featureList: features(
      ['remote', 'trackingFlag', 'configurationChange'],
      'server'
    ),
  },
  {
    key: 'javascript',
    name: 'Javascript',
    paradigm: ['Client'],
    faLogo: 'devicon-javascript-plain colored',
    badgeUrl:
      'https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue&style=flat-square&logo=npm',
    docLink: 'client_providers/openfeature_javascript',
    featureList: features(
      [
        'localCache',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
        'trackingEvents',
      ],
      'client'
    ),
  },
  {
    key: 'typescript',
    name: 'Typescript',
    paradigm: ['Client'],
    faLogo: 'devicon-typescript-plain colored',
    badgeUrl:
      'https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue&style=flat-square&logo=npm',
    docLink: 'client_providers/openfeature_javascript',
    featureList: features(
      [
        'localCache',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
        'trackingEvents',
      ],
      'client'
    ),
  },
  {
    key: 'react',
    name: 'React',
    paradigm: ['Client'],
    faLogo: 'devicon-react-original colored',
    badgeUrl:
      'https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue&style=flat-square&logo=npm',
    docLink: 'client_providers/openfeature_react',
    featureList: features(
      [
        'localCache',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
        'trackingEvents',
      ],
      'client'
    ),
  },
  {
    key: 'angular',
    name: 'Angular',
    paradigm: ['Client'],
    faLogo: 'devicon-angularjs-plain colored',
    badgeUrl:
      'https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-web-provider?color=blue&style=flat-square&logo=npm',
    docLink: 'client_providers/openfeature_angular',
    featureList: features(
      [
        'localCache',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
        'trackingEvents',
      ],
      'client'
    ),
  },
  {
    key: 'swift',
    name: 'Swift',
    paradigm: ['Client'],
    faLogo: 'devicon-swift-plain colored',
    docLink: 'client_providers/openfeature_swift',
    badgeUrl:
      'https://img.shields.io/github/v/release/go-feature-flag/openfeature-swift-provider?label=Swift&amp;display_name=tag&style=flat-square&logo=Swift',
    featureList: features(
      [
        'localCache',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
      ],
      'client'
    ),
  },
  {
    key: 'android',
    name: 'Kotlin / Android',
    paradigm: ['Client'],
    faLogo: 'devicon-android-plain colored',
    docLink: 'client_providers/openfeature_android',
    badgeUrl:
      'https://img.shields.io/maven-central/v/org.gofeatureflag.openfeature/gofeatureflag-kotlin-provider?color=blue&style=flat-square&logo=android',
    featureList: features(
      [
        'localCache',
        'remote',
        'trackingFlag',
        'configurationChange',
        'providerEvents',
      ],
      'client'
    ),
  },
  {
    key: 'nodejs',
    name: 'Node.JS',
    paradigm: ['Server'],
    faLogo: 'devicon-nodejs-plain colored',
    docLink: 'server_providers/openfeature_javascript',
    badgeUrl:
      'https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-provider?color=blue&style=flat-square&logo=npm',
    featureList: features(
      ['remote', 'trackingFlag', 'configurationChange', 'providerEvents'],
      'server'
    ),
  },
  {
    key: 'php',
    name: 'PHP',
    paradigm: ['Server'],
    faLogo: 'devicon-php-plain colored',
    badgeUrl:
      'https://img.shields.io/packagist/v/open-feature/go-feature-flag-provider?logo=php&color=blue&style=flat-square',
    docLink: 'server_providers/openfeature_php',
    featureList: features(['remote', 'trackingFlag'], 'server'),
  },
  {
    key: 'ruby',
    name: 'Ruby',
    paradigm: ['Server'],
    faLogo: 'devicon-ruby-plain colored',
    badgeUrl:
      'https://img.shields.io/gem/v/openfeature-go-feature-flag-provider?color=blue&style=flat-square&logo=ruby',
    docLink: 'server_providers/openfeature_ruby',
    featureList: features(['remote', 'trackingFlag'], 'server'),
  },
  {
    key: 'nestjs',
    name: 'NestJS',
    paradigm: ['Server'],
    faLogo: 'devicon-nestjs-plain colored',
    badgeUrl:
      'https://img.shields.io/npm/v/%40openfeature%2Fgo-feature-flag-provider?color=blue&style=flat-square&logo=npm',
    docLink: 'server_providers/openfeature_nestjs',
    featureList: features(
      ['remote', 'trackingFlag', 'configurationChange', 'providerEvents'],
      'server'
    ),
  },
];
