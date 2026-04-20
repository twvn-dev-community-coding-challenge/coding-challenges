.PHONY: test coverage

test:
	go test ./... -count=1

# Coverage for application packages (excludes root main and cmd/stressworker, which are integration entrypoints).
coverage:
	go test $$(go list ./... | grep -v vendor | grep -v node_modules | grep -v 'github.com/dotdak/sms-otp$$' | grep -v '/cmd/stressworker' | tr '\n' ' ') -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -1
