# Pre requisites 

* Go 1.13
* Artifactory 

# Execute

1. set the following environment variables 

```
  ART_URL="http://192.168.51.51:8081/artifactory/")
  ART_USER="admin"
  ART_PASS="password"
  LOG_LEVEL="DEBUG"

```

2. in the folder, run 

```
$ ts=$(date --rfc-3339=seconds); go run main.go yann-build-info 104 "$ts" mvn-greeting/0.0.1
```

it should pull Go modules (dependencies) and generate a buildinfo.json

