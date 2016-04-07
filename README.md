# broadway
Broadway 

[![Build Status](https://travis-ci.org/namely/broadway.svg?branch=master)](https://travis-ci.org/namely/broadway)

Broadway helps developers deploy their projects with defining their deployment
workflows in Broadway playbooks. It runs as a service interacting with
Kubernetes and allowing users to interact with it through various interfaces
(Slack, CLI, Web).

## Playbook
A Broadway Playbook is a YAML file for defining a project's deployment tasks.

`id`, `name`, and at least one item in `tasks` are mandatory fields.

Each task item must have one or more of: a `manifests` array, a `pod_manifest`
item.

These manifest items must match .yml files in the `manifests` directory, e.g.
the "Deploy Postgres" task below expects files `manifests/postgres-rc.yml` and
`manifests/postgres-service.yml`.

```yaml
---
id: web
name: Web Project
meta:
  team: Web team
  email: webteam@namely.com
  slack: web
vars:
  - version
  - assets_version
  - owner
tasks:
  - name: Deploy Postgres
    manifests:
      - postgres-rc
      - postgres-service
  - name: Deploy Redis
    manifests:
      - redis-rc
      - redis-service
  - name: Database Migration
    pod_manifest: migration-pod
    wait_for:
      - success
  - name: Deploy Web
    manifests:
      - web-rc
      - web-service
      - worker-rc
```

## Setup
You should have prerequisites
[Kubernetes](http://kubernetes.io/docs/getting-started-guides/binary_release/)
and [Docker](https://docs.docker.com/engine/installation/) installed already.
You should have a running, active Docker machine in this terminal:

    $ kubectl version
      > "Client Version... Server Version..."
    $ docker-machine status default
      > "Running"

If you have an inactive docker machine, start it:

    $ eval $(docker-machine env default)

Clone the broadway repo and run the startup script:

    $ git clone https://github.com/namely/broadway; cd broadway
    $ ./broadway-dev-up.sh

After lots of docker container setup, you should see output:

    > ...
    > namespace "broadway" created

Now you should have a running Broadway server:
    $ curl localhost:8080
      > {
      >   "paths": [
      >     "/api",
      > ...

## Running Broadway

To run the Broadway server with your playbooks, you can use `broadwayctl` to start it up. The default directory for playbooks is `$(pwd)/playbooks`.

```sh
$ broadwayctl server --playbooks=./playbooks --addr=0.0.0.0:8080
=> starting broadway server...
=> loading playbooks...
```

This will load the directory of playbooks and ensure that everything is hunky dory.

## Instance
An instance represents a Broadway instance that may or may not be deployed.
Good usecase is when a CI server creates an instance in Broadway sending the
version then the user can deploy the instance from Slack.

Instance Statuses:
 - new
 - deploying
 - deployed
 - deleting
 - error

Instance Attributes:
 - playbook id – playbook identifier (String)
 - id – instance identifier (String)
 - status – instance status
 - created – when the instance was created
 - vars – map of String values



## API

1. Create or update Instance

User can post to `/instances` to create or update instances. We allow updates
via POST request to simplify the http interface.


Request:
```
POST /instances

{
  "playbook_id": "web",
  "id": "master",
  "vars": {
    "version": "dc231ba",
    "assets_version": "dc231ba",
    "owner": "bill"
  }
}
```

Response:
```
Status: 201 Created


{
  "playbook_id": "web",
  "id": "master",
  "status": "new",
  "vars": {
    "version": "dc231ba",
    "assets_version": "dc231ba",
    "owner": "bill"
  }
}
```


