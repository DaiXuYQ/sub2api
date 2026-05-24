-- 每日签到奖励记录
-- 用独立表强制每个用户每天只能签到一次，并通过 redeem_code_id 串联余额发放流水。
CREATE TABLE IF NOT EXISTS daily_checkins (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    checkin_date    DATE NOT NULL,
    reward_amount   DECIMAL(20, 8) NOT NULL,
    redeem_code_id  BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, checkin_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_checkins_user_id ON daily_checkins(user_id);
CREATE INDEX IF NOT EXISTS idx_daily_checkins_checkin_date ON daily_checkins(checkin_date);
