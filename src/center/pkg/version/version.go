package version

// 版本信息 - 由构建系统注入
var (
	// Version 版本号
	Version = "0.9.0"
	// GitCommit Git提交哈希
	GitCommit = "unknown"
	// BuildTime 构建时间
	BuildTime = "unknown"
	// GoVersion Go版本
	GoVersion = "unknown"
)

// BuildInfo 构建信息
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

// GetBuildInfo 获取构建信息
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
	}
}

// GetVersion 获取版本号
func GetVersion() string {
	return Version
}

// IsRelease 是否为发布版本
func IsRelease() bool {
	return Version != "dev" && !containsDev(Version)
}

func containsDev(v string) bool {
	for i := 0; i < len(v)-2; i++ {
		if v[i:i+3] == "dev" {
			return true
		}
	}
	return false
}