FROM fedora:33 
RUN yum update -y && yum install jq -y
RUN echo 'export TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)' >> ~/.bashrc
COPY log.sh /
RUN chmod +x /log.sh
ARG APISERVER
ENV APISERVER ${APISERVER}
ENTRYPOINT ["/bin/bash"]
