-- 回滚全文搜索支持
DROP TRIGGER IF EXISTS articles_search_vector_trigger ON articles;
DROP FUNCTION IF EXISTS articles_search_vector_update();
DROP INDEX IF EXISTS idx_articles_search_vector;
ALTER TABLE articles DROP COLUMN IF EXISTS search_vector;

