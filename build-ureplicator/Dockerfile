FROM openjdk:8-jre

ADD https://github.com/kelseyhightower/confd/releases/download/v0.15.0/confd-0.15.0-linux-amd64 /usr/local/bin/confd

ADD https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.3.1/jmx_prometheus_javaagent-0.3.1.jar /jmx_prometheus_javaagent-0.3.1.jar

COPY tmp/uReplicator-master/uReplicator-Distribution/target/uReplicator-Distribution-pkg /uReplicator

COPY tmp/uReplicator-master/config uReplicator/config

COPY confd /etc/confd

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && \
    chmod +x /usr/local/bin/confd && \
    chmod +x /uReplicator/bin/*.sh

ENV JAVA_OPTS "${JAVA_OPTS} -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:+UseCGroupMemoryLimitForHeap -XX:MaxRAMFraction=1"

ENTRYPOINT [ "/entrypoint.sh" ]