# Bhojpur.NET Platform - Self-Hosted

The [Bhojpur.NET Platform](https://www.bhojpur.net) is a software-as-a-service hosting environment that provides a tenant
with a ready-to-use Network, Security, and IT applications or services in the Cloud using a server cluster.

This [Helm](https://helm.sh) chart allows you to deploy and operate an instance of the Bhojpur.NET Platform on your
own Cloud-aware infrastructure.


## Prerequisites

- Kubernetes 1.13+
- Helm 3+


## Get Repo Info

```console
git clone https://github.com/bhojpur/platform
cd bhojpur/chart

helm repo add charts.bhojpur.net https://charts.bhojpur.net
helm repo add stable https://charts.helm.sh/stable
helm repo add minio https://helm.min.io/
helm repo update
helm dep up
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._


## Install Chart

```console
$ helm install bhojpur .
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._


## Dependencies

By default this chart installs additional, dependent charts:

- [stable/docker-registry](https://github.com/helm/charts/tree/master/stable/docker-registry)
- [stable/minio](https://github.com/minio/charts)
- [stable/mysql](https://github.com/helm/charts/tree/master/stable/mysql)

_See [configuration](#configuration) for options to replace those dependencies._

_See [helm dependency](https://helm.sh/docs/helm/helm_dependency/) for command documentation._


## Uninstall Chart

```console
$ helm uninstall bhojpur
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._


## Upgrading Chart

```console
$ helm upgrade --install bhojpur .
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._


## Recommended Configuration

The default installation of this Chart works out-of-the box in the majority of scenarios. The following section
introduces the most important options you likely want to review and tune for your particular use case.


### Ingress, Domain and HTTPS

| Parameter            | Description                                    | Default                                                 |
|----------------------|------------------------------------------------|---------------------------------------------------------|
| `hostname`           | The Hostname your installation is available at | `localhost`                                             |
| `certificatesSecret` | Configures certificates for your domain        | `{}`                                                    |

Compare [values.yaml](./values.yaml) for details.

For more details and a complete example using `hosts` see [here](https://docs.bhojpur.net/self-hosted/latest/install/configure-ingress/).


### OAuth

See [here](https://docs.bhojpur.net/self-hosted/latest/install/oauth/) on how to pre-configure OAuth providers.


### Database

The default installation comes with a MySQL that runs inside the same cluster.

See [here](https://docs.bhojpur.net/self-hosted/latest/install/database/) on how to configure a custom database.


### Storage

See [here](https://docs.bhojpur.net/self-hosted/latest/install/storage/) on how to configure a custom storage provider.


### Docker Registry

See [here](https://docs.bhojpur.net/self-hosted/latest/install/docker-registry/) on how to configure a custom docker registry.


## Advanced Configuration Reference

 > Note: This is work-in-progress.


### Node Filesystem Layout

See [here](https://docs.bhojpur.net/self-hosted/latest/install/nodes/) on how to configure custom node paths.

### Application sizing

See [here](https://docs.bhojpur.net/self-hosted/latest/install/applications/) on how to configure different application sizings.
