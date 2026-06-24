INSERT INTO wallets (id, balance)
    VALUES ('5519fa56-30ba-416f-a7f8-1e60ea44e4d2', 0) ON CONFLICT (id) DO NOTHING;