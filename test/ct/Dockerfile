#todo change this to be a stable path!
FROM artifactory.algol60.net/csm-docker/unstable/hms-test:3.0.0-20220225025438.c37b99e

COPY smoke/ /src/app
COPY functional/ /src/app
COPY tavern_global_config_ct_test.yaml /src/app/tavern_global_config_ct_test.yaml

ENV PATH="/src/libs:${PATH}"
ENV PATH="/src/app:${PATH}"

USER root
RUN chown  -R 65534:65534 /src
USER 65534:65534

# this is inherited from the hms-test container
ENTRYPOINT [ "entrypoint.sh" ]