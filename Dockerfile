FROM alpine:3.8

ARG version=0.3

ADD run.sh /
ADD https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 /sbin/dumb-init
ADD https://github.com/MySocialApp/k8s-dns-updater/releases/download/v${version}/k8s-dns-updater_${version}_linux_amd64.tar.gz /tmp/k8s-dns-updater.tar.gz

RUN apk add --update --no-cache libc6-compat bash && \
    cd /tmp && \
    mkdir /etc/k8s-dns-updater && \
    tar -xvf /tmp/k8s-dns-updater.tar.gz && \
    mv k8s-dns-updater /usr/bin && \
    chmod 755 /run.sh /sbin/dumb-init /usr/bin/k8s-dns-updater && \
    rm -Rf /tmp/* /var/cache/apk/*

ADD config.yaml.example /etc/k8s-dns-updater/config.yaml

CMD ["/sbin/dumb-init", "/bin/bash", "/run.sh"]