-- ============================================================================
-- SUPABASE DATABASE PERFORMANCE & INDEXING OPTIMIZATION
-- Run this SQL in Supabase SQL Editor to eliminate query lag and sequential scans
-- ============================================================================

-- 1. Index on `diaries` table
CREATE INDEX IF NOT EXISTS idx_diaries_entry_date_desc 
ON public.diaries (entry_date DESC);

-- 2. Index on `daily_habits` table
CREATE INDEX IF NOT EXISTS idx_daily_habits_start_time 
ON public.daily_habits (start_time ASC);

CREATE INDEX IF NOT EXISTS idx_daily_habits_created_at 
ON public.daily_habits (created_at DESC);

-- 3. Index on `workspace_tasks` table
CREATE INDEX IF NOT EXISTS idx_workspace_tasks_status 
ON public.workspace_tasks (status);

CREATE INDEX IF NOT EXISTS idx_workspace_tasks_created_at_desc 
ON public.workspace_tasks (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workspace_tasks_scheduled_date 
ON public.workspace_tasks (scheduled_date);

CREATE INDEX IF NOT EXISTS idx_workspace_tasks_habit_id 
ON public.workspace_tasks (habit_id);

-- 4. Index on `session_logs` table
CREATE INDEX IF NOT EXISTS idx_session_logs_created_at_desc 
ON public.session_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_session_logs_status 
ON public.session_logs (status);

-- 5. Realtime Publication verification
ALTER PUBLICATION supabase_realtime ADD TABLE daily_habits, workspace_tasks, current_session, diaries;
