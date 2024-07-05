proto-gen:
	@echo "********Compiling protobuf********"
	make -C quiz_lib/pb
	@echo ""
check-lint:
	golangci-lint run --config ci/lint_runner/.golangci.yml
fix-lint:
	golangci-lint run --config ci/lint_runner/.golangci.yml --fix
