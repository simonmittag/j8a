install:
	CGO_ENABLED=0 go install github.com/simonmittag/j8a/cmd/j8a

test:
	CGO_ENABLED=0 go test -v -bench .

performance:
	rm local.yml && circleci config process .circleci/local.yml > local.yml && circleci local execute -c local.yml localperformance
