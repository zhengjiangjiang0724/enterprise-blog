export interface Article {
  id: string;
  title: string;
  slug: string;
  content: string;
  excerpt: string;
  cover_image: string;
  status: string;
  author_id: string;
  view_count: number;
  like_count: number;
  comment_count: number;
  published_at?: string;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResponse<T> {
  code: number;
  message: string;
  data: T[];
  meta?: {
    page: number;
    page_size: number;
    total: number;
    total_page: number;
  };
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}


