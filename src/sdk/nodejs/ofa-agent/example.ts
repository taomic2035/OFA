/**
 * OFA Node.js Agent Example
 * Sprint 29: Node.js Agent SDK
 */

import { OFAAgent, AgentConfig, registerBuiltinSkills, MessageType } from '../src';

async function main() {
  // 创建配置
  const config: Partial<AgentConfig> = {
    agentId: 'nodejs-demo-001',
    name: 'Node.js Demo Agent',
    centerUrl: 'localhost:9090',
    connectionType: 'http',
    skills: ['echo', 'text.process', 'calculator', 'json.process'],
  };

  // 创建Agent
  const agent = new OFAAgent(config);

  // 注册内置技能
  registerBuiltinSkills({
    register: (id, handler) => agent.registerSkill(id, handler),
  });

  console.log(`Agent ID: ${agent.id}`);
  console.log(`Agent Name: ${agent.info.name}`);

  // 连接到Center
  // await agent.connect();

  // 注册消息处理器
  agent.onMessage('custom', (msg) => {
    console.log('Received custom message:', msg);
  });

  // 发送消息
  // await agent.sendMessage('target-agent', MessageType.MESSAGE, { hello: 'world' });

  // 运行Agent
  // await agent.run();

  console.log('Demo completed');

  // 断开连接
  // await agent.disconnect();
}

main().catch(console.error);