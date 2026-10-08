-- 9. Tambahkan kolom repeat_days ke daily_habits untuk filtering hari aktif
ALTER TABLE daily_habits ADD COLUMN IF NOT EXISTS repeat_days INT[] DEFAULT '{1,2,3,4,5,6,7}';
