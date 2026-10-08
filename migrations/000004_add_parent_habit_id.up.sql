-- 7. Tambah kolom parent_habit_id di tabel tasks untuk relasi 2-layer
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS parent_habit_id UUID REFERENCES tasks(id) ON DELETE SET NULL;
