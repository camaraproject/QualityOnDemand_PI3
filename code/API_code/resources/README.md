# Resources for dev/test

This dir contains resources that aid in setup of development/test environment and it is not intended
for `production`.

The `mongodb` dir contains an example javascript that is used to provision MongoDB.

The `keycloak` dir contains an example realm called `sfn.camara` that will be imported by KeyCloak
during bring up. It is preconfigured with a clientId called `nftest` to be used with `ClientCredentials`
OAuth2 workflow.

The `docker` dir contains docker compose yaml file to bring up the containers.
