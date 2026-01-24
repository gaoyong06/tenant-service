# Tenant Service Makefile (devops-tools integrated)

.PHONY: all

SERVICE_NAME=tenant-service
SERVICE_DISPLAY_NAME=Tenant Service
HTTP_PORT=8124
GRPC_PORT=9124
API_PROTO_DIR=api
API_PROTO_PATH=api/tenant_service/v1/tenant.proto
TEST_CONFIG=test/api/api-test-config.yaml

DEVOPS_TOOLS_DIR := $(shell cd .. && pwd)/devops-tools
include $(DEVOPS_TOOLS_DIR)/Makefile.common
