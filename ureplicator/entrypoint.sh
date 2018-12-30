#!/bin/bash -ex

if [[ "${LOGICAL_PROCESSORS}" == "" ]]; then
  LOGICAL_PROCESSORS=`getconf _NPROCESSORS_ONLN`
fi

export JAVA_OPTS="${JAVA_OPTS} -XX:ParallelGCThreads=${LOGICAL_PROCESSORS}"


confd -onetime -backend env

cd /uReplicator/bin/

if [ "${SERVICE_TYPE}" == "controller" ] ; then
  ./start-controller.sh \
    -port 9000 \
    -zookeeper "${HELIX_ZK_CONNECT}" \
    -helixClusterName "${HELIX_CLUSTER_NAME}" \
    -backUpToGit false \
    -autoRebalanceDelayInSeconds 120 \
    -localBackupFilePath /tmp/uReplicator-controller \
    -enableAutoWhitelist true \
    -enableAutoTopicExpansion true \
    -srcKafkaZkPath "${SRC_ZK_CONNECT}" \
    -destKafkaZkPath "${DST_ZK_CONNECT}" \
    -initWaitTimeInSeconds 10 \
    -refreshTimeInSeconds 20 \
    -env "${HELIX_ENV}"

  until [[ "OK" == "$(curl --silent http://localhost:9000/health)" ]]; do
    echo waiting
    sleep 1
  done

  TOPIC_LIST=( $(echo ${TOPICS} | sed "s/,/ /g") )
  PARTITION_LIST=( $(echo ${PARTITIONS} | sed "s/,/ /g") )

  for index in ${!TOPIC_LIST[*]}; do
    TOPIC="${TOPIC_LIST[$index]}"
    PARTITION="${PARTITION_LIST[$index]}"

    echo "Topic: ${TOPIC}, Partitions: ${PARTITION}"

    curl -X POST -d "{\"topic\": \"${TOPIC}\", \"numPartitions\": \"${PARTITION}\"}" http://localhost:9000/topics || true
  done

elif [ "${SERVICE_TYPE}" == "worker" ] ; then

  WORKER_ABORT_ON_SEND_FAILURE="${WORKER_ABORT_ON_SEND_FAILURE:=false}"

  ./start-worker.sh \
    --helix.config /uReplicator/config/helix.properties \
    --consumer.config /uReplicator/config/consumer.properties \
    --producer.config /uReplicator/config/producer.properties \
    --abort.on.send.failure="${WORKER_ABORT_ON_SEND_FAILURE}"

fi