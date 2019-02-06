REV := `git rev-parse --short HEAD`


####################
# uReplicator docker
####################
U_WORK_DIR := tmp
U_BIN := ureplicator
U_IMAGE := rantav/$(U_BIN)

release: clean build image deploy clean

build:
	mkdir -p $(U_WORK_DIR)
	curl -sL https://github.com/uber/uReplicator/archive/master.tar.gz | tar xz -C $(U_WORK_DIR)
	cd $(U_WORK_DIR)/uReplicator-master && mvn package -DskipTests
	chmod u+x $(U_WORK_DIR)/uReplicator-master/bin/pkg/*.sh

image:
	docker build -t $(U_IMAGE):$(REV) .

deploy: ureplicator-image
	docker push $(U_IMAGE):$(REV)

clean:
	@/bin/rm -rf $(U_WORK_DIR)
