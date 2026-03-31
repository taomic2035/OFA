/**
 * @file protocol.hpp
 * @brief Protocol Module
 * Sprint 29: C++ Agent SDK
 */

#ifndef OFA_PROTOCOL_HPP
#define OFA_PROTOCOL_HPP

#include <vector>
#include <cstdint>
#include "message.hpp"

namespace ofa {

/**
 * @brief 协议常量
 */
struct ProtocolConstants {
    static constexpr const char* VERSION = "8.1.0";
    static constexpr const char* MAGIC = "OFA";
    static constexpr size_t HEADER_SIZE = 16;
};

/**
 * @brief 协议编解码
 */
class Protocol {
public:
    /**
     * @brief 编码消息
     */
    static std::vector<uint8_t> encode(const Message& msg);

    /**
     * @brief 解码消息
     */
    static std::optional<Message> decode(const std::vector<uint8_t>& data);

private:
    static std::vector<uint8_t> makeHeader(size_t length, MessageType type);
    static std::pair<MessageType, size_t> parseHeader(const uint8_t* header);
};

} // namespace ofa

#endif // OFA_PROTOCOL_HPP