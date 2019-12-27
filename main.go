package main 

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "strings"
  "strconv"
  "time"
  "github.com/jfrog/jfrog-client-go/utils/log"
  "github.com/jfrog/jfrog-client-go/artifactory"
  "github.com/jfrog/jfrog-client-go/artifactory/services"
  utils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
  "github.com/jfrog/jfrog-client-go/artifactory/auth"
//  "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
  custom "github.com/cyan21/build-info-creator/buildinfo"
  "github.com/cyan21/build-info-creator/result"
)


const FULL_AQL4DOCKER string = "items.find({'path': { '$match' : 'IMAGE_PATH'}, 'type':'file' }).include('name','sha256', 'actual_sha1', 'actual_md5')"
 
const AQL4DOCKER string = "{'path': { '$match' : 'IMAGE_PATH'}, 'type':'file'}"
const TIMESTAMP_FORMAT string = "2006-01-02T15:04:05.000Z07:00"

type buildInfoCreator struct {
  imageId string
  artUrl string
  artUser string
  artPass string
  buildName string
  buildNumber string
  buildTimestamp string
  aql string
  deps string
  rtManager *artifactory.ArtifactoryServicesManager
}



func NewBuildInfoCreator(buildName string, buildNumber string, buildTimestamp string, imageId string, deps string) *buildInfoCreator {
  
  var err1 error
  var bic *buildInfoCreator 
  
  // init log file
  file, _ := os.Create("./buildInfoCreator.log")

  if os.Getenv("LOG_LEVEL") != "" {   
    switch os.Getenv("LOG_LEVEL") {
      case "ERROR": log.SetLogger(log.NewLogger(log.ERROR, file))
      case "DEBUG": log.SetLogger(log.NewLogger(log.DEBUG, file))
      default: log.SetLogger(log.NewLogger(log.INFO, file))
    }
  } else {
    log.SetLogger(log.NewLogger(log.INFO, file))
  }

  // check env variable for connection to Artifactory
  if os.Getenv("ART_URL") != "" && os.Getenv("ART_USER") != "" && os.Getenv("ART_PASS") != "" {  
    bic = new(buildInfoCreator) 
  } else {
    log.Error("[NewBuildInfoCreator] ART_URL, ART_USER, ART_PASS are required environment variables !")
    os.Exit(2)
  } 

  //bic := new(buildInfoCreator) 
  bic.artUrl = os.Getenv("ART_URL")
  bic.artUser = os.Getenv("ART_USER")
  bic.artPass =os.Getenv("ART_PASS")
  bic.imageId = imageId 
  bic.buildName = buildName 
  bic.buildNumber = buildNumber 
  bic.deps = deps
  

  // set up connection to Artifactory
  rtDetails := auth.NewArtifactoryDetails()
  rtDetails.SetUrl(os.Getenv("ART_URL"))
  rtDetails.SetUser(os.Getenv("ART_USER"))
  rtDetails.SetPassword(os.Getenv("ART_PASS"))

  serviceConfig, err := artifactory.NewConfigBuilder().
    SetArtDetails(rtDetails).
    SetDryRun(false).
    Build()

  if err != nil {
    log.Error("[NewBuildInfoCreator] Init service config failed with url: ", os.Getenv("ART_URL"),", user: ",os.Getenv("ART_USER"))
  }   

  bic.rtManager, err1 = artifactory.New(&rtDetails, serviceConfig)

  if err1 != nil {
    log.Error("[NewBuildInfoCreator] Init Artifactory failed with url: ", os.Getenv("ART_URL"),", user: ",os.Getenv("ART_USER"))
  }   
   
  // run AQL query
  bic.aql = strings.Replace(strings.Replace(FULL_AQL4DOCKER, "IMAGE_PATH", bic.imageId, -1), "'", "\"", -1)
  log.Debug("[NewBuildInfoCreator] AQL query to retrieve docker image: ", bic.aql)

  // expecting result of date --rfc-3339=seconds 
  biTimestamp := buildTimestamp 
  log.Debug("[NewBuildInfoCreator] date formated to RFC3339 : ", biTimestamp)

  tmpTS, _ := time.Parse(time.RFC3339, strings.Replace(biTimestamp, " ", "T", -1))
  log.Debug("[NewBuildInfoCreator] date formated to RFC3339 without 'T': ", tmpTS)

  bic.buildTimestamp = tmpTS.Format(TIMESTAMP_FORMAT)
  log.Debug("[NewBuildInfoCreator] date formated for BuildInfo  ", bic.buildTimestamp)

  return bic 
}

func (bic *buildInfoCreator) generateBuildInfo() {

  var arrRes, arrDeps result.AQLResult

  log.Info("[generateBuildInfo] Running AQL: ", bic.aql, " ...")

  // Get docker layers of an image
  toParse, aql_err := bic.rtManager.Aql(bic.aql)

  if aql_err != nil {
    log.Error("[generateBuildInfo] Failed executing AQL query :", bic.aql)
    log.Error("[generateBuildInfo] Error message : ", aql_err)
  }

  err1 := json.Unmarshal(toParse, &arrRes)

  if err1 != nil {
    log.Error("[generateBuildInfo] Failed unmarshalling result of AQL query :", bic.aql )
    log.Error("[generateBuildInfo] Error message : ", err1)
  } 

  log.Info("[generateBuildInfo] AQL executed successfully")
  log.Debug("[generateBuildInfo] AQL result stored into array: ", arrRes)

  if bic.deps != "" {

    log.Info("[generateBuildInfo] Build Info dependencies found", bic.deps)
    log.Info("[generateBuildInfo] Running AQL: ", bic.aql, " ...")
    toParse, aql_err = bic.rtManager.Aql(buildAQLDeps(bic.deps))

    if aql_err != nil {
      log.Error("[generateBuildInfo] Failed executing AQL query for dependencies")
      log.Error("[generateBuildInfo] Error message : ", aql_err)
    }

    err1 = json.Unmarshal(toParse, &arrDeps)

    if err1 != nil {
      log.Error("[generateBuildInfo] Failed unmarshalling result of AQL query for deps")
      log.Error("[generateBuildInfo] Error message : ", err1)
    } 

    log.Info("[generateBuildInfo] AQL executed successfully")
    log.Debug("[generateBuildInfo] Stored AQL result for deps into array: ", arrDeps)
  } 

  log.Info("[generateBuildInfo] Initializing Build Info")
  myBuild := custom.NewBuildInfo(bic.buildName, bic.buildNumber, bic.buildTimestamp, "359999", "yannc")
  myBuild.SetModules(bic.imageId, bic.buildName, bic.buildNumber, bic.buildTimestamp, &arrRes)

  if bic.deps != "" {
    log.Info("[generateBuildInfo] Appending dependencies to Build Info ...")

    myBuild.SetBuildDeps(&arrDeps) 

    // check if Build Dependency is the result of another Build Info
    aql_start := "builds.find({\"module.artifact.name\": \""
    var aqlRes result.AQLBuildResult 
    arrBuildDeps := strings.Split(bic.deps, ",") 
    aqlBuildResult := make([]result.BuildResult, 0)

    for i := 0; i < len(arrBuildDeps); i++ {

      aql := aql_start + arrBuildDeps[i] + "\"}).include(\"name\",\"number\",\"created\")"
      log.Debug("[generateBuildInfo] Running AQL: ", aql, " ...")
      
      toParse, aql_err = bic.rtManager.Aql(aql)
      err1 = json.Unmarshal(toParse, &aqlRes)
 
      if err1 != nil {
        log.Error("[generateBuildInfo] Failed unmarshalling result of AQL for deps")
        log.Error("[generateBuildInfo] Error message : ", err1)
      } 

      log.Debug("[generateBuildInfo] Stored AQL result into array: ", aqlRes)

      if len(aqlRes.Results) > 0 {
        log.Debug("[generateBuildInfo] Build Name found:", aqlRes.Results[0].BuildName) 

        aqlBuildResult = append(aqlBuildResult, result.BuildResult{ aqlRes.Results[0].BuildName, aqlRes.Results[0].BuildNumber, aqlRes.Results[0].BuildCreated })
      }
    } 
    
    if len(aqlBuildResult) > 0 {
      log.Debug("[generateBuildInfo] Build Info dependencies: ", aqlBuildResult)
      myBuild.AddChildBuild(&aqlBuildResult) 
    }

    log.Info("[generateBuildInfo] Build Info dependencies added successfully")
  } 


//  myBuild.print()

  buildinfo_json, _ := json.MarshalIndent(myBuild, "", " ")

  log.Info("[generateBuildInfo] Generating buildinfo.json ...")
  _ = ioutil.WriteFile("buildinfo.json", buildinfo_json, 0644)
  log.Info("[generateBuildInfo] buildinfo.json successfully generated")

}

func (bic *buildInfoCreator) setBuildInfoProps() {

  log.Info("[setBuildInfoProps] Setting Build Info properties ... ")

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

  log.Info("[setBuildInfoProps] SAQL: ", searchParams.Aql)

  resultItems, err := bic.rtManager.SearchFiles(searchParams)

  if err != nil {
    log.Error("[setBuildInfoProps] AQL raised an error")
    log.Error("[setBuildInfoProps] Error message : ", err)
  } 

  propsParams := services.NewPropsParams()

  log.Debug("[setBuildInfoProps] artifact to be tagged with props: ", resultItems)
  propsParams.Items = resultItems

  log.Info("[setBuildInfoProps] Properties to be added: ", buffer.String())
  propsParams.Props = buffer.String() 

  bic.rtManager.SetProps(propsParams)
  log.Info("[setBuildInfoProps] Set Build Info properties done")

}

func (bic *buildInfoCreator) publishBuildInfo() {

  // reading build info file
  jsonFile, err := os.Open("buildinfo.json")
  if err != nil {
    log.Error("[publishBuildInfo] couldn't opened buildinfo.json", err)
  }
  defer jsonFile.Close()
  byteValue, _ := ioutil.ReadAll(jsonFile)
  
  // preparing HTTP request
  client := &http.Client{}
  req, err := http.NewRequest("PUT", bic.artUrl + "api/build", bytes.NewBuffer(byteValue))

  if err != nil {
    log.Error("[publishBuildInfo] error occured when creating HTTP request", err)
  }
  req.Header.Set("Content-Type", "application/json")
  req.SetBasicAuth(bic.artUser, bic.artPass)

  resp, err := client.Do(req)

  log.Info("[publishBuildInfo] result publish Build Info: ", resp)

  if err != nil {
    log.Error("[publishBuildInfo] error occured when sending HTTP request", err)
  } else {
    log.Info("[publishBuildInfo] Published BuildInfo successfully :", resp.Status)
  }
}

func buildAQLDeps(deps string) string {

  arrDeps := strings.Split(deps, ",")
  var buffer bytes.Buffer
  buffer.WriteString("items.find({\"$or\": [")

  // Build AQL query 
  /* ========================================
    items.find({
      "$or": [
        { "name": {"$eq" : "Carefirst.jpg"}},
        { "name": {"$eq" : "artifactory-papi-6.1.0.jar"}}
      ]
    }).include("sha256","actual_sha1","actual_md5","name")
   ===========================================*/

  for i := 0; i < len(arrDeps); i++ {
    buffer.WriteString("{ \"name\": {\"$eq\" : \"" + arrDeps[i] + "\"}},")
  }

  aql := strings.TrimSuffix(buffer.String(), ",")
  aql += "]}).include(\"sha256\",\"actual_sha1\",\"actual_md5\",\"name\")"
   
  log.Info("[buildAQLDeps] Generated AQL :", aql)

  return aql
}


/////////////////////////////////////////// 

func usage() {
  fmt.Println("[USAGE] ", os.Args[0], " buildName buildNumber buildTimestamp imageID")
  fmt.Println("\t buildName : any string")
  fmt.Println("\t buildNumber : any number")
  fmt.Println("\t buildTimestamp : formatted following 'date --rfc-3339=seconds' command")
  fmt.Println("\t imageID : imageName/tag Not imageName:tag")
  fmt.Println("\t dependencies : artifact names separated by comma")
}

func main() {

  deps := "" 

  if len(os.Args) < 5 {
     fmt.Println("[ERROR] missing parameters") 
     usage()
     os.Exit(2)
  } 

  if len(os.Args) == 6 { deps = os.Args[5] }

  var bc = NewBuildInfoCreator(os.Args[1], os.Args[2], os.Args[3], os.Args[4], deps)
  bc.generateBuildInfo()
//  bc.setBuildInfoProps()
//  bc.publishBuildInfo()
}
