import azbloblogo from '@site/static/docs/collectors/azblob.png';
import bitbucketlogo from '@site/static/docs/retrievers/bitbucket.png';
import s3logo from '@site/static/docs/collectors/s3.png';
import webhooklogo from '@site/static/docs/collectors/webhook.png';
import kinesislogo from '@site/static/docs/collectors/kinesis.png';
import pubsublogo from '@site/static/docs/collectors/pubsub.png';
import sqslogo from '@site/static/docs/collectors/sqs.png';
import teamslogo from '@site/static/docs/notifier/teams.png';

export const integrations = {
  retrievers: [
    {
      name: 'Azure Blob Storage',
      description: 'Retrieves the configuration from an Azure Blob Storage.',
      bgColor: 'cornflowerblue',
      logo: azbloblogo,
    },
    {
      name: 'Bitbucket',
      description:
        'Fetch the configuration from files stored in a GIT repository.',
      bgColor: '#0052cc',
      logo: bitbucketlogo,
    },
    {
      name: 'File System',
      description: 'Fetch the configuration from a local file.',
      bgColor: '#000000',
      faLogo: 'fas fa-file fa-stack-1x fa-inverse',
    },
    {
      name: 'Google Cloud Storage',
      description:
        'Retrieves the configuration from a Google Cloud Storage bucket.',
      bgColor: 'cornflowerblue',
      faLogo: 'devicon-googlecloud-plain',
    },
    {
      name: 'GitHub',
      description:
        'Fetch the configuration from files stored in a GIT repository.',
      bgColor: '#000000',
      faLogo: 'fab fa-github fa-stack-1x fa-inverse',
    },
    {
      name: 'GitLab',
      description:
        'Fetch the configuration from files stored in a GIT repository.',
      bgColor: '#D1D0D3',
      faLogo: 'devicon-gitlab-plain colored',
    },
    {
      name: 'HTTP',
      description: 'Fetches the configuration from a remote URL over HTTP(S).',
      bgColor: '#34a853',
      logo: webhooklogo,
    },
    {
      name: 'Kubernetes ConfigMap',
      description: 'Loads the configuration from a Kubernetes ConfigMap.',
      bgColor: 'cornflowerblue',
      faLogo: 'devicon-kubernetes-plain',
    },
    {
      name: 'MongoDB',
      description: 'Load the configuration from a MongoDB collection.',
      bgColor: '#023430',
      faLogo: 'devicon-mongodb-plain-wordmark colored',
    },
    {
      name: 'Redis',
      description: 'Load the configuration from Redis using a specific prefix.',
      bgColor: '#000000',
      faLogo: 'devicon-redis-plain-wordmark colored',
    },
    {
      name: 'AWS S3',
      description: 'Retrieves the configuration from an AWS S3 bucket.',
      bgColor: '#222e3c',
      logo: s3logo,
    },
  ],
  exporters: [
    {
      name: 'Azure Blob Storage',
      description: 'Export evaluation data to an Azure Blob Storage.',
      type: 'async',
      bgColor: 'cornflowerblue',
      logo: azbloblogo,
    },
    {
      name: 'File System',
      description: 'Export evaluation data to a directory in your file system.',
      type: 'async',
      bgColor: '#000000',
      faLogo: 'fas fa-file fa-stack-1x fa-inverse',
    },
    {
      name: 'Google Cloud Storage',
      description: 'Export evaluation data to a Google Cloud Storage Bucket.',
      type: 'async',
      bgColor: 'cornflowerblue',
      faLogo: 'devicon-googlecloud-plain',
    },
    {
      name: 'Apache Kafka',
      description: 'Export evaluation data inside a Kafka topic.',
      type: 'sync',
      bgColor: '#eee',
      faLogo: 'devicon-apachekafka-original colored',
    },
    {
      name: 'AWS Kinesis',
      description: 'Export evaluation data inside a Kafka Kinesis stream.',
      type: 'sync',
      bgColor: '#222e3c',
      logo: kinesislogo,
    },
    {
      name: 'Log',
      description: 'Export evaluation data inside the application logger.',
      type: 'sync',
      bgColor: '#000000',
      faLogo: 'fas fa-file-lines fa-stack-1x fa-inverse',
    },
    {
      name: 'Google Cloud PubSub',
      description: 'Export evaluation data inside a GCP PubSub topic.',
      type: 'sync',
      bgColor: 'rgb(194, 223, 255)',
      logo: pubsublogo,
    },
    {
      name: 'AWS S3',
      description: 'Export evaluation data to a AWS S3 Bucket.',
      type: 'async',
      bgColor: '#222e3c',
      logo: s3logo,
    },
    {
      name: 'AWS SQS',
      description: 'Export evaluation data inside a AWS SQS queue.',
      type: 'sync',
      bgColor: '#222e3c',
      logo: sqslogo,
    },
    {
      name: 'Webhook',
      description: 'Export evaluation data by calling a HTTP Webhook.',
      type: 'sync',
      bgColor: '#34a853',
      logo: webhooklogo,
    },
  ],
  notifiers: [
    {
      name: 'Discord',
      description: 'Send notifications to a Discord channel.',
      bgColor: '#5661ea',
      faLogo: 'fa-brands fa-discord',
    },
    {
      name: 'Log',
      description: 'Send notifications as a log in your application.',
      bgColor: '#000000',
      faLogo: 'fas fa-file-lines fa-stack-1x fa-inverse',
    },
    {
      name: 'Microsoft Teams',
      description:
        "Send notificaimport teamslogo from '@site/static/docs/notifier/teams.png';tions to a Microsoft Teams channel.",
      bgColor: '#222e3c',
      logo: teamslogo,
    },
    {
      name: 'Slack',
      description: 'Send notifications to a Slack channel.',
      bgColor: '#4a154b',
      faLogo: 'fab fa-slack fa-stack-1x fa-inverse',
    },
    {
      name: 'Webhook',
      description: 'Send notifications to a Webhook in a specific format.',
      bgColor: '#34a853',
      logo: webhooklogo,
    },
  ],
};
