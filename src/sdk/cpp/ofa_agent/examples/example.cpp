/**
 * @file example.cpp
 * @brief OFA C++ Agent Example
 * Sprint 29: C++ Agent SDK
 */

#include <iostream>
#include "ofa/agent.hpp"
#include "ofa/builtin.hpp"

int main() {
    std::cout << "OFA C++ Agent SDK v" << ofa::Version::STRING << std::endl;

    // 创建配置
    auto config = ofa::AgentConfigBuilder()
        .agentId("cpp-demo-001")
        .name("C++ Demo Agent")
        .centerUrl("localhost:9090")
        .connectionType("grpc")
        .heartbeatInterval(30)
        .addSkill("echo")
        .addSkill("text.process")
        .addSkill("calculator")
        .build();

    // 创建Agent
    ofa::Agent agent(config);

    // 注册内置技能
    // registerBuiltinSkills需要在skillExecutor上调用

    std::cout << "Agent ID: " << agent.id() << std::endl;
    std::cout << "Agent Name: " << agent.info().name << std::endl;

    // 连接到Center
    // auto connected = agent.connect().get();
    // if (connected) {
    //     std::cout << "Connected successfully" << std::endl;
    // }

    // 运行Agent
    // agent.run().get();

    // 断开连接
    // agent.disconnect().get();

    std::cout << "Demo completed" << std::endl;
    return 0;
}