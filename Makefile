export ART_URL=http://192.168.41.41:8081/artifactory/
export ART_USER=admin
export ART_PASS=password
export LOG_LEVEL=INFO

ts=$(date --rfc-3339=seconds)
img_id=my-hello-world/0.0.1

test: 
	go run main.go yann-build-info 104 "$(shell date --rfc-3339=seconds)" ${img_id}
clean: 
	rm -rf buildInfoCreator.log buildinfo.json
