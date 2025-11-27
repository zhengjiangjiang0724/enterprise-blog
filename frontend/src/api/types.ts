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

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
}

export interface Tag {
  id: string;
  name: string;
  slug: string;
  color?: string;
  description?: string;
}

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
  category?: Category | null;
  tags?: Tag[];
}

export interface UserProfile {
  id: string;
  username: string;
  email: string;
  role: string;
  avatar: string;
  bio: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ArticlePayload {
  title: string;
  content: string;
  excerpt?: string;
  cover_image?: string;
  status?: string;
  category_id?: string | null;
  tag_ids?: string[];
}

export interface Comment {
  id: string;
  article_id: string;
  user_id?: string | null;
  parent_id?: string | null;
  content: string;
  author: string;
  email?: string;
  website?: string;
  created_at: string;
}

export interface CommentPayload {
  parent_id?: string | null;
  content: string;
  author?: string;
  email?: string;
  website?: string;
  article_id: string;
}



