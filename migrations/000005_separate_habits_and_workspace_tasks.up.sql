-- 8. Pisahkan menjadi 2 Tabel Bersih: daily_habits dan workspace_tasks (menggantikan tasks single-table)

-- A. Tabel Daily Habits (Master Kebiasaan & Rutinitas Harian)
CREATE TABLE IF NOT EXISTS daily_habits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'Habit & Routine',
    start_time TEXT,
    end_time TEXT,
    completed_dates TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- B. Tabel Workspace Tasks (Kerjaan Laptop / Target Harian / Child Focus)
CREATE TABLE IF NOT EXISTS workspace_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_id UUID REFERENCES daily_habits(id) ON DELETE SET NULL, -- Relasi Foreign Key ke Daily Habit
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

-- C. Aktifkan Realtime Publication untuk kedua tabel di Supabase
ALTER PUBLICATION supabase_realtime ADD TABLE daily_habits;
ALTER PUBLICATION supabase_realtime ADD TABLE workspace_tasks;
