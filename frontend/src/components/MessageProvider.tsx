import {
  createContext,
  useCallback,
  useContext,
  useState,
  type ReactNode
} from "react";

type MessageType = "success" | "error" | "info";

interface MessageState {
  type: MessageType;
  text: string;
}

interface MessageContextValue {
  showMessage: (type: MessageType, text: string) => void;
  showSuccess: (text: string) => void;
  showError: (text: string) => void;
}

const MessageContext = createContext<MessageContextValue | null>(null);

export function MessageProvider({ children }: { children: ReactNode }) {
  const [message, setMessage] = useState<MessageState | null>(null);

  const clear = useCallback(() => setMessage(null), []);

  const showMessage = useCallback((type: MessageType, text: string) => {
    setMessage({ type, text });
    // 自动 3 秒后消失
    setTimeout(() => {
      clear();
    }, 3000);
  }, [clear]);

  const showSuccess = useCallback(
    (text: string) => showMessage("success", text),
    [showMessage]
  );

  const showError = useCallback(
    (text: string) => showMessage("error", text),
    [showMessage]
  );

  const value: MessageContextValue = {
    showMessage,
    showSuccess,
    showError
  };

  return (
    <MessageContext.Provider value={value}>
      {children}
      {message && (
        <div className={`message-bar message-${message.type}`}>
          <span>{message.text}</span>
        </div>
      )}
    </MessageContext.Provider>
  );
}

export function useMessage() {
  const ctx = useContext(MessageContext);
  if (!ctx) {
    throw new Error("useMessage must be used within a MessageProvider");
  }
  return ctx;
}


