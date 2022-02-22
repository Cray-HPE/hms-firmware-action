FROM artifactory.algol60.net/csm-docker/unstable/hms-test:3.0.0-20220222173337.0adf3fd

COPY entrypoint.sh /usr/bin/entrypoint.sh
COPY smoke/ /src/app
COPY functional/ /src/app

ENV PATH="/src/libs:${PATH}"
ENV PATH="/src/utils:${PATH}"
ENV PATH="/src/app:${PATH}"

USER root
RUN chown  -R 65534:65534 /src
USER 65534:65534

ENTRYPOINT [ "entrypoint.sh" ]