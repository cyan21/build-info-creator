# Pre requisites 

* Go 1.13
* Artifactory 

# Execute

1. update the artifactory setup (line 155 to 157)

```
  rtDetails.SetUrl("http://192.168.51.51:8081/artifactory/")
  rtDetails.SetUser("admin")
  rtDetails.SetPassword("password")


```

2. in the folder, run 

```
$ go run main.go yann-build-info 103 "2019-11-06 14:14:22+01:00" mvn-greeting/0.0.1
```

it should pull Go modules (dependencies) and generate a buildinfo.json

