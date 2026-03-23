export interface Message {
  role: 'user' | 'assistant';
  content: string;
}

export interface Usage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

export interface ChatStats {
  tokens?: number;
  duration?: string;
  toolCalls?: number;
}
