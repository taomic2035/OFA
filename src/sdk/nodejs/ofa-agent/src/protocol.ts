/**
 * Protocol Module
 * Sprint 29: Node.js Agent SDK
 */

import { Message, MessageType } from './message';

const PROTOCOL_VERSION = '8.1.0';
const MAGIC = 'OFA';
const HEADER_SIZE = 16;

/**
 * 协议编码/解码
 */
export class Protocol {
  /**
   * 编码消息
   */
  static encode(msg: Message): Buffer {
    const jsonStr = JSON.stringify(msg);
    const jsonBuf = Buffer.from(jsonStr, 'utf-8');
    const header = this.makeHeader(jsonBuf.length, msg.type);
    return Buffer.concat([header, jsonBuf]);
  }

  /**
   * 解码消息
   */
  static decode(data: Buffer): Message | null {
    if (data.length < HEADER_SIZE) {
      return null;
    }

    const header = data.subarray(0, HEADER_SIZE);
    const body = data.subarray(HEADER_SIZE);

    const { type, length } = this.parseHeader(header);
    if (body.length !== length) {
      return null;
    }

    try {
      return JSON.parse(body.toString('utf-8'));
    } catch {
      return null;
    }
  }

  private static makeHeader(length: number, type: MessageType): Buffer {
    const header = Buffer.alloc(HEADER_SIZE);

    // Magic (3 bytes)
    header.write(MAGIC, 0, 'ascii');

    // Type (4 bytes)
    const typeStr = this.typeToString(type).padEnd(4, ' ').slice(0, 4);
    header.write(typeStr, 3, 'ascii');

    // Length (4 bytes)
    header.writeUInt32BE(length, 7);

    // Version (4 bytes)
    header.write('8.1 ', 12, 'ascii');

    return header;
  }

  private static parseHeader(header: Buffer): { type: MessageType; length: number } {
    const magic = header.toString('ascii', 0, 3);
    if (magic !== MAGIC) {
      throw new Error('Invalid magic number');
    }

    const typeStr = header.toString('ascii', 3, 7).trim();
    const type = this.stringToType(typeStr);
    const length = header.readUInt32BE(7);

    return { type, length };
  }

  private static typeToString(type: MessageType): string {
    switch (type) {
      case MessageType.REGISTER: return 'reg';
      case MessageType.HEARTBEAT: return 'hbt';
      case MessageType.TASK: return 'tsk';
      case MessageType.TASK_RESULT: return 'tr';
      case MessageType.MESSAGE: return 'msg';
      case MessageType.BROADCAST: return 'bct';
      case MessageType.DISCOVERY: return 'dsc';
      case MessageType.ERROR: return 'err';
      case MessageType.ACK: return 'ack';
      default: return 'unk';
    }
  }

  private static stringToType(str: string): MessageType {
    switch (str) {
      case 'reg': return MessageType.REGISTER;
      case 'hbt': return MessageType.HEARTBEAT;
      case 'tsk': return MessageType.TASK;
      case 'tr': return MessageType.TASK_RESULT;
      case 'msg': return MessageType.MESSAGE;
      case 'bct': return MessageType.BROADCAST;
      case 'dsc': return MessageType.DISCOVERY;
      case 'err': return MessageType.ERROR;
      case 'ack': return MessageType.ACK;
      default: return MessageType.MESSAGE;
    }
  }
}