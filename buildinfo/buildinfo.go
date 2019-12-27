package buildinfo

import (
  "fmt"
  "strconv"
  "time"
  "github.com/cyan21/build-info-creator/result"
  "github.com/jfrog/jfrog-client-go/utils/log"
)

type Module struct {
  Id string `json:"id"`
  Artifacts []Artifact `json:"artifacts"`
  Dependencies []Dependency `json:"dependencies"`
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

type Dependency struct {
  Id string `json:"id"`
  Sha256 string `json:"sha256"`
  Sha1 string `json:"sha1"`
  Md5 string  `json:"md5"`
}

type BuildAgentInfo struct {
  name string 
  version string
}

type VcsInfo struct {
  url string 
  revision string
}

type BuildDep struct {
  Name string `json:"name"`
  Number string `json:"number"`
  Started string `json:"started"`
}
/*
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

type BuildResult struct {
  BuildName string  `json:"build.name"`
  BuildNumber string `json:"build.number"`
  BuildCreated string `json:"build.created"`
}

type AQLBuildResult struct {
  Results []BuildResult
}
*/
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
  BuildDependencies []BuildDep `json:"buildDependencies"`

}

func NewBuildInfo (biName string, biNumber string, biStart string, biDuration string, biPrincipal string) *BuildInfo {

  bi := new(BuildInfo)
  bi.Name = biName 
  bi.Number = biNumber 

  bi.Started =  biStart 
  bi.DurationMillis = biDuration 
  bi.Principal = biPrincipal 

  // build info version
  bi.Version = "1.0.1" 
  bi.BuildType = "GENERIC" 
  
  return bi 
}

func (bi * BuildInfo) SetModules (moduleName string, buildName string, buildNumber string, buildTimestamp string, arrRes *result.AQLResult) {

  bi.Modules = make([]Module, 1)
  bi.Modules[0].Id = moduleName 
   
  log.Debug("[SetModules] Result size: ", len((*arrRes).Results))
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

func (bi * BuildInfo) SetBuildDeps (arrDeps *result.AQLResult) {

  bi.Modules[0].Dependencies = make([]Dependency, len((*arrDeps).Results))
  for i := 0; i < len((*arrDeps).Results); i++ {
    bi.Modules[0].Dependencies[i].Sha256 = (*arrDeps).Results[i].Sha256 
    bi.Modules[0].Dependencies[i].Sha1 =   (*arrDeps).Results[i].Actual_sha1 
    bi.Modules[0].Dependencies[i].Md5 =    (*arrDeps).Results[i].Actual_md5 
    bi.Modules[0].Dependencies[i].Id =   (*arrDeps).Results[i].Name
  }
  
}

func (bi *BuildInfo) AddChildBuild(arrBuild * []result.BuildResult) {
  bi.BuildDependencies = make([]BuildDep, len(*arrBuild))
  log.Debug("[AddChildBuild] len: ", len(*arrBuild),"; cap: ", cap(*arrBuild))

  for i := 0; i < len(*arrBuild); i++ {
    bi.BuildDependencies[i].Name = (*arrBuild)[i].BuildName 
    bi.BuildDependencies[i].Number = (*arrBuild)[i].BuildNumber  
    bi.BuildDependencies[i].Started = (*arrBuild)[i].BuildCreated
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
