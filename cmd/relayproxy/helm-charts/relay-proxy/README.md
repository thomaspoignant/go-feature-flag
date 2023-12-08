# relay-proxy

![Version: 1.20.0](https://img.shields.io/badge/Version-1.20.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.20.0](https://img.shields.io/badge/AppVersion-v1.20.0-informational?style=flat-square)

A Helm chart to deploy go-feature-flag-relay proxy into Kubernetes

## How to use the chart

Please replace the keys `relayproxy.config` in  the `Values.yaml` to fit
your configuration. This file will be stored as `configmap` in your cluster and
be mount as a volume for the `relay-proxy`.

After changing the working directory to `cmd/relayproxy/helm-charts/relay-proxy`,
run the below command:

```shell
helm install . --name-template=go-feature-flag-relay-proxy
```

It will install the chart in your cluster.

**Homepage:** <https://gofeatureflag.org>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| thomaspoignant | <thomas.poignant@gofeatureflag.org> | <https://gofeatureflag.org> |

## Source Code

* <https://github.com/thomaspoignant/go-feature-flag>

## Values

<table>
	<thead>
		<th>Key</th>
		<th>Type</th>
		<th>Default</th>
		<th>Description</th>
	</thead>
	<tbody>
		<tr>
			<td id="affinity">
				<a href="./values.yaml#L122">affinity</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				Affinity settings for pod assignment to nodes
			</td>
		</tr>
		<tr>
			<td id="autoscaling">
				<a href="./values.yaml#L103">autoscaling</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{
  "enabled": false,
  "maxReplicas": 100,
  "minReplicas": 1,
  "targetCPUUtilizationPercentage": 80,
  "targetMemoryUtilizationPercentage": 80
}
</pre>
</div>
			</td>
			<td>
				automatically scale the deployment up and down based on observed CPU and memory utilization
			</td>
		</tr>
		<tr>
			<td id="autoscaling--enabled">
				<a href="./values.yaml#L105">autoscaling.enabled</a>
            </td>
			<td>
bool
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
false
</pre>
</div>
			</td>
			<td>
				enable autoscaling
			</td>
		</tr>
		<tr>
			<td id="autoscaling--maxReplicas">
				<a href="./values.yaml#L109">autoscaling.maxReplicas</a>
            </td>
			<td>
int
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
100
</pre>
</div>
			</td>
			<td>
				max replicas to scale to
			</td>
		</tr>
		<tr>
			<td id="autoscaling--minReplicas">
				<a href="./values.yaml#L107">autoscaling.minReplicas</a>
            </td>
			<td>
int
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
1
</pre>
</div>
			</td>
			<td>
				min replicas to scale to
			</td>
		</tr>
		<tr>
			<td id="autoscaling--targetCPUUtilizationPercentage">
				<a href="./values.yaml#L111">autoscaling.targetCPUUtilizationPercentage</a>
            </td>
			<td>
int
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
80
</pre>
</div>
			</td>
			<td>
				target CPU utilization percentage to spin up new pods
			</td>
		</tr>
		<tr>
			<td id="autoscaling--targetMemoryUtilizationPercentage">
				<a href="./values.yaml#L113">autoscaling.targetMemoryUtilizationPercentage</a>
            </td>
			<td>
int
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
80
</pre>
</div>
			</td>
			<td>
				target memory utilization percentage to spin up new pods
			</td>
		</tr>
		<tr>
			<td id="env">
				<a href="./values.yaml#L15">env</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				Environment variables to pass to the relay proxy
			</td>
		</tr>
		<tr>
			<td id="fullnameOverride">
				<a href="./values.yaml#L42">fullnameOverride</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
""
</pre>
</div>
			</td>
			<td>
				Completely override the deployment name for kubernetes objects
			</td>
		</tr>
		<tr>
			<td id="image--pullPolicy">
				<a href="./values.yaml#L33">image.pullPolicy</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
"IfNotPresent"
</pre>
</div>
			</td>
			<td>
				The image is pulled only if it is not already present locally
			</td>
		</tr>
		<tr>
			<td id="image--repository">
				<a href="./values.yaml#L31">image.repository</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
"thomaspoignant/go-feature-flag"
</pre>
</div>
			</td>
			<td>
				The image repository to pull from
			</td>
		</tr>
		<tr>
			<td id="image--tag">
				<a href="./values.yaml#L35">image.tag</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
""
</pre>
</div>
			</td>
			<td>
				Overrides the image tag whose default is the chart appVersion
			</td>
		</tr>
		<tr>
			<td id="imagePullSecrets">
				<a href="./values.yaml#L38">imagePullSecrets</a>
            </td>
			<td>
list
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
[]
</pre>
</div>
			</td>
			<td>
				Specify imagePullSecrets to be used for the deployment
			</td>
		</tr>
		<tr>
			<td id="ingress">
				<a href="./values.yaml#L76">ingress</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{
  "annotations": {},
  "className": "",
  "enabled": false,
  "hosts": [
    {
      "host": "chart-example.local",
      "paths": [
        {
          "path": "/",
          "pathType": "ImplementationSpecific"
        }
      ]
    }
  ],
  "tls": []
}
</pre>
</div>
			</td>
			<td>
				Ingress configuration
			</td>
		</tr>
		<tr>
			<td id="ingress--annotations">
				<a href="./values.yaml#L82">ingress.annotations</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				Annotations to add to the ingress
			</td>
		</tr>
		<tr>
			<td id="ingress--className">
				<a href="./values.yaml#L80">ingress.className</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
""
</pre>
</div>
			</td>
			<td>
				Ingress class name
			</td>
		</tr>
		<tr>
			<td id="ingress--enabled">
				<a href="./values.yaml#L78">ingress.enabled</a>
            </td>
			<td>
bool
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
false
</pre>
</div>
			</td>
			<td>
				Enable ingress
			</td>
		</tr>
		<tr>
			<td id="nameOverride">
				<a href="./values.yaml#L40">nameOverride</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
""
</pre>
</div>
			</td>
			<td>
				replaces the name of the chart in the Chart.yaml file
			</td>
		</tr>
		<tr>
			<td id="nodeSelector">
				<a href="./values.yaml#L116">nodeSelector</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				Node labels for pod assignment
			</td>
		</tr>
		<tr>
			<td id="podAnnotations">
				<a href="./values.yaml#L54">podAnnotations</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				Pod annotations to add to the deployment
			</td>
		</tr>
		<tr>
			<td id="podSecurityContext">
				<a href="./values.yaml#L57">podSecurityContext</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				A security context defines privilege and access control settings for a Pod
			</td>
		</tr>
		<tr>
			<td id="relayproxy--config">
				<a href="./values.yaml#L4">relayproxy.config</a>
            </td>
			<td>
tpl/object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="tpl">
relayproxy.config: |
  listen: 1031
  pollingInterval: 1000
  startWithRetrieverError: false
  retriever:
    kind: http
    url: https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.goff.yaml
  exporter:
    kind: log
 
</pre>
</div>
			</td>
			<td>
				Define this for extra Django environment variables
			</td>
		</tr>
		<tr>
			<td id="replicaCount">
				<a href="./values.yaml#L27">replicaCount</a>
            </td>
			<td>
int
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
1
</pre>
</div>
			</td>
			<td>
				The number of replicas to create for the deployment
			</td>
		</tr>
		<tr>
			<td id="resources--requests--cpu">
				<a href="./values.yaml#L100">resources.requests.cpu</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
"500m"
</pre>
</div>
			</td>
			<td>
				The amount of cpu to request for the container
			</td>
		</tr>
		<tr>
			<td id="resources--requests--memory">
				<a href="./values.yaml#L98">resources.requests.memory</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
"128Mi"
</pre>
</div>
			</td>
			<td>
				The amount of memory to request for the container
			</td>
		</tr>
		<tr>
			<td id="securityContext">
				<a href="./values.yaml#L61">securityContext</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				A security context defines privilege and access control settings for a Container
			</td>
		</tr>
		<tr>
			<td id="service--port">
				<a href="./values.yaml#L73">service.port</a>
            </td>
			<td>
int
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
1031
</pre>
</div>
			</td>
			<td>
				The port to expose on the service
			</td>
		</tr>
		<tr>
			<td id="service--type">
				<a href="./values.yaml#L71">service.type</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
"ClusterIP"
</pre>
</div>
			</td>
			<td>
				The type of service to create
			</td>
		</tr>
		<tr>
			<td id="serviceAccount--annotations">
				<a href="./values.yaml#L48">serviceAccount.annotations</a>
            </td>
			<td>
object
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
{}
</pre>
</div>
			</td>
			<td>
				Annotations to add to the service account
			</td>
		</tr>
		<tr>
			<td id="serviceAccount--create">
				<a href="./values.yaml#L46">serviceAccount.create</a>
            </td>
			<td>
bool
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
true
</pre>
</div>
			</td>
			<td>
				Specifies whether a service account should be created
			</td>
		</tr>
		<tr>
			<td id="serviceAccount--name">
				<a href="./values.yaml#L51">serviceAccount.name</a>
            </td>
			<td>
string
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
""
</pre>
</div>
			</td>
			<td>
				The name of the service account to use. If not set and create is true, a name is generated using the fullname template
			</td>
		</tr>
		<tr>
			<td id="tolerations">
				<a href="./values.yaml#L119">tolerations</a>
            </td>
			<td>
list
</td>
			<td>
				<div style="max-width: 300px;">
<pre lang="json">
[]
</pre>
</div>
			</td>
			<td>
				Tolerations for pod assignment
			</td>
		</tr>
	</tbody>
</table>

## Advanced
You can edit the `values.yaml` file to enable an ingress or the autoscaling.
