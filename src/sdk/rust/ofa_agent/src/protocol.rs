//! Protocol Module
//! Sprint 29: Rust Agent SDK

use crate::message::Message;
use crate::error::{Error, Result};

/// 协议版本
pub const PROTOCOL_VERSION: &str = "8.1.0";
pub const MAGIC: &[u8; 3] = b"OFA";
pub const HEADER_SIZE: usize = 16;

/// 协议编码器
pub struct Protocol;

impl Protocol {
    /// 编码消息
    pub fn encode(msg: &Message) -> Result<Vec<u8>> {
        let json = serde_json::to_vec(msg)?;
        let header = Self::make_header(json.len(), &msg.msg_type);
        Ok([&header[..], &json[..]].concat())
    }

    /// 解码消息
    pub fn decode(data: &[u8]) -> Result<Message> {
        if data.len() < HEADER_SIZE {
            return Err(Error::Protocol("Data too short".to_string()));
        }

        let (header, body) = data.split_at(HEADER_SIZE);
        let (_, length) = Self::parse_header(header)?;

        if body.len() != length {
            return Err(Error::Protocol("Length mismatch".to_string()));
        }

        let msg: Message = serde_json::from_slice(body)?;
        Ok(msg)
    }

    fn make_header(length: usize, msg_type: &crate::message::MessageType) -> [u8; HEADER_SIZE] {
        let mut header = [0u8; HEADER_SIZE];

        // Magic (3 bytes)
        header[0..3].copy_from_slice(MAGIC);

        // Type (4 bytes)
        let type_str = match msg_type {
            crate::message::MessageType::Register => "reg ",
            crate::message::MessageType::Heartbeat => "hbt ",
            crate::message::MessageType::Task => "tsk ",
            crate::message::MessageType::TaskResult => "tr  ",
            crate::message::MessageType::Message => "msg ",
            crate::message::MessageType::Broadcast => "bct ",
            crate::message::MessageType::Discovery => "dsc ",
            crate::message::MessageType::Error => "err ",
            crate::message::MessageType::Ack => "ack ",
        };
        header[3..7].copy_from_slice(type_str.as_bytes());

        // Length (4 bytes)
        header[7..11].copy_from_slice(&(length as u32).to_be_bytes());

        // Version (4 bytes)
        header[12..16].copy_from_slice(b"8.1 ");

        header
    }

    fn parse_header(header: &[u8]) -> Result<(String, usize)> {
        if &header[0..3] != MAGIC {
            return Err(Error::Protocol("Invalid magic".to_string()));
        }

        let type_str = String::from_utf8_lossy(&header[3..7])
            .trim_end().to_string();
        let length = u32::from_be_bytes([header[7], header[8], header[9], header[10]]) as usize;

        Ok((type_str, length))
    }
}