// Package lite - 轻量级连接协议
package lite

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ProtocolVersion 协议版本
const ProtocolVersion = 0x01

// PacketType 数据包类型
type PacketType byte

const (
	PacketHandshake  PacketType = 0x01
	PacketHeartbeat  PacketType = 0x02
	PacketData       PacketType = 0x03
	PacketAck        PacketType = 0x04
	PacketError      PacketType = 0x0F
)

// PacketHeader 数据包头(8字节)
// Version(1) + Type(1) + Flags(1) + Reserved(1) + Length(2) + Seq(2)
type PacketHeader struct {
	Version  byte
	Type     PacketType
	Flags    byte
	Reserved byte
	Length   uint16
	Seq      uint16
}

// LiteProtocol 轻量级协议
type LiteProtocol struct {
	sequence   uint16
	encryptKey []byte
	compress   bool
	timeout    time.Duration
	mu         sync.Mutex
}

// NewLiteProtocol 创建协议实例
func NewLiteProtocol() *LiteProtocol {
	return &LiteProtocol{
		timeout: 10 * time.Second,
	}
}

// SetEncryption 设置加密密钥
func (p *LiteProtocol) SetEncryption(key []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.encryptKey = key
}

// SetCompress 设置压缩
func (p *LiteProtocol) SetCompress(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.compress = enabled
}

// EncodePacket 编码数据包
func (p *LiteProtocol) EncodePacket(pktType PacketType, data []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 压缩
	if p.compress && len(data) > 64 {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write(data)
		gz.Close()
		data = buf.Bytes()
	}

	// 加密
	if p.encryptKey != nil {
		encrypted, err := p.encrypt(data)
		if err != nil {
			return nil, err
		}
		data = encrypted
	}

	// 构建包
	p.sequence++
	header := PacketHeader{
		Version: ProtocolVersion,
		Type:    pktType,
		Length:  uint16(len(data)),
		Seq:     p.sequence,
	}

	// 标志位
	if p.compress {
		header.Flags |= 0x01
	}
	if p.encryptKey != nil {
		header.Flags |= 0x02
	}

	// 序列化
	buf := make([]byte, 8+len(data))
	buf[0] = header.Version
	buf[1] = byte(header.Type)
	buf[2] = header.Flags
	buf[3] = header.Reserved
	binary.BigEndian.PutUint16(buf[4:6], header.Length)
	binary.BigEndian.PutUint16(buf[6:8], header.Seq)
	copy(buf[8:], data)

	return buf, nil
}

// DecodePacket 解码数据包
func (p *LiteProtocol) DecodePacket(data []byte) (*PacketHeader, []byte, error) {
	if len(data) < 8 {
		return nil, nil, fmt.Errorf("数据长度不足")
	}

	header := &PacketHeader{
		Version:  data[0],
		Type:     PacketType(data[1]),
		Flags:    data[2],
		Reserved: data[3],
		Length:   binary.BigEndian.Uint16(data[4:6]),
		Seq:      binary.BigEndian.Uint16(data[6:8]),
	}

	if len(data) < int(8+header.Length) {
		return nil, nil, fmt.Errorf("数据不完整")
	}

	payload := data[8 : 8+header.Length]

	// 解密
	if header.Flags&0x02 != 0 && p.encryptKey != nil {
		decrypted, err := p.decrypt(payload)
		if err != nil {
			return nil, nil, err
		}
		payload = decrypted
	}

	// 解压
	if header.Flags&0x01 != 0 {
		gz, err := gzip.NewReader(bytes.NewReader(payload))
		if err != nil {
			return nil, nil, err
		}
		defer gz.Close()

		var buf bytes.Buffer
		io.Copy(&buf, gz)
		payload = buf.Bytes()
	}

	return header, payload, nil
}

// encrypt 加密(AES简化版)
func (p *LiteProtocol) encrypt(data []byte) ([]byte, error) {
	// 实际实现应使用AES-GCM
	// 这里简化处理
	return data, nil
}

// decrypt 解密
func (p *LiteProtocol) decrypt(data []byte) ([]byte, error) {
	return data, nil
}

// NextSequence 获取下一个序列号
func (p *LiteProtocol) NextSequence() uint16 {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sequence++
	return p.sequence
}

// === TCP连接 ===

// TCPConnection TCP连接实现
type TCPConnection struct {
	addr     string
	conn     net.Conn
	protocol *LiteProtocol
	recvChan chan []byte
	sendChan chan []byte
	errChan  chan error
	closed   bool
	mu       sync.Mutex
}

// NewTCPConnection 创建TCP连接
func NewTCPConnection(addr string) *TCPConnection {
	return &TCPConnection{
		addr:     addr,
		protocol: NewLiteProtocol(),
		recvChan: make(chan []byte, 10),
		sendChan: make(chan []byte, 10),
		errChan:  make(chan error, 1),
	}
}

// Connect 连接
func (tc *TCPConnection) Connect(ctx context.Context) error {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", tc.addr)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	tc.mu.Lock()
	tc.conn = conn
	tc.closed = false
	tc.mu.Unlock()

	// 握手
	if err := tc.handshake(); err != nil {
		conn.Close()
		return err
	}

	// 启动读写协程
	go tc.readLoop()
	go tc.writeLoop()

	return nil
}

// handshake 握手
func (tc *TCPConnection) handshake() error {
	// 发送握手包
	handshakeData := []byte{ProtocolVersion, 0x00, 0x00, 0x00} // 版本+特性
	pkt, err := tc.protocol.EncodePacket(PacketHandshake, handshakeData)
	if err != nil {
		return err
	}

	tc.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := tc.conn.Write(pkt); err != nil {
		return fmt.Errorf("握手失败: %w", err)
	}

	// 等待响应
	buf := make([]byte, 1024)
	tc.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := tc.conn.Read(buf)
	if err != nil {
		return fmt.Errorf("握手响应失败: %w", err)
	}

	header, _, err := tc.protocol.DecodePacket(buf[:n])
	if err != nil {
		return err
	}

	if header.Type != PacketHandshake {
		return fmt.Errorf("无效握手响应")
	}

	return nil
}

// Disconnect 断开连接
func (tc *TCPConnection) Disconnect() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.closed {
		return
	}

	tc.closed = true
	if tc.conn != nil {
		tc.conn.Close()
	}
	close(tc.recvChan)
	close(tc.sendChan)
}

// Send 发送消息
func (tc *TCPConnection) Send(msg *LiteMessage) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.closed {
		return fmt.Errorf("连接已关闭")
	}

	pkt, err := tc.protocol.EncodePacket(PacketData, msg.Encode())
	if err != nil {
		return err
	}

	tc.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = tc.conn.Write(pkt)
	return err
}

// Receive 接收消息
func (tc *TCPConnection) Receive() (*LiteMessage, error) {
	select {
	case data := <-tc.recvChan:
		return DecodeMessage(data)
	case err := <-tc.errChan:
		return nil, err
	}
}

// IsConnected 检查连接状态
func (tc *TCPConnection) IsConnected() bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return !tc.closed && tc.conn != nil
}

// readLoop 读循环
func (tc *TCPConnection) readLoop() {
	buf := make([]byte, 4096)
	for {
		tc.mu.Lock()
		conn := tc.conn
		closed := tc.closed
		tc.mu.Unlock()

		if closed {
			return
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			tc.errChan <- err
			return
		}

		header, payload, err := tc.protocol.DecodePacket(buf[:n])
		if err != nil {
			continue
		}

		switch header.Type {
		case PacketData:
			select {
			case tc.recvChan <- payload:
			default:
				// 缓冲区满，丢弃
			}
		case PacketHeartbeat:
			// 心跳，发送ACK
			ack, _ := tc.protocol.EncodePacket(PacketAck, nil)
			conn.Write(ack)
		case PacketAck:
			// 确认
		}
	}
}

// writeLoop 写循环
func (tc *TCPConnection) writeLoop() {
	for {
		tc.mu.Lock()
		closed := tc.closed
		tc.mu.Unlock()

		if closed {
			return
		}

		select {
		case data := <-tc.sendChan:
			tc.mu.Lock()
			conn := tc.conn
			tc.mu.Unlock()

			if conn != nil {
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				conn.Write(data)
			}
		}
	}
}

// === BLE连接(蓝牙低功耗) ===

// BLEConnection BLE连接实现(模拟)
type BLEConnection struct {
	deviceID   string
	charUUID   string
	connected  bool
	protocol   *LiteProtocol
	sendQueue  chan []byte
	recvQueue  chan []byte
	mu         sync.Mutex
}

// NewBLEConnection 创建BLE连接
func NewBLEConnection(deviceID, charUUID string) *BLEConnection {
	return &BLEConnection{
		deviceID:  deviceID,
		charUUID:  charUUID,
		protocol:  NewLiteProtocol(),
		sendQueue: make(chan []byte, 10),
		recvQueue: make(chan []byte, 10),
	}
}

// Connect 连接(平台实现)
func (bc *BLEConnection) Connect(ctx context.Context) error {
	// 实际实现需要调用平台BLE API
	// 这里是模拟
	bc.mu.Lock()
	bc.connected = true
	bc.mu.Unlock()
	return nil
}

// Disconnect 断开连接
func (bc *BLEConnection) Disconnect() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.connected = false
}

// Send 发送消息
func (bc *BLEConnection) Send(msg *LiteMessage) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if !bc.connected {
		return fmt.Errorf("未连接")
	}

	// BLE MTU限制，需要分片
	data := msg.Encode()
	return bc.sendChunks(data)
}

// sendChunks 分片发送
func (bc *BLEConnection) sendChunks(data []byte) error {
	// BLE MTU通常20字节
	mtu := 20
	chunkCount := (len(data) + mtu - 1) / mtu

	for i := 0; i < chunkCount; i++ {
		start := i * mtu
		end := start + mtu
		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		// 实际发送由平台实现
		_ = chunk
	}

	return nil
}

// Receive 接收消息
func (bc *BLEConnection) Receive() (*LiteMessage, error) {
	select {
	case data := <-bc.recvQueue:
		return DecodeMessage(data)
	default:
		return nil, nil
	}
}

// IsConnected 检查连接状态
func (bc *BLEConnection) IsConnected() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.connected
}

// OnDataReceived 数据接收回调(平台调用)
func (bc *BLEConnection) OnDataReceived(data []byte) {
	select {
	case bc.recvQueue <- data:
	default:
	}
}

// === WebSocket连接 ===

// WSConnection WebSocket连接实现
type WSConnection struct {
	url      string
	protocol *LiteProtocol
	conn     interface{} // websocket.Conn
	connected bool
	mu       sync.Mutex
}

// NewWSConnection 创建WebSocket连接
func NewWSConnection(url string) *WSConnection {
	return &WSConnection{
		url:      url,
		protocol: NewLiteProtocol(),
	}
}

// Connect 连接
func (ws *WSConnection) Connect(ctx context.Context) error {
	// 实际实现需要使用websocket库
	ws.mu.Lock()
	ws.connected = true
	ws.mu.Unlock()
	return nil
}

// Disconnect 断开连接
func (ws *WSConnection) Disconnect() {
	ws.mu.Lock()
	ws.connected = false
	ws.mu.Unlock()
}

// Send 发送消息
func (ws *WSConnection) Send(msg *LiteMessage) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.connected {
		return fmt.Errorf("未连接")
	}

	// WebSocket发送二进制数据
	data := msg.Encode()
	_ = data
	return nil
}

// Receive 接收消息
func (ws *WSConnection) Receive() (*LiteMessage, error) {
	return nil, nil
}

// IsConnected 检查连接状态
func (ws *WSConnection) IsConnected() bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.connected
}

// === 工具函数 ===

// GenerateNonce 生成随机数
func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

// CalculateChecksum 计算校验和
func CalculateChecksum(data []byte) uint16 {
	var sum uint16
	for _, b := range data {
		sum += uint16(b)
	}
	return sum
}

// AESEncrypt AES加密(简化版)
func AESEncrypt(key, plaintext []byte) ([]byte, error) {
	// 实际实现需要使用crypto/aes
	// 这里返回原数据
	return plaintext, nil
}

// AESDecrypt AES解密
func AESDecrypt(key, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

// GCMEncrypt AES-GCM加密
func GCMEncrypt(key, nonce, plaintext []byte) ([]byte, error) {
	// 实际实现
	block, _ := cipher.NewGCM(nil)
	_ = block
	return plaintext, nil
}

// GCMDecrypt AES-GCM解密
func GCMDecrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}