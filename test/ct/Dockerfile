FROM artifactory.algol60.net/csm-docker/unstable/hms-test:2.0.0-20220214224310.c1517ff

COPY entrypoint.sh /usr/bin/entrypoint.sh
COPY ncn-smoke/ /opt/cray/tests/ncn-smoke/hms/hms-firmware-action/
COPY ncn-functional/ /opt/cray/tests/ncn-functional/hms/hms-firmware-action/

ENTRYPOINT [ "entrypoint.sh" ]