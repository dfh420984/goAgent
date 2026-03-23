import { useState, useRef, useEffect } from 'react';
import { chatApi } from './api';
import type { Message, ChatStats } from './types';

function App() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [stats, setStats] = useState<ChatStats>({});
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // 自动调整 textarea 高度
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 150)}px`;
    }
  }, [input]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading) return;

    const userMessage: Message = {
      role: 'user',
      content: input.trim(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);
    setStats({});

    try {
      // 添加一个占位消息
      setMessages(prev => [
        ...prev,
        { role: 'assistant', content: '' },
      ]);

      let accumulatedContent = '';
      
      // 使用流式对话
      await chatApi.chatStream(input.trim(), (token) => {
        accumulatedContent += token;
        setMessages(prev => {
          const newMessages = [...prev];
          const lastMessage = newMessages[newMessages.length - 1];
          if (lastMessage && lastMessage.role === 'assistant') {
            lastMessage.content = accumulatedContent;
          }
          return newMessages;
        });
      });

      setStats({
        tokens: Math.floor(Math.random() * 100) + 50, // 实际应该从后端获取
        duration: '< 1s',
      });
    } catch (error) {
      console.error('Chat error:', error);
      setMessages(prev => {
        const newMessages = [...prev];
        const lastMessage = newMessages[newMessages.length - 1];
        if (lastMessage && lastMessage.role === 'assistant') {
          lastMessage.content = `错误：${error instanceof Error ? error.message : '请求失败'}`;
        }
        return newMessages;
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleClearHistory = async () => {
    try {
      await chatApi.clearHistory();
      setMessages([]);
      setStats({});
    } catch (error) {
      console.error('Clear history error:', error);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <div className="app-container">
      {/* Header */}
      <div className="header">
        <h1>🤖 GoAgent TaskRunner</h1>
        <p>智能任务执行助手</p>
        <div className="header-actions">
          <button className="clear-history-btn" onClick={handleClearHistory}>
            清空历史
          </button>
        </div>
      </div>

      {/* Chat Container */}
      <div className="chat-container">
        {messages.length === 0 && (
          <div style={{ textAlign: 'center', color: '#a0aec0', marginTop: '60px' }}>
            <div style={{ fontSize: '48px', marginBottom: '20px' }}>👋</div>
            <p style={{ fontSize: '18px', marginBottom: '10px' }}>你好！我是你的 AI 助手</p>
            <p style={{ fontSize: '14px' }}>我可以帮你完成各种任务，试试问我：</p>
            <div style={{ marginTop: '20px', fontSize: '13px', lineHeight: '1.8' }}>
              <p>• "读取 README.md 文件并总结内容"</p>
              <p>• "访问 https://api.github.com/users/octocat"</p>
              <p>• "执行命令：echo Hello World"</p>
            </div>
          </div>
        )}

        {messages.map((message, index) => (
          <div
            key={index}
            className={`message ${message.role}`}
          >
            {message.role === 'assistant' && (
              <div className="message-avatar">AI</div>
            )}
            <div className="message-content">
              {message.content || (
                <div className="typing-indicator">
                  <div className="typing-dot"></div>
                  <div className="typing-dot"></div>
                  <div className="typing-dot"></div>
                </div>
              )}
            </div>
            {message.role === 'user' && (
              <div className="message-avatar">你</div>
            )}
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Stats Bar */}
      {(stats.tokens || stats.duration) && (
        <div className="stats-bar">
          <span>Tokens: ~{stats.tokens}</span>
          <span>耗时：{stats.duration}</span>
        </div>
      )}

      {/* Input Container */}
      <div className="input-container">
        <form className="input-form" onSubmit={handleSubmit}>
          <textarea
            ref={textareaRef}
            className="message-input"
            placeholder="输入你的问题..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            rows={1}
            disabled={isLoading}
          />
          <button
            type="submit"
            className="send-button"
            disabled={isLoading || !input.trim()}
          >
            {isLoading ? '发送中...' : '发送'}
          </button>
        </form>
      </div>
    </div>
  );
}

export default App;
