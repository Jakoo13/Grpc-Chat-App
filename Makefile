build_app:
	docker build --tag=docker_example .

run_app:
	docker run -it -p 8080:8080 docker_example