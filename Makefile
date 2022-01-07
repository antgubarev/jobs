mocks:
	go generate ./...

coverage:
	go test --count=1 -coverprofile=coverage.out ./... ; \
		cat coverage.out | \
		awk 'BEGIN {cov=0; stat=0;} \
			$3!="" { cov+=($3==1?$2:0); stat+=$2; } \
    	END {printf("Total coverage: %.2f%% of statements\n", (cov/stat)*100);}'

test:
	go test ./...

lint:
	golangci-lint run