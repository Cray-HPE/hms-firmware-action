# FAS Test Environment

FAS has several testing environment paradigms:

 1. docker-compose based
 2. kubernetes based
 3. raw application based

FAS has several layers of testing:

 1. Unit tests
 2. Integration (API driven) tests
     3. Smoke tests
     4. functional tests

Eventually FAS testing may be expanded to include deployment tests, upgrade/downgrade tests, but for now the focus is on the operation and functional correctness of the FAS binary (as opposed to the pod). 

## Push button testing

### unit tests
To execute the unit tests in a docker-compose environment (as would be run in the pipeline) run:

``` $> ./runUnitTest.sh```

This will spin up the environment, build FAS unit tests and execute them, then exit 0 (pass) or exit 1 (fail).

### integration tests

<strong>TODO</strong> This is not yet available, but once it is, it will look like:

To execute the integration tests in a docker-compose environment (as would be run in the pipeline) run:

``` $> ./runIntegration.sh```

This will spin up the environment, build FAS, run the integration tests, then exit 0 (pass) or exit 1 (fail).

## Paradigms 
### Docker Compose
FAS comes with several docker-compose files that will help you setup a testing environment:

 * docker-compose.developer.environment.yaml
 * docker-compose.developer.full.yaml
 * docker-compose.test.integration.yaml
 * docker-compose.test.unit.yaml

The 'test' environments will spin up the testing environment with no ports exposed.  This testing environment is a full production like environment using: 

 * production hardware-state-manager with: 
     * a production like vault
     * postgres
 * production hm-collector with:
     * kafka
     * zookeeper
 * Redfish Simulator - via RTS, for red fish devices
     * with redis 

In the 'test' environments NO ports are exposed as they will be used in the pipeline.

The 'developer' environments will spin up the same environment but will have ports exposed, so you can interact with it directly as if you were inside the service mesh.

'.integration' & '.full' include the FAS pod, whereas the others do not; as  you may want to bring your own FAS.   Be sure to read the docker-compose files and the dockerfiles to get an idea of the environment that is required and the ENV VARs that must be set. 

### Kubernetes
Currently there are helm charts that describe this service, and that could be used in Vshasta or on a real system.  At the time of this writing those files aren't ready yet, but YOU could make it happen.

<strong>TODO</strong> update this section once kubernetes testing is available.

### RAW Application
In addition to running the FAS environment in a containerized environment a user could manually build and execute FAS as a binary on the local system.  In this configuration the user would still rely on the docker-compose environment using the `docker-compose.developer.environment.yaml` file to spin up the full environment.  The user would then ` go build` and run the binary.  This could be helpful if trying to attach a debugger or do iterative development.  In both cases is is VITAL the user understand that they must export the ENV VARs in the `unittesting.Dockerfile` or `integration.Dockerfile`; otherwise it is VERY likely that the connection to VAULT will fail!

## FAQ

### Why do I have to use `hms-creds` instead of `secrets/hms-creds` for the `VAULT_PATH` env var?

In a production system HSM puts all the hms-creds under the default kv storage secret, which is a v1 API and why that works with `secrets/hms-creds` in production. However locally that doesn't work and we have to manually create a new engine with v1 API capabilities and therefore that's why there is a difference in names.

Which is also why when you use the CLI on a production system you do `vault kv list secret/hms-creds/bla` and locally it's `vault kv list hms-creds/bla`, it all comes down to the kv store and the API version. 