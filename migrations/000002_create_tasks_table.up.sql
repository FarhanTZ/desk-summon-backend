-- 5. Tabel Tasks (Dikelola oleh Flutter Mobile App)
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    categories TEXT[] DEFAULT '{}',
    urls TEXT[] DEFAULT '{}',
    open_vscode BOOLEAN DEFAULT FALSE,
    scheduled_date DATE,
    start_time TEXT,
    end_time TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Aktifkan realtime stream untuk tabel tasks di Supabase
ALTER PUBLICATION supabase_realtime ADD TABLE tasks;
