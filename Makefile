run: start-consul

start-consul:
	@docker ps --filter "name=dev-consul" --format '{{.Names}}' | grep -q 'dev-consul' || \
	docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul consul:1.13

stop: stop-consul

stop-consul:
	@docker ps --filter "name=dev-consul" --format '{{.Names}}' | grep -q 'dev-consul' && \
	docker stop dev-consul && docker rm dev-consul || \
	echo "Consul konteyneri zaten durmuş ya da mevcut değil"
