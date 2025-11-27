interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  // 在此处按需扩展更多 VITE_ 前缀的环境变量
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}


