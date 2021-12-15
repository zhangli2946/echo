.PHONY:

image:
	@docker build -t zhangli2946/echo:develop .

server: .PHONY
	@go build -mod=vendor -o server echo/cmd/server

serverctl: .PHONY
	@go build -mod=vendor -o serverctl echo/cmd/serverctl