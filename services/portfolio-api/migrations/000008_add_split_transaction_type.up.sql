-- 000008_add_split_transaction_type.up.sql
ALTER TYPE portfolio.transaction_type ADD VALUE IF NOT EXISTS 'SPLIT';
