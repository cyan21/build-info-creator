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

2. update the imageID (line 147) 

```
  bic.imageId = "mvn-greeting/0.0.1"
```

in the folder, run 

```
$ go run main.go
```
it should pull Go modules (dependencies) and generate a buildinfo.json

