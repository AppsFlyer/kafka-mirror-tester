include go.mk
include k8s.mk

####################
# uReplicator docker
####################
ureplicator-release:
	cd build-ureplicator; make release
