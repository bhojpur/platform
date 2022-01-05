# Integration Tests

This directory containts Bhojpur.NET Platform's integration tests, including the framework that makes them possible.

Integration tests work by instrumenting Bhojpur.NET Platform's components to modify and verify its state.
Such tests are for example:
  - [create bucket] by executing code within bp-daemon's context that loads the config file,
    creates a remote storage instance, and attempts to create a bucket.
  - [start application] by obtaining a Bhojpur.NET Platform API token, calling "createApplication" and watching
    for successful startup events.
  - [task start] by starting an application using the bp-manager interface, instrumenting the
    application container and ensuring that tasks have run.

## Integrations
- instrumentation: agents that are compiled before/during the test, uploaded to a pod and executed there.
                   They communicate with the test using net/rpc.
- API access: to all internal APIs, including bp-manager, bp-daemon, image-builder, registry-facade, server
- DB access to the Bhojpur.NET Platform DB

## Running the tests

### Automatically at Bhojpur.NET

There is a [Bhojpur Piro job](../.piro/run-integration-tests.yaml) that runs the integration tests against `core-dev` preview environments.

 > For tests that require an existing user the framework tries to automatically select one from the DB.
 > - On preview envs make sure to create one before running tests against it!
 > - If it's important to use a certain user (with fixed settings, for example) pass the additional `username` parameter.

Example command:
```
piro job run github -j .piro/run-integration-tests.yaml -a namespace=staging-gpl-2658-int-tests -a version=gpl-2658-int-tests.57 -f
```

### Manually against a Kubernetes cluster

You may want to run tests to assert whether your Bhojpur.NET Platform installation is successfully integrated.

To test your Bhojpur.NET Platform installation:

1. Set your kubectl context to the cluster you want to test
2. Integrate the Bhojpur.NET Platform installation with OAuth for GitHub and/or GitLab, otherwise related tests may fail
3. Clone this repo, and `cd` to `./bhojpur/test`
4. Run the tests like so
  ```bash
  go test -v ./... \
  -kubeconfig=<path_to_kube_config_file> \
  -namespace=<namespace_where_bhojpur_net_is_installed> \
  -username=<a_user_in_the_bhojpur_net_database>
  ```