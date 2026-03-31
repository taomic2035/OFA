// Package transfer - 文件分片传输
package transfer

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TransferConfig 传输配置
type TransferConfig struct {
	ChunkSize       int64         `json:"chunk_size"`        // 分片大小(默认1MB)
	MaxConcurrent   int           `json:"max_concurrent"`    // 最大并发传输数
	Timeout         time.Duration `json:"timeout"`          // 传输超时
	RetryCount      int           `json:"retry_count"`      // 重试次数
	RetryInterval   time.Duration `json:"retry_interval"`   // 重试间隔
	TempDir         string        `json:"temp_dir"`         // 临时目录
	ChecksumEnabled bool          `json:"checksum_enabled"` // 启用校验
	CompressEnabled bool          `json:"compress_enabled"` // 启用压缩
}

// FileMetadata 文件元数据
type FileMetadata struct {
	FileID       string            `json:"file_id"`
	FileName     string            `json:"file_name"`
	FileSize     int64             `json:"file_size"`
	ChunkCount   int               `json:"chunk_count"`
	ChunkSize    int64             `json:"chunk_size"`
	Checksum     string            `json:"checksum"`      // 完整文件校验和
	MimeType     string            `json:"mime_type"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// FileChunk 文件分片
type FileChunk struct {
	FileID      string    `json:"file_id"`
	ChunkIndex  int       `json:"chunk_index"`
	ChunkOffset int64     `json:"chunk_offset"`
	ChunkSize   int64     `json:"chunk_size"`
	ChunkData   []byte    `json:"chunk_data"`
	Checksum    string    `json:"checksum"`     // 分片校验和
	Timestamp   int64     `json:"timestamp"`
	IsLast      bool      `json:"is_last"`
}

// TransferState 传输状态
type TransferState struct {
	FileID        string        `json:"file_id"`
	Direction     string        `json:"direction"`      // upload, download
	Status        string        `json:"status"`         // pending, transferring, paused, completed, failed
	TotalChunks   int           `json:"total_chunks"`
	ReceivedChunks int          `json:"received_chunks"`
	ReceivedBytes  int64        `json:"received_bytes"`
	TotalBytes    int64         `json:"total_bytes"`
	Progress      float64       `json:"progress"`       // 0-100
	StartTime     time.Time     `json:"start_time"`
	EndTime       *time.Time    `json:"end_time,omitempty"`
	Error         string        `json:"error,omitempty"`
	Chunks        map[int]bool  `json:"chunks"`         // chunk_index -> received
}

// TransferManager 传输管理器
type TransferManager struct {
	config       TransferConfig
	transfers    map[string]*TransferState
	chunkBuffers map[string]map[int][]byte // file_id -> chunk_index -> data
	tempFiles    map[string]*os.File
	mu           sync.RWMutex
}

// NewTransferManager 创建传输管理器
func NewTransferManager(config TransferConfig) *TransferManager {
	// 默认配置
	if config.ChunkSize == 0 {
		config.ChunkSize = 1024 * 1024 // 1MB
	}
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 5
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Minute
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = 5 * time.Second
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}

	return &TransferManager{
		config:       config,
		transfers:    make(map[string]*TransferState),
		chunkBuffers: make(map[string]map[int][]byte),
		tempFiles:    make(map[string]*os.File),
	}
}

// === 文件发送 ===

// CreateUpload 创建上传任务
func (tm *TransferManager) CreateUpload(filePath string) (*FileMetadata, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 计算分片数量
	fileSize := stat.Size()
	chunkCount := int((fileSize + tm.config.ChunkSize - 1) / tm.config.ChunkSize)

	// 计算文件校验和
	checksum := ""
	if tm.config.ChecksumEnabled {
		checksum, err = tm.calculateFileChecksum(filePath)
		if err != nil {
			return nil, err
		}
	}

	// 生成文件ID
	fileID := generateFileID()

	// 创建元数据
	metadata := &FileMetadata{
		FileID:     fileID,
		FileName:   filepath.Base(filePath),
		FileSize:   fileSize,
		ChunkCount: chunkCount,
		ChunkSize:  tm.config.ChunkSize,
		Checksum:   checksum,
		MimeType:   detectMimeType(filePath),
		CreatedAt:  time.Now(),
	}

	// 创建传输状态
	tm.mu.Lock()
	tm.transfers[fileID] = &TransferState{
		FileID:      fileID,
		Direction:   "upload",
		Status:      "pending",
		TotalChunks: chunkCount,
		TotalBytes:  fileSize,
		Chunks:      make(map[int]bool),
	}
	tm.mu.Unlock()

	return metadata, nil
}

// ReadChunk 读取分片
func (tm *TransferManager) ReadChunk(filePath string, fileID string, chunkIndex int) (*FileChunk, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 计算偏移量
	offset := int64(chunkIndex) * tm.config.ChunkSize

	// 定位到偏移位置
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("定位失败: %w", err)
	}

	// 读取数据
	buffer := make([]byte, tm.config.ChunkSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("读取失败: %w", err)
	}

	chunkData := buffer[:n]

	// 计算分片校验和
	chunkChecksum := ""
	if tm.config.ChecksumEnabled {
		chunkChecksum = tm.calculateChunkChecksum(chunkData)
	}

	// 获取文件信息
	stat, _ := file.Stat()
	isLast := offset+int64(n) >= stat.Size()

	return &FileChunk{
		FileID:      fileID,
		ChunkIndex:  chunkIndex,
		ChunkOffset: offset,
		ChunkSize:   int64(n),
		ChunkData:   chunkData,
		Checksum:    chunkChecksum,
		Timestamp:   time.Now().UnixNano(),
		IsLast:      isLast,
	}, nil
}

// === 文件接收 ===

// CreateDownload 创建下载任务
func (tm *TransferManager) CreateDownload(metadata *FileMetadata) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 创建传输状态
	tm.transfers[metadata.FileID] = &TransferState{
		FileID:        metadata.FileID,
		Direction:     "download",
		Status:        "pending",
		TotalChunks:   metadata.ChunkCount,
		TotalBytes:    metadata.FileSize,
		Chunks:        make(map[int]bool),
	}

	// 创建分片缓冲
	tm.chunkBuffers[metadata.FileID] = make(map[int][]byte)

	// 创建临时文件
	tempPath := filepath.Join(tm.config.TempDir, metadata.FileID+".tmp")
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tm.tempFiles[metadata.FileID] = tempFile

	// 预分配文件大小
	tempFile.Truncate(metadata.FileSize)

	return nil
}

// WriteChunk 写入分片
func (tm *TransferManager) WriteChunk(chunk *FileChunk) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	state, ok := tm.transfers[chunk.FileID]
	if !ok {
		return fmt.Errorf("传输任务不存在: %s", chunk.FileID)
	}

	// 验证分片校验和
	if tm.config.ChecksumEnabled && chunk.Checksum != "" {
		expectedChecksum := tm.calculateChunkChecksum(chunk.ChunkData)
		if expectedChecksum != chunk.Checksum {
			return fmt.Errorf("分片校验失败: chunk %d", chunk.ChunkIndex)
		}
	}

	// 获取临时文件
	tempFile, ok := tm.tempFiles[chunk.FileID]
	if !ok {
		return fmt.Errorf("临时文件不存在: %s", chunk.FileID)
	}

	// 定位并写入
	_, err := tempFile.Seek(chunk.ChunkOffset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("定位失败: %w", err)
	}

	_, err = tempFile.Write(chunk.ChunkData)
	if err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}

	// 更新状态
	state.Chunks[chunk.ChunkIndex] = true
	state.ReceivedChunks++
	state.ReceivedBytes += chunk.ChunkSize
	state.Progress = float64(state.ReceivedChunks) / float64(state.TotalChunks) * 100

	if state.Status == "pending" {
		state.Status = "transferring"
		state.StartTime = time.Now()
	}

	// 检查是否完成
	if state.ReceivedChunks == state.TotalChunks {
		state.Status = "completed"
		now := time.Now()
		state.EndTime = &now
	}

	return nil
}

// CompleteDownload 完成下载
func (tm *TransferManager) CompleteDownload(fileID, destPath string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	state, ok := tm.transfers[fileID]
	if !ok {
		return fmt.Errorf("传输任务不存在: %s", fileID)
	}

	if state.Status != "completed" {
		return fmt.Errorf("传输未完成")
	}

	// 关闭临时文件
	tempFile, ok := tm.tempFiles[fileID]
	if ok {
		tempFile.Close()
		delete(tm.tempFiles, fileID)
	}

	// 移动临时文件到目标位置
	tempPath := filepath.Join(tm.config.TempDir, fileID+".tmp")
	err := os.Rename(tempPath, destPath)
	if err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	// 清理
	delete(tm.chunkBuffers, fileID)

	return nil
}

// === 断点续传 ===

// GetTransferState 获取传输状态
func (tm *TransferManager) GetTransferState(fileID string) (*TransferState, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	state, ok := tm.transfers[fileID]
	if !ok {
		return nil, fmt.Errorf("传输任务不存在: %s", fileID)
	}
	return state, nil
}

// GetMissingChunks 获取缺失的分片索引
func (tm *TransferManager) GetMissingChunks(fileID string) []int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	state, ok := tm.transfers[fileID]
	if !ok {
		return nil
	}

	missing := make([]int, 0)
	for i := 0; i < state.TotalChunks; i++ {
		if !state.Chunks[i] {
			missing = append(missing, i)
		}
	}
	return missing
}

// PauseTransfer 暂停传输
func (tm *TransferManager) PauseTransfer(fileID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	state, ok := tm.transfers[fileID]
	if !ok {
		return fmt.Errorf("传输任务不存在: %s", fileID)
	}

	if state.Status == "transferring" {
		state.Status = "paused"
	}

	return nil
}

// ResumeTransfer 恢复传输
func (tm *TransferManager) ResumeTransfer(fileID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	state, ok := tm.transfers[fileID]
	if !ok {
		return fmt.Errorf("传输任务不存在: %s", fileID)
	}

	if state.Status == "paused" {
		state.Status = "transferring"
	}

	return nil
}

// CancelTransfer 取消传输
func (tm *TransferManager) CancelTransfer(fileID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 关闭临时文件
	if tempFile, ok := tm.tempFiles[fileID]; ok {
		tempFile.Close()
		delete(tm.tempFiles, fileID)
	}

	// 删除临时文件
	tempPath := filepath.Join(tm.config.TempDir, fileID+".tmp")
	os.Remove(tempPath)

	// 清理
	delete(tm.transfers, fileID)
	delete(tm.chunkBuffers, fileID)

	return nil
}

// === 校验机制 ===

// VerifyFile 验证文件完整性
func (tm *TransferManager) VerifyFile(filePath, expectedChecksum string) (bool, error) {
	if !tm.config.ChecksumEnabled {
		return true, nil
	}

	actualChecksum, err := tm.calculateFileChecksum(filePath)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}

// calculateFileChecksum 计算文件校验和
func (tm *TransferManager) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// calculateChunkChecksum 计算分片校验和
func (tm *TransferManager) calculateChunkChecksum(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// === 辅助函数 ===

// generateFileID 生成文件ID
func generateFileID() string {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b, uint64(time.Now().UnixNano()))
	return hex.EncodeToString(b)
}

// detectMimeType 检测MIME类型
func detectMimeType(filePath string) string {
	ext := filepath.Ext(filePath)
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".zip":  "application/zip",
		".json": "application/json",
		".xml":  "application/xml",
	}
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

// === 流式传输 ===

// ChunkStream 分片流
type ChunkStream struct {
	fileID    string
	filePath  string
	chunkSize int64
	total     int
	current   int
	file      *os.File
}

// CreateChunkStream 创建分片流
func (tm *TransferManager) CreateChunkStream(filePath string) (*ChunkStream, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	fileID := generateFileID()
	total := int((stat.Size() + tm.config.ChunkSize - 1) / tm.config.ChunkSize)

	return &ChunkStream{
		fileID:    fileID,
		filePath:  filePath,
		chunkSize: tm.config.ChunkSize,
		total:     total,
		current:   0,
		file:      file,
	}, nil
}

// Next 获取下一个分片
func (cs *ChunkStream) Next() (*FileChunk, error) {
	if cs.current >= cs.total {
		return nil, io.EOF
	}

	buffer := make([]byte, cs.chunkSize)
	n, err := cs.file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	chunk := &FileChunk{
		FileID:      cs.fileID,
		ChunkIndex:  cs.current,
		ChunkOffset: int64(cs.current) * cs.chunkSize,
		ChunkSize:   int64(n),
		ChunkData:   buffer[:n],
		Timestamp:   time.Now().UnixNano(),
		IsLast:      cs.current == cs.total-1,
	}

	cs.current++
	return chunk, nil
}

// Close 关闭流
func (cs *ChunkStream) Close() error {
	return cs.file.Close()
}

// === 统计 ===

// TransferStats 传输统计
type TransferStats struct {
	TotalTransfers   int64         `json:"total_transfers"`
	ActiveTransfers  int           `json:"active_transfers"`
	CompletedTransfers int64       `json:"completed_transfers"`
	FailedTransfers  int64         `json:"failed_transfers"`
	TotalBytes       int64         `json:"total_bytes"`
	AvgSpeed         int64         `json:"avg_speed"` // bytes/s
}

// GetStats 获取统计信息
func (tm *TransferManager) GetStats() *TransferStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	active := 0
	var completed, failed int64

	for _, state := range tm.transfers {
		switch state.Status {
		case "transferring", "pending", "paused":
			active++
		case "completed":
			completed++
		case "failed":
			failed++
		}
	}

	return &TransferStats{
		TotalTransfers:    int64(len(tm.transfers)),
		ActiveTransfers:   active,
		CompletedTransfers: completed,
		FailedTransfers:   failed,
	}
}

// ListTransfers 列出所有传输
func (tm *TransferManager) ListTransfers() []*TransferState {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transfers := make([]*TransferState, 0, len(tm.transfers))
	for _, state := range tm.transfers {
		transfers = append(transfers, state)
	}
	return transfers
}

// ExportState 导出状态(用于持久化)
func (tm *TransferManager) ExportState(fileID string) ([]byte, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	state, ok := tm.transfers[fileID]
	if !ok {
		return nil, fmt.Errorf("传输任务不存在: %s", fileID)
	}

	return json.Marshal(state)
}

// ImportState 导入状态(用于恢复)
func (tm *TransferManager) ImportState(data []byte) error {
	var state TransferState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	tm.mu.Lock()
	tm.transfers[state.FileID] = &state
	tm.chunkBuffers[state.FileID] = make(map[int][]byte)
	tm.mu.Unlock()

	return nil
}