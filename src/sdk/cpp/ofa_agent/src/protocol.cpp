/**
 * @file protocol.cpp
 * @brief Protocol Implementation
 * Sprint 29: C++ Agent SDK
 */

#include "ofa/protocol.hpp"
#include "ofa/error.hpp"
#include <cstring>

namespace ofa {

std::vector<uint8_t> Protocol::encode(const Message& msg) {
    json j = msg.toJson();
    std::string jsonStr = j.dump();
    std::vector<uint8_t> jsonBytes(jsonStr.begin(), jsonStr.end());

    auto header = makeHeader(jsonBytes.size(), msg.type);

    std::vector<uint8_t> result;
    result.reserve(header.size() + jsonBytes.size());
    result.insert(result.end(), header.begin(), header.end());
    result.insert(result.end(), jsonBytes.begin(), jsonBytes.end());

    return result;
}

std::optional<Message> Protocol::decode(const std::vector<uint8_t>& data) {
    if (data.size() < ProtocolConstants::HEADER_SIZE) {
        return std::nullopt;
    }

    auto [type, length] = parseHeader(data.data());

    if (data.size() < ProtocolConstants::HEADER_SIZE + length) {
        return std::nullopt;
    }

    std::string jsonStr(data.begin() + ProtocolConstants::HEADER_SIZE,
                         data.begin() + ProtocolConstants::HEADER_SIZE + length);

    try {
        json j = json::parse(jsonStr);
        return Message::fromJson(j);
    } catch (const json::parse_error&) {
        return std::nullopt;
    }
}

std::vector<uint8_t> Protocol::makeHeader(size_t length, MessageType type) {
    std::vector<uint8_t> header(ProtocolConstants::HEADER_SIZE, 0);

    // Magic (3 bytes)
    std::memcpy(header.data(), ProtocolConstants::MAGIC, 3);

    // Type (4 bytes)
    std::string typeStr;
    switch (type) {
        case MessageType::Register: typeStr = "reg "; break;
        case MessageType::Heartbeat: typeStr = "hbt "; break;
        case MessageType::Task: typeStr = "tsk "; break;
        case MessageType::TaskResult: typeStr = "tr  "; break;
        case MessageType::Message: typeStr = "msg "; break;
        case MessageType::Broadcast: typeStr = "bct "; break;
        case MessageType::Discovery: typeStr = "dsc "; break;
        case MessageType::Error: typeStr = "err "; break;
        case MessageType::Ack: typeStr = "ack "; break;
    }
    std::memcpy(header.data() + 3, typeStr.c_str(), 4);

    // Length (4 bytes, big endian)
    uint32_t len = static_cast<uint32_t>(length);
    header[7] = (len >> 24) & 0xFF;
    header[8] = (len >> 16) & 0xFF;
    header[9] = (len >> 8) & 0xFF;
    header[10] = len & 0xFF;

    // Version (4 bytes)
    std::memcpy(header.data() + 12, "8.1 ", 4);

    return header;
}

std::pair<MessageType, size_t> Protocol::parseHeader(const uint8_t* header) {
    // 检查Magic
    if (std::memcmp(header, ProtocolConstants::MAGIC, 3) != 0) {
        throw ProtocolException("Invalid magic number");
    }

    // 解析类型
    std::string typeStr(reinterpret_cast<const char*>(header + 3), 4);
    MessageType type = MessageType::Message;
    if (typeStr.find("reg") != std::string::npos) type = MessageType::Register;
    else if (typeStr.find("hbt") != std::string::npos) type = MessageType::Heartbeat;
    else if (typeStr.find("tsk") != std::string::npos) type = MessageType::Task;
    else if (typeStr.find("tr") != std::string::npos) type = MessageType::TaskResult;
    else if (typeStr.find("bct") != std::string::npos) type = MessageType::Broadcast;
    else if (typeStr.find("dsc") != std::string::npos) type = MessageType::Discovery;
    else if (typeStr.find("err") != std::string::npos) type = MessageType::Error;
    else if (typeStr.find("ack") != std::string::npos) type = MessageType::Ack;

    // 解析长度
    size_t length = (static_cast<size_t>(header[7]) << 24) |
                    (static_cast<size_t>(header[8]) << 16) |
                    (static_cast<size_t>(header[9]) << 8) |
                    static_cast<size_t>(header[10]);

    return {type, length};
}

} // namespace ofa