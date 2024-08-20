FROM ubuntu:24.04

COPY deploy-controller /usr/local/bin/

RUN chmod +x /usr/local/bin/deploy-controller

CMD ["deploy-controller"]