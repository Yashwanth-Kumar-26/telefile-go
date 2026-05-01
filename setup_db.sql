-- TeleFile Master Database Setup

-- 1. Users Table
CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,
    username VARCHAR(255),
    display_name VARCHAR(255),
    storage_limit_gb BIGINT DEFAULT 20,
    storage_used_bytes BIGINT DEFAULT 0,
    referred_by BIGINT,
    is_admin BOOLEAN DEFAULT FALSE,
    is_banned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Batches Table (For Grouped Links)
CREATE TABLE IF NOT EXISTS batches (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    unique_link_id VARCHAR(12) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Files Table
CREATE TABLE IF NOT EXISTS files (
    id VARCHAR(36) PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    file_name VARCHAR(255),
    file_size BIGINT,
    file_type VARCHAR(50),
    telegram_file_id VARCHAR(255),
    unique_link VARCHAR(12) UNIQUE,
    password VARCHAR(255),
    download_limit INTEGER DEFAULT -1,
    download_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    batch_id INTEGER REFERENCES batches(id)
);

-- 4. Gifts Table (Promo Codes)
CREATE TABLE IF NOT EXISTS gifts (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE,
    storage_gb BIGINT,
    max_claims INTEGER DEFAULT 1,
    claim_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Gift Claims Table
CREATE TABLE IF NOT EXISTS gift_claims (
    id SERIAL PRIMARY KEY,
    gift_id INTEGER REFERENCES gifts(id),
    user_id BIGINT REFERENCES users(user_id),
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(gift_id, user_id)
);

-- 6. Subscriptions Table (Plans)
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(user_id),
    plan_name VARCHAR(50),
    storage_gb BIGINT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Referrals Table
CREATE TABLE IF NOT EXISTS referrals (
    id SERIAL PRIMARY KEY,
    referrer_id BIGINT REFERENCES users(user_id),
    referred_id BIGINT REFERENCES users(user_id) UNIQUE,
    bonus_awarded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);