FROM alpine:3.8

ARG version=0.1

ADD https://github.com/MySocialApp/k8s-dns-updater/releases/download/v${version}/k8s-dns-updater_${version}_linux_amd64.tar.gz /tmp/k8s-dns-updater.tar.gz
WORKDIR /tmp
CMD mkdir /etc/k8s-dns-updater && \
    tar -xvf /tmp/k8s-dns-updater.tar.gz && \
    mv k8s-dns-updater /usr/bin && \
    rm -Rf /tmp/*
ADD config.yaml.example /etc/k8s-dns-updater/
WORKDIR /etc/k8s-dns-updater/

CMD ["/bin/sh", "/usr/bin/k8s-dns-updater"]