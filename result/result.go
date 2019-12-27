package result 

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
