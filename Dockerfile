FROM armhf/debian

LABEL maintainer="amioranza@mdcnet.ninja"
LABEL description="controllerpi"

WORKDIR /

COPY controllerpi /controllerpi

ENTRYPOINT [ "/controllerpi" ]