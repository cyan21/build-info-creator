package main

import (
  "fmt"
  "os"
  "strings"
  "encoding/json"
  "io/ioutil"
  "github.com/jfrog/jfrog-client-go/utils/log"
  "github.com/jfrog/jfrog-client-go/artifactory"
  "github.com/jfrog/jfrog-client-go/artifactory/auth"
)

///// AQL 

type Property struct {
  Key string
  Value string
}

type Result struct {
  Name string 
  Sha256 string
  Actual_sha1 string
  Actual_md5 string
//  properties []Property
}

type AQLResult struct {
  Results []Result
}

///// BuildInfo

type Module struct {
  Id string `json:"id"`
  Artifacts []Artifact `json:"artifacts"`
  Dependencies []Artifact `json:"dependencies"`
} 

type BuildInfoProperty struct {
  Timestamp string 
  Number string
  Name string 
}

type Artifact struct {
  Name string `json:"name"`
  Sha256 string `json:"sha256"`
  Sha1 string `json:"sha1"`
  Md5 string  `json:"md5"`
  Properties []BuildInfoProperty `json:"properties"`
}

type BuildAgentInfo struct {
  name string 
  version string
}

type VcsInfo struct {
  url string 
  revision string
}


/////////////////////////////////////////// 

type BuildInfo struct {
  Version string `json:"version"`
  Name string `json:"name"`
  Number string `json:"number"`
  BuildType string `json:"type"`
  BuildAgent BuildAgentInfo `json:"buildAgent"`
  Agent BuildAgentInfo`json:"agent"`
  Started string`json:"started"`
  DurationMillis string`json:"durationMillis"`
  Principal string`json:"principal"`
  ArtifactoryPrincipal string`json:"artifactoryPrincipal"`
  ArtifactoryPluginVersion string`json:"artifactoryPluginVersion"`
  Url string`json:"url"`
  Vcs []VcsInfo`json:"vcs"`
  VcsRevision string`json:"vcsRevision"`
  VcsUrl string `json:"vcsUrl"`
  Modules []Module `json:"modules"`
}

func NewBuildInfo (biName string, biNumber string, biStart string, biDuration string, biPrincipal string) *BuildInfo {
  bi := new(BuildInfo)

  // required params
  bi.Name = biName 
  bi.Number = biNumber 
  bi.Started = biStart 
  bi.DurationMillis = biDuration 
  bi.Principal = biPrincipal 

  // build info version
  bi.Version = "1.0.1" 
  bi.BuildType = "GENERIC" 
  
  return bi 
}

func (bi * BuildInfo) setModules (moduleName string, arrRes *AQLResult) {

  bi.Modules = make([]Module, 1)
  bi.Modules[0].Id = moduleName 
   
//  fmt.Println("arrRes size: ", len((*arrRes).Results))
  bi.Modules[0].Artifacts = make([]Artifact, len((*arrRes).Results))

  i:= 0

  for _,res  := range arrRes.Results {
    fmt.Println("layer name:", res.Actual_sha1)
    bi.Modules[0].Artifacts[i].Sha256 = res.Sha256 
    bi.Modules[0].Artifacts[i].Sha1 = res.Actual_sha1 
    bi.Modules[0].Artifacts[i].Md5 = res.Actual_md5 
    bi.Modules[0].Artifacts[i].Name = res.Name
    i++
  }
}

func (bi* BuildInfo) print() {

  for _,res  := range bi.Modules[0].Artifacts {
    fmt.Println("Artifact Name: ", res.Name)    
    fmt.Println("\tsha256: ", res.Sha256)    
    fmt.Println("\tsha1: ", res.Sha1)    
    fmt.Println("\tmd5: ", res.Md5)    
  }
}
/////////////////////////////////////////// 

type buildInfoCreator struct {
  imageId string
  rtManager *artifactory.ArtifactoryServicesManager
}

func NewBuildInfoCreator() *buildInfoCreator {
  
  var err1 error

  // read param file and extract data

  bic := new(buildInfoCreator) 
  bic.imageId = "mvn-greeting/0.0.1"

  // init log file
  file, _ := os.Create("./buildInfoCreator.log")
  log.SetLogger(log.NewLogger(log.DEBUG, file))

  // set up connection to Artifactory
  rtDetails := auth.NewArtifactoryDetails()
  rtDetails.SetUrl("http://192.168.51.51:8081/artifactory/")
  rtDetails.SetUser("admin")
  rtDetails.SetPassword("password")

  serviceConfig, err := artifactory.NewConfigBuilder().
    SetArtDetails(rtDetails).
    SetDryRun(false).
    Build()

  if err != nil {
    log.Error("Issue while initializing the service config")
  }   

  bic.rtManager, err1 = artifactory.New(&rtDetails, serviceConfig)

  if err1 != nil {
    fmt.Println("Issue while initializing the connection to Artifactory")
  }   
   
  log.Debug("Init done")  

  return bic 
}

func (bic *buildInfoCreator) process() {
  fmt.Println("process method")
  log.Debug("process method")  

  var arrRes AQLResult

  // run AQL query
  aql_template := "items.find({\"path\": { \"$match\" : \"IMAGE_PATH\"}, \"type\":\"file\" }).include(\"name\",\"sha256\", \"actual_sha1\", \"actual_md5\")"
 
  aql := strings.Replace(aql_template, "IMAGE_PATH", bic.imageId, -1) 
  log.Debug("AQL query", aql)
  toParse, aql_err := bic.rtManager.Aql(aql)

  fmt.Println("AQL result", string(toParse))

  if aql_err != nil {
    log.Error(aql_err)
  }

  err1 := json.Unmarshal(toParse, &arrRes)

  fmt.Println(arrRes)

  if err1 != nil {
    log.Error("Issue while unmarshalling")
  } 
 
  myBuild := NewBuildInfo("yann-build-info","777","1234567891011","360000","yannc")
  myBuild.setModules("simple-app:1.0.0",&arrRes)
 
  myBuild.print()

  buildinfo_json, _ := json.MarshalIndent(myBuild, "", " ")

//  fmt.Println(buildinfo_json)

  _ = ioutil.WriteFile("buildinfo.json", buildinfo_json, 0644)

  // set result in BuildInfo
  fmt.Println("imageID:",bic.imageId)
}

func (*buildInfoCreator) publish() {
  fmt.Println("publish method")
  log.Debug("publish method")  
  // hit on Artifactory Rest API  
}


/////////////////////////////////////////// 


func main() {

  var bc = NewBuildInfoCreator()
  bc.process()

}
