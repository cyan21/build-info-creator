package main

import (
  "fmt"
  "strconv"
  "bytes"
  "time"
  "os"
  "strings"
  "encoding/json"
  "io/ioutil"
  "github.com/jfrog/jfrog-client-go/utils/log"
  "github.com/jfrog/jfrog-client-go/artifactory"
  "github.com/jfrog/jfrog-client-go/artifactory/services"
  utils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
  "github.com/jfrog/jfrog-client-go/artifactory/auth"
)


const FULL_AQL4DOCKER string = "items.find({'path': { '$match' : 'IMAGE_PATH'}, 'type':'file' }).include('name','sha256', 'actual_sha1', 'actual_md5')"
 
const AQL4DOCKER string = "{'path': { '$match' : 'IMAGE_PATH'}, 'type':'file'}"
const TIMESTAMP_FORMAT string = "2006-01-02T15:04:05.000Z07:00"

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
  Name string `json:"build.name"`
  Number string `json:"build.number"`
  Timestamp string `json:"build.timestamp"`

}

type Artifact struct {
  Name string `json:"name"`
  Sha256 string `json:"sha256"`
  Sha1 string `json:"sha1"`
  Md5 string  `json:"md5"`
  Properties BuildInfoProperty `json:"properties"`
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
  bi.Name = biName 
  bi.Number = biNumber 

//  fmt.Println("[NewBuildInfo] biStart: ", biStart)
  bi.Started =  biStart 
//  fmt.Println("[NewBuildInfo] bi.Started: ", bi.Started)
  bi.DurationMillis = biDuration 
  bi.Principal = biPrincipal 

  // build info version
  bi.Version = "1.0.1" 
  bi.BuildType = "GENERIC" 
  
  return bi 
}

func (bi * BuildInfo) setModules (moduleName string, buildName string, buildNumber string, buildTimestamp string, arrRes *AQLResult) {

  bi.Modules = make([]Module, 1)
  bi.Modules[0].Id = moduleName 
   
//  fmt.Println("arrRes size: ", len((*arrRes).Results))
  bi.Modules[0].Artifacts = make([]Artifact, len((*arrRes).Results))

  i:= 0
  startBi, _ := time.Parse(time.RFC3339, buildTimestamp)
  epochMs := strconv.FormatInt(startBi.Unix()*1000, 10)

  for _,res  := range arrRes.Results {
    bi.Modules[0].Artifacts[i].Sha256 = res.Sha256 
    bi.Modules[0].Artifacts[i].Sha1 = res.Actual_sha1 
    bi.Modules[0].Artifacts[i].Md5 = res.Actual_md5 
    bi.Modules[0].Artifacts[i].Name = res.Name
    bi.Modules[0].Artifacts[i].Properties = BuildInfoProperty{buildName, buildNumber, epochMs }
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
  buildName string
  buildNumber string
  buildTimestamp string
  aql string
  rtManager *artifactory.ArtifactoryServicesManager
}

func NewBuildInfoCreator() *buildInfoCreator {
  
  var err1 error

  // read param file and extract data

  bic := new(buildInfoCreator) 
  bic.imageId = "mvn-greeting/0.0.1"
  bic.buildName = "yann-mvn"
  bic.buildNumber = "101"
  // expecting result of date --rfc-3339=seconds 
  biTimestamp := "2019-11-06 14:14:22+01:00"
  tmpTS, _ := time.Parse(time.RFC3339, strings.Replace(biTimestamp, " ", "T", -1))
  bic.buildTimestamp = tmpTS.Format(TIMESTAMP_FORMAT)
  
//  fmt.Println("[NewBuildInfoCreator] tmpTS: ", tmpTS)
//  fmt.Println("[NewBuildInfoCreator] bic.buildTimestamp: ", bic.buildTimestamp)

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
   
  // run AQL query
  bic.aql = strings.Replace(strings.Replace(FULL_AQL4DOCKER, "IMAGE_PATH", bic.imageId, -1), "'", "\"", -1)
  log.Debug("AQL query", bic.aql)

  log.Debug("Init done")  

  return bic 
}

func (bic *buildInfoCreator) process() {

  var arrRes AQLResult

  toParse, aql_err := bic.rtManager.Aql(bic.aql)

//  fmt.Println("AQL result", string(toParse))

  if aql_err != nil {
    log.Error(aql_err)
  }

  err1 := json.Unmarshal(toParse, &arrRes)

//  fmt.Println(arrRes)

  if err1 != nil {
    log.Error("Issue while unmarshalling")
  } 

  myBuild := NewBuildInfo(bic.buildName, bic.buildNumber, bic.buildTimestamp, "360000", "yannc")
  myBuild.setModules(bic.imageId, bic.buildName, bic.buildNumber, bic.buildTimestamp, &arrRes)
 
//  myBuild.print()

  buildinfo_json, _ := json.MarshalIndent(myBuild, "", " ")

//  fmt.Println(buildinfo_json)

  _ = ioutil.WriteFile("buildinfo.json", buildinfo_json, 0644)

  // set result in BuildInfo
//  fmt.Println("imageID:",bic.imageId)
}

func (bic *buildInfoCreator) setBuildInfoProps() {

  var buffer bytes.Buffer
  tmpTS, _ := time.Parse(time.RFC3339, bic.buildTimestamp)

  buffer.WriteString("build.name=")
  buffer.WriteString(bic.buildName)
  buffer.WriteString(";build.number=")
  buffer.WriteString(bic.buildNumber)
  buffer.WriteString(";build.timestamp=")
  buffer.WriteString(strconv.FormatInt(tmpTS.Unix()*1000, 10))

  searchParams := services.NewSearchParams()

  searchParams.Aql = utils.Aql{strings.Replace(strings.Replace(AQL4DOCKER, "'", "\"", -1), "IMAGE_PATH", bic.imageId, -1)}

  resultItems, err := bic.rtManager.SearchFiles(searchParams)

  if err != nil {
    fmt.Println(err)
  } 

  propsParams := services.NewPropsParams()
  propsParams.Items = resultItems
  propsParams.Props = buffer.String() 

  bic.rtManager.SetProps(propsParams)

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
  bc.setBuildInfoProps()
}
