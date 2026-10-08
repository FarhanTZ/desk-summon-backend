-- 6. Tambah kolom dukungan Daily Routine di tabel tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS is_daily BOOLEAN DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS completed_dates TEXT[] DEFAULT '{}';
