INSERT INTO apps (id, name)
VALUES (1, 'test')
    ON CONFLICT DO NOTHING;