package desktop

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// SystemInfoSkill provides system information
type SystemInfoSkill struct{}

func (s *SystemInfoSkill) ID() string          { return "system.info" }
func (s *SystemInfoSkill) Name() string        { return "System Information" }
func (s *SystemInfoSkill) Operations() []string { return []string{"get", "cpu", "memory", "disk", "os", "all"} }

func (s *SystemInfoSkill) Execute(operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "get", "all":
		return s.getAllInfo()
	case "cpu":
		return s.getCPUInfo()
	case "memory":
		return s.getMemoryInfo()
	case "disk":
		return s.getDiskInfo(params)
	case "os":
		return s.getOSInfo()
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (s *SystemInfoSkill) getAllInfo() (interface{}, error) {
	return map[string]interface{}{
		"os":      s.getOSInfo(),
		"cpu":     s.getCPUInfo(),
		"memory":  s.getMemoryInfo(),
		"runtime": runtime.Version(),
	}, nil
}

func (s *SystemInfoSkill) getCPUInfo() interface{} {
	return map[string]interface{}{
		"arch":    runtime.GOARCH,
		"cpus":    runtime.NumCPU(),
	}
}

func (s *SystemInfoSkill) getMemoryInfo() interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_mb":      m.Alloc / 1024 / 1024,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":        m.Sys / 1024 / 1024,
		"num_gc":        m.NumGC,
	}
}

func (s *SystemInfoSkill) getDiskInfo(params map[string]interface{}) interface{} {
	path := "/"
	if p, ok := params["path"].(string); ok {
		path = p
	}

	total, free, used, err := getDiskUsage(path)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"path":         path,
		"total_gb":     total / 1024 / 1024 / 1024,
		"free_gb":      free / 1024 / 1024 / 1024,
		"used_gb":      used / 1024 / 1024 / 1024,
		"used_percent": float64(used) / float64(total) * 100,
	}
}

func (s *SystemInfoSkill) getOSInfo() interface{} {
	return map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": runtime.Version(),
	}
}

// FileOperationSkill provides file system operations
type FileOperationSkill struct{}

func (s *FileOperationSkill) ID() string          { return "file.operation" }
func (s *FileOperationSkill) Name() string        { return "File Operations" }
func (s *FileOperationSkill) Operations() []string { return []string{"read", "write", "delete", "list", "exists", "mkdir", "copy", "move"} }

func (s *FileOperationSkill) Execute(operation string, params map[string]interface{}) (interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter required")
	}

	switch operation {
	case "read":
		return s.readFile(path)
	case "write":
		content, _ := params["content"].(string)
		return s.writeFile(path, content)
	case "delete":
		return s.deleteFile(path)
	case "list":
		return s.listFiles(path)
	case "exists":
		return s.fileExists(path), nil
	case "mkdir":
		return s.makeDir(path)
	case "copy":
		dest, _ := params["destination"].(string)
		return s.copyFile(path, dest)
	case "move":
		dest, _ := params["destination"].(string)
		return s.moveFile(path, dest)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (s *FileOperationSkill) readFile(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"content": string(data),
		"size":    len(data),
	}, nil
}

func (s *FileOperationSkill) writeFile(path, content string) (interface{}, error) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"success": true,
		"size":    len(content),
	}, nil
}

func (s *FileOperationSkill) deleteFile(path string) (interface{}, error) {
	err := os.Remove(path)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true}, nil
}

func (s *FileOperationSkill) listFiles(path string) (interface{}, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, map[string]interface{}{
			"name":     entry.Name(),
			"is_dir":   entry.IsDir(),
			"size":     info.Size(),
			"modified": info.ModTime(),
		})
	}

	return map[string]interface{}{
		"path":  path,
		"files": files,
		"count": len(files),
	}, nil
}

func (s *FileOperationSkill) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (s *FileOperationSkill) makeDir(path string) (interface{}, error) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true}, nil
}

func (s *FileOperationSkill) copyFile(src, dest string) (interface{}, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(dest, data, 0644)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true}, nil
}

func (s *FileOperationSkill) moveFile(src, dest string) (interface{}, error) {
	err := os.Rename(src, dest)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true}, nil
}

// CommandSkill executes shell commands
type CommandSkill struct {
	dataDir string
}

func (s *CommandSkill) ID() string          { return "command.execute" }
func (s *CommandSkill) Name() string        { return "Command Execution" }
func (s *CommandSkill) Operations() []string { return []string{"run", "run_script"} }

func (s *CommandSkill) Execute(operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "run":
		cmd, ok := params["command"].(string)
		if !ok {
			return nil, fmt.Errorf("command parameter required")
		}
		return s.runCommand(cmd, params)
	case "run_script":
		script, ok := params["script"].(string)
		if !ok {
			return nil, fmt.Errorf("script parameter required")
		}
		return s.runScript(script, params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (s *CommandSkill) runCommand(cmd string, params map[string]interface{}) (interface{}, error) {
	// Security: restrict allowed commands
	allowedCommands := []string{"ls", "dir", "echo", "date", "whoami", "pwd", "cat", "head", "tail"}
	cmdParts := strings.Fields(cmd)
	if len(cmdParts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	baseCmd := cmdParts[0]
	allowed := false
	for _, a := range allowedCommands {
		if baseCmd == a {
			allowed = true
			break
		}
	}

	if !allowed {
		return nil, fmt.Errorf("command not allowed: %s", baseCmd)
	}

	// Execute command
	var execCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		execCmd = exec.Command("cmd", "/c", cmd)
	} else {
		execCmd = exec.Command("sh", "-c", cmd)
	}

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"output":  string(output),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"output":  string(output),
	}, nil
}

func (s *CommandSkill) runScript(script string, params map[string]interface{}) (interface{}, error) {
	// Execute script file
	scriptPath := script
	if !filepath.IsAbs(script) {
		scriptPath = filepath.Join(s.dataDir, "scripts", script)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script not found: %s", script)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if strings.HasSuffix(script, ".ps1") {
			cmd = exec.Command("powershell", "-File", scriptPath)
		} else {
			cmd = exec.Command("cmd", "/c", scriptPath)
		}
	} else {
		if strings.HasSuffix(script, ".py") {
			cmd = exec.Command("python3", scriptPath)
		} else if strings.HasSuffix(script, ".js") {
			cmd = exec.Command("node", scriptPath)
		} else {
			cmd = exec.Command("sh", scriptPath)
		}
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"output":  string(output),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"output":  string(output),
	}, nil
}

// EchoSkill echoes input back
type EchoSkill struct{}

func (s *EchoSkill) ID() string          { return "echo" }
func (s *EchoSkill) Name() string        { return "Echo" }
func (s *EchoSkill) Operations() []string { return []string{"echo", "ping"} }

func (s *EchoSkill) Execute(operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "echo":
		message, _ := params["message"].(string)
		return map[string]interface{}{
			"message":   message,
			"timestamp": time.Now(),
		}, nil
	case "ping":
		return map[string]interface{}{
			"pong":      true,
			"timestamp": time.Now(),
		}, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// ExecuteScript executes a script file
func ExecuteScript(scriptPath, operation string, params map[string]interface{}) (interface{}, error) {
	// Read script file
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	// Create input
	input := map[string]interface{}{
		"operation": operation,
		"params":    params,
	}
	inputJSON, _ := json.Marshal(input)

	// Execute script
	var cmd *exec.Cmd
	ext := strings.ToLower(filepath.Ext(scriptPath))

	switch ext {
	case ".py":
		cmd = exec.Command("python3", "-c", string(data))
		cmd.Stdin = strings.NewReader(string(inputJSON))
	case ".js":
		cmd = exec.Command("node", "-e", string(data))
		cmd.Stdin = strings.NewReader(string(inputJSON))
	case ".sh":
		cmd = exec.Command("sh", "-c", string(data))
		cmd.Stdin = strings.NewReader(string(inputJSON))
	default:
		return nil, fmt.Errorf("unsupported script type: %s", ext)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse output
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return map[string]interface{}{"output": string(output)}, nil
	}

	return result, nil
}

// ExecuteBinary executes a binary skill
func ExecuteBinary(binaryPath, operation string, params map[string]interface{}) (interface{}, error) {
	paramsJSON, _ := json.Marshal(params)

	cmd := exec.Command(binaryPath, operation)
	cmd.Stdin = strings.NewReader(string(paramsJSON))

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return map[string]interface{}{"output": string(output)}, nil
	}

	return result, nil
}