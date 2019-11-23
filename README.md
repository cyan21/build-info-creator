# Environment 

* Go 1.13
* Artifactory Pro 

# Pre requisites 

Set the following environment variables 

```
$ export ART_URL="http://192.168.51.51:8081/artifactory/" \
  ART_USER="admin" \
  ART_PASS="password" \
  LOG_LEVEL="DEBUG"
```
# Workflow 

1. Publish to Artifactory the result of your build

2. Run the program and specify the build name, build number, the build start date and the path to the published artifact(s) in Artifactory

3. The program will : 
  * generate a buildinfo.json
  * apply build properties to the published artifact(s)
  * publish the buidinfo.json to Artifactory


# Test the code 

> it should pull Go modules (dependencies) as well  

Run :
```
$ ts=$(date --rfc-3339=seconds); go run main.go yann-build-info 104 "$ts" mvn-greeting/0.0.1 "cve.json,binaries.jpg"
```

# Generate the exec 

Run the following command to generate "bic" binary:
```
$ go build -o bic main.go
```

To test it :
```
$ ts=$(date --rfc-3339=seconds); bic yann-build-info 104 "$ts" mvn-greeting/0.0.1 "cve.json,binaries.jpg"

```

# TO DO 

* improve logging
* use JFrog CLI go client to manage Build Info creation
* improve param management : 
  * Artifactory params are passed via env variables
  * Build Info params passed as arguments + need to fill other Build info field 
