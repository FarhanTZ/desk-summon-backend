-- 1. State Realtime (Single-row untuk sinkronisasi remote HP <-> Laptop)
CREATE TABLE IF NOT EXISTS current_session (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    state TEXT NOT NULL DEFAULT 'IDLE',          -- 'IDLE' | 'FOCUSING' | 'SURRENDERED'
    topic TEXT,
    project_path TEXT,
    doc_url TEXT,
    started_at TIMESTAMPTZ,
    target_minutes INT DEFAULT 25,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Inisialisasi baris tunggal jika belum ada
INSERT INTO current_session (id, state) VALUES (1, 'IDLE')
ON CONFLICT (id) DO NOTHING;

-- 2. Log Riwayat Sesi (Dikelola oleh Go Backend)
CREATE TABLE IF NOT EXISTS session_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    actual_minutes INT DEFAULT 0,
    status TEXT NOT NULL,                        -- 'COMPLETED' | 'CANCELLED' | 'SURRENDERED'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Log Intervensi Distraksi (Ditulis oleh Go saat mendeteksi tab/app blacklist)
CREATE TABLE IF NOT EXISTS distraction_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES session_logs(id) ON DELETE CASCADE,
    app_or_url TEXT NOT NULL,
    decision TEXT NOT NULL,                      -- 'BACK_TO_WORK' | 'TOOK_BREAK'
    triggered_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Arsip Diary Harian (Hasil pemrosesan Gemini via Go jam 00:00)
CREATE TABLE IF NOT EXISTS diaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_date DATE UNIQUE NOT NULL,
    total_focus_minutes INT NOT NULL DEFAULT 0,
    completed_sessions INT NOT NULL DEFAULT 0,
    distraction_count INT NOT NULL DEFAULT 0,
    gave_up BOOLEAN NOT NULL DEFAULT FALSE,
    content_markdown TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
