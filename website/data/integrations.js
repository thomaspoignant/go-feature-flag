import azbloblogo from '@site/static/docs/collectors/azblob.png';
import httplogo from '@site/static/docs/retrievers/http.png';
import bitbucketlogo from '@site/static/docs/retrievers/bitbucket.png';
import s3logo from '@site/static/docs/collectors/s3.png';
import webhooklogo from '@site/static/docs/collectors/webhook.png';
import kinesislogo from '@site/static/docs/collectors/kinesis.png';
import pubsublogo from '@site/static/docs/collectors/pubsub.png';
import sqslogo from '@site/static/docs/collectors/sqs.png';
import teamslogo from '@site/static/docs/notifier/teams.png';
import k8slogo from '@site/static/docs/retrievers/k8s.png';
import filelogo from '@site/static/docs/retrievers/file.png';
import googlelogo from '@site/static/docs/retrievers/google.png';
import githublogo from '@site/static/docs/retrievers/github.png';
import gitlablogo from '@site/static/docs/retrievers/gitlab.png';
import mongodblogo from '@site/static/docs/retrievers/mongodb.png';
import redislogo from '@site/static/docs/retrievers/redis.png';
import postgreslogo from '@site/static/docs/retrievers/postgresql.png';
import kafkalogo from '@site/static/docs/collectors/kafka.png';
import discordlogo from '@site/static/docs/notifier/discord_logo.png';
import slacklogo from '@site/static/docs/notifier/slack.png';
import opentelemetrylogo from '@site/static/docs/collectors/opentelemetry.png';

export const integrations = {
  retrievers: [
    {
      name: 'HTTP(S)',
      description: 'Fetches the configuration from a remote URL over HTTP(S).',
      longDescription:
        'Fetches the configuration from a remote URL over HTTP(S). This retriever is useful when you want to fetch the configuration from a remote server endpoint.',
      bgColor: '#34a853',
      logo: httplogo,
      docLink: 'http',
    },
    {
      name: 'File System',
      description: 'Fetch the configuration from a local file.',
      longDescription: `Fetch the configuration from a local file. This retriever is useful when you want to load the configuration from a file in your file system.`,
      bgColor: '#000000',
      faLogo: 'fas fa-file fa-stack-1x fa-inverse',
      logo: filelogo,
      docLink: 'file',
    },
    {
      name: 'Kubernetes ConfigMap',
      description: 'Loads the configuration from a Kubernetes ConfigMap.',
      longDescription: `Loads the configuration from a Kubernetes ConfigMap. This retriever is useful when you are using Kubernetes and want to use ConfigMaps to store your configuration files.`,
      bgColor: 'cornflowerblue',
      faLogo: 'devicon-kubernetes-plain',
      logo: k8slogo,
      docLink: 'kubernetes-configmap',
    },
    {
      name: 'AWS S3',
      description: 'Retrieves the configuration from an AWS S3 bucket.',
      longDescription: `Retrieves the configuration from an AWS S3 bucket. This retriever is useful when you are using AWS and want to use S3 to store you configuration files.`,
      bgColor: '#222e3c',
      logo: s3logo,
      docLink: 'aws-s3',
    },
    {
      name: 'Google Cloud Storage',
      description:
        'Retrieves the configuration from a Google Cloud Storage bucket.',
      longDescription: `Retrieves the configuration from a Google Cloud Storage bucket. This retriever is useful when you are using Google Cloud and want to use GCS to store your configuration files.`,
      bgColor: 'cornflowerblue',
      faLogo: 'devicon-googlecloud-plain',
      logo: googlelogo,
      docLink: 'google-cloud-storage',
    },
    {
      name: 'Azure Blob Storage',
      description: 'Retrieves the configuration from an Azure Blob Storage.',
      longDescription: `Retrieves the configuration from an Azure Blob Storage. This retriever is useful when you are using Azure and want to use Azure Blob Storage to store your configuration files.`,
      bgColor: 'cornflowerblue',
      logo: azbloblogo,
      docLink: 'azure-blob-storage',
    },
    {
      name: 'GitHub',
      description:
        'Fetch the configuration from files stored in a GIT repository.',
      longDescription: `Fetch the configuration from files stored in a GitHub repository. This retriever will perform an HTTP Request with your GitHub configuration on the GitHub API to get your flags.`,
      bgColor: '#000000',
      faLogo: 'fab fa-github fa-stack-1x fa-inverse',
      logo: githublogo,
      docLink: 'github',
    },
    {
      name: 'GitLab',
      description:
        'Fetch the configuration from files stored in a GIT repository.',
      longDescription: `Fetch the configuration from files stored in a Gitlab repository. This retriever will perform an HTTP Request with your Gitlab configuration on the Gitlab API to get your flags.`,
      bgColor: '#D1D0D3',
      faLogo: 'devicon-gitlab-plain colored',
      logo: gitlablogo,
      docLink: 'gitlab',
    },
    {
      name: 'Bitbucket',
      description:
        'Fetch the configuration from files stored in a GIT repository.',
      bgColor: '#0052cc',
      logo: bitbucketlogo,
      docLink: 'bitbucket',
    },
    {
      name: 'MongoDB',
      description: 'Load the configuration from a MongoDB collection.',
      longDescription: `Load the configuration from a MongoDB collection. This retriever is useful when you are using MongoDB and want to use a collection to store your configuration files.`,
      bgColor: '#023430',
      faLogo: 'devicon-mongodb-plain-wordmark colored',
      logo: mongodblogo,
      docLink: 'mongodb',
    },
    {
      name: 'Redis',
      description: 'Load the configuration from Redis using a specific prefix.',
      bgColor: '#000000',
      faLogo: 'devicon-redis-plain-wordmark colored',
      logo: redislogo,
      docLink: 'redis',
    },
    {
      name: 'Postgresql',
      description: 'Load the configuration from Postgresql database.',
      longDescription: `Load the configuration from Postgresql database. This retriever is useful when you are using Postgresql and want to use a database to store your configuration files.`,
      bgColor: '#336791',
      faLogo: 'devicon-postgresql-plain',
      logo: postgreslogo,
      docLink: 'postgresql',
      minVersion: 'v1.46.0',
    },
  ],
  exporters: [
    {
      name: 'AWS S3',
      description: 'Export evaluation data to a AWS S3 Bucket.',
      longDescription: `The S3 exporter will collect the data and create a new file in a specific folder everytime we send the data.`,
      type: 'async',
      bgColor: '#222e3c',
      logo: s3logo,
      docLink: 'aws-s3',
    },
    {
      name: 'Azure Blob Storage',
      description: 'Export evaluation data to an Azure Blob Storage.',
      type: 'async',
      bgColor: 'cornflowerblue',
      logo: azbloblogo,
      docLink: 'azure-blob-storage',
    },
    {
      name: 'Google Cloud Storage',
      description: 'Export evaluation data to a Google Cloud Storage Bucket.',
      type: 'async',
      bgColor: 'cornflowerblue',
      faLogo: 'devicon-googlecloud-plain',
      logo: googlelogo,
      docLink: 'google-cloud-storage',
    },
    {
      name: 'File System',
      description: 'Export evaluation data to a directory in your file system.',
      type: 'async',
      bgColor: '#000000',
      faLogo: 'fas fa-file fa-stack-1x fa-inverse',
      logo: filelogo,
      docLink: 'file',
    },
    {
      name: 'Apache Kafka',
      description: 'Export evaluation data inside a Kafka topic.',
      type: 'sync',
      bgColor: '#eee',
      faLogo: 'devicon-apachekafka-original colored',
      logo: kafkalogo,
      docLink: 'kafka',
    },
    {
      name: 'AWS Kinesis',
      description: 'Export evaluation data inside a Kafka Kinesis stream.',
      type: 'sync',
      bgColor: '#222e3c',
      logo: kinesislogo,
      docLink: 'aws-kinesis',
    },
    {
      name: 'Google Cloud PubSub',
      description: 'Export evaluation data inside a GCP PubSub topic.',
      type: 'sync',
      bgColor: 'rgb(194, 223, 255)',
      logo: pubsublogo,
      docLink: 'google-cloud-pubsub',
    },
    {
      name: 'AWS SQS',
      description: 'Export evaluation data inside a AWS SQS queue.',
      type: 'sync',
      bgColor: '#222e3c',
      logo: sqslogo,
      docLink: 'aws-sqs',
    },
    {
      name: 'Webhook',
      description: 'Export evaluation data by calling a HTTP Webhook.',
      type: 'sync',
      bgColor: '#34a853',
      logo: webhooklogo,
      docLink: 'webhook',
    },
    {
      name: 'Log',
      description: 'Export evaluation data inside the application logger.',
      type: 'sync',
      bgColor: '#000000',
      faLogo: 'fa-solid fa-file-lines fa-inverse',
      docLink: 'log',
    },
    {
      name: 'OpenTelemetry',
      description: 'Export evaluation events as OpenTelemetry spans.',
      type: 'sync',
      bgColor: '#4285f4',
      logo: opentelemetrylogo,
      docLink: 'opentelemetry',
    },
  ],
  notifiers: [
    {
      name: 'Slack',
      description: 'Send notifications to a Slack channel.',
      bgColor: '#4a154b',
      faLogo: 'fab fa-slack fa-stack-1x fa-inverse',
      logo: slacklogo,
      docLink: 'slack',
    },
    {
      name: 'Discord',
      description: 'Send notifications to a Discord channel.',
      bgColor: '#5661ea',
      faLogo: 'fa-brands fa-discord',
      logo: discordlogo,
      docLink: 'discord',
    },
    {
      name: 'Microsoft Teams',
      description: 'Send notifications to a Microsoft Teams channel.',
      longDescription:
        'The microsoft teams notifier allows to get notified on your favorite microsoft teams channel when an instance of GO Feature FLag is\n' +
        'detecting changes in the configuration of your flags.',
      bgColor: '#222e3c',
      logo: teamslogo,
      docLink: 'microsoft-teams',
    },
    {
      name: 'Webhook',
      description: 'Send notifications to a Webhook in a specific format.',
      bgColor: '#34a853',
      logo: webhooklogo,
      docLink: 'webhook',
    },
    {
      name: 'Log',
      description: 'Send notifications as a log in your application logger.',
      bgColor: '#000000',
      faLogo: 'fa-solid fa-file-lines fa-inverse',
    },
  ],
};

// Helper function to compare semantic versions
const compareVersions = (version1, version2) => {
  if (!version1 || !version2) return 0;

  // Remove 'v' prefix if present
  const v1 = version1.replace(/^v/, '').split('.').map(Number);
  const v2 = version2.replace(/^v/, '').split('.').map(Number);

  for (let i = 0; i < Math.max(v1.length, v2.length); i++) {
    const num1 = v1[i] || 0;
    const num2 = v2[i] || 0;

    if (num1 > num2) return 1;
    if (num1 < num2) return -1;
  }

  return 0;
};

// Filter integrations based on minVersion
const filterByVersion = (items, targetVersion) => {
  if (!targetVersion) return items;

  return items.filter(item => {
    // If no minVersion is specified, include the item
    if (!item.minVersion) return true;

    // Include item if targetVersion >= minVersion
    return compareVersions(targetVersion, item.minVersion) >= 0;
  });
};

export const getIntegrations = version => {
  return {
    retrievers: filterByVersion(integrations.retrievers, version),
    exporters: filterByVersion(integrations.exporters, version),
    notifiers: filterByVersion(integrations.notifiers, version),
  };
};
