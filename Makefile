# Tenant Service Makefile (devops-tools integrated)

.PHONY: all

SERVICE_NAME=tenant-service
SERVICE_DISPLAY_NAME=Tenant Service
HTTP_PORT=8002
GRPC_PORT=9002
API_PROTO_DIR=
API_PROTO_PATH=
TEST_CONFIG=

DEVOPS_TOOLS_DIR := $(shell cd .. && pwd)/devops-tools
include $(DEVOPS_TOOLS_DIR)/Makefile.common
