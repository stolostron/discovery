<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Contributing guidelines](#contributing-guidelines)
    - [Contributions](#contributions)
    - [Certificate of Origin](#certificate-of-origin)
    - [DCO Sign Off](#dco-sign-off)
    - [Contributing A Patch](#contributing-a-patch)
    - [Issue and Pull Request Management](#issue-and-pull-request-management)
    - [Pre-check before submitting a PR](#pre-check-before-submitting-a-pr)
    - [Build images](#build-images)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Contributing guidelines

## Contributions

All contributions to the repository must be submitted under the terms of the [Apache Public License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

## DCO Sign Off

You must sign off your commit to state that you certify the [DCO](https://github.com/open-cluster-management-io/community/blob/main/DCO). To certify your commit for DCO, add a line like the following at the end of your commit message:

```
Signed-off-by: John Smith <john@example.com>
```

This can be done with the `--signoff` option to `git commit`. See the [Git documentation](https://git-scm.com/docs/git-commit#Documentation/git-commit.txt--s) for details.

## Contributing A Patch

1. Submit an issue describing your proposed change to the repo in question.
1. The [repo owners](OWNERS) will respond to your issue promptly.
1. Fork the desired repo, develop and test your code changes.
1. Submit a pull request.

## Issue and Pull Request Management

Anyone may comment on issues and submit reviews for pull requests. However, in
order to be assigned an issue or pull request, you must be a member of the
[stolostron](https://github.com/stolostron) GitHub organization.

Repo maintainers can assign you an issue or pull request by leaving a
`/assign <your Github ID>` comment on the issue or pull request.

## Pull Request Etiquette

Anyone may submit a pull request, although in order to ensure that the pull request 
can be responded to properly, link the associated issue in the PR.

We request that all PRs which contribute new code into the repo also contain their 
associated unit and function tests.


## Pre-check before submitting a PR

After your PR is ready to commit, please run following commands to check your code.

```shell
make manifests              ## Regenerate manifests if necessary
make test                   ## Run unit tests
make deploy-and-test        ## Run functional tests
                        OR
make verify                 ## Runs all of the above commands
```

## Build images

Make sure your code build passed.

```shell
make manifests              ## Regenerate manifests if necessary
make docker-build           ## Ensure build succeeds
```

Now, you can follow the [README](./README.md) to work with the stolostron/discovery API repository.

## Running tests

The functional tests use a test server to mock OCM responses. Before running the tests both the discovery-operator
and the mock-ocm-server must be running against a cluster.

### Outside the cluster

Follow the instructions for installing the operator outside the cluster. Then, in a separate terminal, run the testserver.

```shell
make server-run
```

With both components running, initiate the tests with the following command:

```shell
make integration-tests-local
```

### Inside the cluster

Follow the instructions for installing the operator inside the cluster. Then do the same for the test server.

1. Build the image:
    ```shell
    make server-docker-build SERVER_URL=<registry>/<imagename>:<tag>
    ```
2. Push the image:
    ```shell
    make server-docker-push SERVER_URL=<registry>/<imagename>:<tag>
    ```
3. Deploy the server:
    ```shell
    make server-deploy SERVER_URL=<registry>/<imagename>:<tag>
    ```

With both deployments running in the same namespace, initiate the tests with the following command:

```shell
make integration-tests
```
