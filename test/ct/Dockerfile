#FROM artifactory.algol60.net/csm-docker/unstable/hms-test:3.0.0-20220223143802.52b701b
#todo change this!
FROM hms-test:3.0.0

COPY smoke/ /src/app
COPY functional/ /src/app

ENV PATH="/src/libs:${PATH}"
ENV PATH="/src/utils:${PATH}"
ENV PATH="/src/app:${PATH}"

USER root
RUN chown  -R 65534:65534 /src
USER 65534:65534

# this is inherited from the hms-test container
ENTRYPOINT [ "entrypoint.sh" ]