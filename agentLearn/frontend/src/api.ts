import axios from 'axios';
import type { Message, Usage } from './types';

const API_BASE_URL = '/api';

export const chatApi = {
  // 普通对话
  async chat(message: string): Promise<{ response: string; usage: Usage; duration: string }> {
    const resp = await axios.post(`${API_BASE_URL}/chat`, { message });
    return resp.data;
  },

  // 流式对话
  async chatStream(message: string, onToken: (token: string) => void): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/chat/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP ${response.status}: ${errorText}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error('ReadableStream not supported');
      }

      const decoder = new TextDecoder();
      let fullContent = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        try {
          const chunk = decoder.decode(value);
          const lines = chunk.split('\n');

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const data = line.slice(6);
              if (data === '[DONE]') {
                continue;
              }
              if (data.startsWith('[ERROR:')) {
                throw new Error(data);
              }
              if (data.trim()) {  // 忽略空数据
                fullContent += data;
                onToken(data);
              }
            }
          }
        } catch (decodeError) {
          console.error('Decode error:', decodeError);
          // 继续处理下一个 chunk
        }
      }
    } catch (error) {
      console.error('ChatStream error:', error);
      if (error instanceof Error && error.message.includes('already finished')) {
        throw new Error('服务器已断开连接，请检查后端服务是否正常');
      }
      throw error;
    }
  },

  // 获取历史消息
  async getHistory(): Promise<Message[]> {
    const resp = await axios.get(`${API_BASE_URL}/history`);
    return resp.data.history;
  },

  // 清空历史
  async clearHistory(): Promise<void> {
    await axios.delete(`${API_BASE_URL}/history`);
  },
};
