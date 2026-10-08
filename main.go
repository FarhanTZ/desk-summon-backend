package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/supabase-community/supabase-go"
	"google.golang.org/api/option"
)

type CurrentSession struct {
	ID            int        `json:"id"`
	State         string     `json:"state"`
	Topic         *string    `json:"topic"`
	ProjectPath   *string    `json:"project_path"`
	DocURL        *string    `json:"doc_url"`
	StartedAt     *time.Time `json:"started_at"`
	TargetMinutes int        `json:"target_minutes"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type SessionLog struct {
	Topic         string    `json:"topic"`
	StartedAt     time.Time `json:"started_at"`
	EndedAt       time.Time `json:"ended_at"`
	ActualMinutes int       `json:"actual_minutes"`
	Status        string    `json:"status"` // COMPLETED | CANCELLED | SURRENDERED
}

type DailyHabit struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Category       string   `json:"category"`
	StartTime      *string  `json:"start_time"`
	EndTime        *string  `json:"end_time"`
	CompletedDates []string `json:"completed_dates"`
	RepeatDays     []int    `json:"repeat_days"`
}

type WorkspaceTask struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Status        string  `json:"status"`
	ScheduledDate *string `json:"scheduled_date"`
}

type DiaryEntry struct {
	EntryDate         string `json:"entry_date"`
	TotalFocusMinutes int    `json:"total_focus_minutes"`
	CompletedSessions int    `json:"completed_sessions"`
	DistractionCount  int    `json:"distraction_count"`
	GaveUp            bool   `json:"gave_up"`
	ContentMarkdown   string `json:"content_markdown"`
}

var (
	supabaseClient          *supabase.Client
	currentSessionStartTime time.Time
	currentSessionTopic     string
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca environment bawaan")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if supabaseURL == "" || supabaseKey == "" || geminiKey == "" {
		log.Fatal("Error: Pastikan SUPABASE_URL, SUPABASE_ANON_KEY, dan GEMINI_API_KEY sudah terisi di .env")
	}

	// 2. Jalankan Auto-Migrations jika DATABASE_URL ada
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		if err := RunMigrations(dbURL); err != nil {
			log.Printf("⚠️ Peringatan Migrasi: %v", err)
		}
	}

	// 3. Inisialisasi Supabase Client
	var err error
	supabaseClient, err = supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		log.Fatalf("Gagal inisialisasi Supabase: %v", err)
	}

	// 4. Setup Cron Job Jam 00:00:00 (Tengah Malam)
	c := cron.New()
	_, err = c.AddFunc("0 0 * * *", func() {
		fmt.Println("\n⏰ Jam 12 malam terdeteksi! Memulai penyusunan Auto-Diary komprehensif via Gemini Flash...")
		runMidnightDiaryPipeline()
	})
	if err != nil {
		log.Fatalf("Gagal setup cron job: %v", err)
	}
	c.Start()

	fmt.Println("🚀 Desk Summon Daemon Aktif!")
	fmt.Println("📅 Scheduler jam 00:00:00 siap berjalan di background.")
	fmt.Println("🤖 Integrasi Gemini Flash API Aktif (Habits + Workspace Tasks + Sesi Fokus).")

	// Uji coba langsung pipeline diary saat daemon pertama kali menyala (Instant Testing)
	fmt.Println("\n🧪 Menjalankan uji coba pipeline Auto-Diary komprehensif...")
	runMidnightDiaryPipeline()

	fmt.Println("\n📡 Menunggu perintah fokus dari Flutter Mobile (current_session)...")

	// 5. Loop Polling State
	var lastHandledState string = "IDLE"

	for {
		data, _, err := supabaseClient.From("current_session").Select("*", "exact", false).Eq("id", "1").Single().Execute()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		var session CurrentSession
		if err := json.Unmarshal(data, &session); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		// A. Transisi ke FOCUSING
		if session.State == "FOCUSING" && lastHandledState != "FOCUSING" {
			currentSessionStartTime = time.Now()
			currentSessionTopic = ptrToString(session.Topic)

			fmt.Printf("\n⚡ Sesi Belajar Dimulai: %s\n", currentSessionTopic)

			targetPath := "."
			if session.ProjectPath != nil && *session.ProjectPath != "" {
				targetPath = *session.ProjectPath
			}

			_ = openVSCode(targetPath)

			// Buka browser URLs (support multiple URLs dipisah koma)
			if session.DocURL != nil && *session.DocURL != "" {
				rawURLs := strings.Split(*session.DocURL, ",")
				for _, u := range rawURLs {
					cleanURL := strings.TrimSpace(u)
					if cleanURL != "" {
						_ = openBrowser(cleanURL)
						time.Sleep(300 * time.Millisecond)
					}
				}
			}

			lastHandledState = "FOCUSING"
		}

		// B. Transisi ke SURRENDERED (End Focus Session dari HP)
		if session.State == "SURRENDERED" && lastHandledState == "FOCUSING" {
			fmt.Println("🛑 Protokol End Focus Session diterima dari HP.")
			duration := int(time.Since(currentSessionStartTime).Minutes())
			if duration < 1 {
				duration = 1
			}
			logSession(currentSessionTopic, currentSessionStartTime, time.Now(), duration, "SURRENDERED")
			lastHandledState = "SURRENDERED"
		}

		// C. Transisi kembali ke IDLE (Selesai Normal / Complete)
		if session.State == "IDLE" && lastHandledState == "FOCUSING" {
			fmt.Println("✅ Sesi fokus selesai normal.")
			duration := int(time.Since(currentSessionStartTime).Minutes())
			if duration < 1 {
				duration = 1
			}
			logSession(currentSessionTopic, currentSessionStartTime, time.Now(), duration, "COMPLETED")
			lastHandledState = "IDLE"
		}

		time.Sleep(2 * time.Second)
	}
}

// Simpan riwayat sesi ke tabel session_logs
func logSession(topic string, start time.Time, end time.Time, minutes int, status string) {
	record := SessionLog{
		Topic:         topic,
		StartedAt:     start,
		EndedAt:       end,
		ActualMinutes: minutes,
		Status:        status,
	}

	_, _, err := supabaseClient.From("session_logs").Insert(record, false, "", "", "").Execute()
	if err != nil {
		log.Printf("Gagal mencatat log sesi ke Supabase: %v", err)
	} else {
		fmt.Printf("📝 Riwayat sesi tersimpan: %s (%d menit, Status: %s)\n", topic, minutes, status)
	}
}

// Pipeline Midnight Diary Komprehensif (Habit + Tasks + Sesi)
func runMidnightDiaryPipeline() {
	ctx := context.Background()
	now := time.Now()
	today := now.Format("2006-01-02")

	// ISO Weekday: Monday = 1, ..., Sunday = 7
	isoWeekday := int(now.Weekday())
	if isoWeekday == 0 {
		isoWeekday = 7
	}

	// 1. Ambil seluruh Daily Habits & Rutinitas
	habitsData, _, err := supabaseClient.From("daily_habits").Select("*", "exact", false).Execute()
	var allHabits []DailyHabit
	if err == nil && len(habitsData) > 0 {
		_ = json.Unmarshal(habitsData, &allHabits)
	}

	var scheduledHabits []DailyHabit
	var completedHabits []DailyHabit
	var missedHabits []DailyHabit

	for _, h := range allHabits {
		// Cek apakah habit dijadwalkan untuk hari ini
		isScheduledToday := false
		if len(h.RepeatDays) == 0 {
			isScheduledToday = true
		} else {
			for _, d := range h.RepeatDays {
				if d == isoWeekday {
					isScheduledToday = true
					break
				}
			}
		}

		if !isScheduledToday {
			continue
		}

		scheduledHabits = append(scheduledHabits, h)

		// Cek apakah selesai pada tanggal hari ini
		isDone := false
		for _, dt := range h.CompletedDates {
			if dt == today {
				isDone = true
				break
			}
		}

		if isDone {
			completedHabits = append(completedHabits, h)
		} else {
			missedHabits = append(missedHabits, h)
		}
	}

	// 2. Ambil Workspace Tasks hari ini
	tasksData, _, err := supabaseClient.From("workspace_tasks").Select("*", "exact", false).Eq("scheduled_date", today).Execute()
	var allTasks []WorkspaceTask
	if err == nil && len(tasksData) > 0 {
		_ = json.Unmarshal(tasksData, &allTasks)
	}

	var doneTasks []WorkspaceTask
	var pendingTasks []WorkspaceTask
	for _, t := range allTasks {
		if t.Status == "done" {
			doneTasks = append(doneTasks, t)
		} else {
			pendingTasks = append(pendingTasks, t)
		}
	}

	// 3. Ambil log sesi deep work hari ini dari database
	logsData, _, err := supabaseClient.From("session_logs").Select("*", "exact", false).Gte("started_at", today+"T00:00:00Z").Execute()
	if err != nil {
		log.Printf("⚠️ Gagal membaca session_logs: %v", err)
	}

	var logs []SessionLog
	if len(logsData) > 0 {
		_ = json.Unmarshal(logsData, &logs)
	}

	totalFocusMinutes := 0
	completedSessions := 0
	gaveUp := false

	for _, l := range logs {
		totalFocusMinutes += l.ActualMinutes
		if l.Status == "COMPLETED" {
			completedSessions++
		}
		if l.Status == "SURRENDERED" {
			gaveUp = true
		}
	}

	// 4. Susun ringkasan metrik komprehensif untuk prompt AI
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== REKAP AKTIVITAS LENGKAP PENGGUNA HARI INI (%s) ===\n\n", today))

	// Section A: Rutinitas & Habits
	sb.WriteString(fmt.Sprintf("📋 [1. DAILY HABITS & RUTINITAS HARIAN]\n"))
	sb.WriteString(fmt.Sprintf("- Total Dijadwalkan: %d | Selesai Dikerjakan: %d | Terlewat/Belum: %d\n", len(scheduledHabits), len(completedHabits), len(missedHabits)))
	if len(completedHabits) > 0 {
		sb.WriteString("  ✅ Habit Berhasil Diselesaikan:\n")
		for _, h := range completedHabits {
			timeStr := ""
			if h.StartTime != nil {
				timeStr = fmt.Sprintf(" (%s - %s)", ptrToString(h.StartTime), ptrToString(h.EndTime))
			}
			sb.WriteString(fmt.Sprintf("    * %s [%s]%s\n", h.Title, h.Category, timeStr))
		}
	}
	if len(missedHabits) > 0 {
		sb.WriteString("  ❌ Habit yang DILEWATI / TIDAK DIKERJAKAN:\n")
		for _, h := range missedHabits {
			timeStr := ""
			if h.StartTime != nil {
				timeStr = fmt.Sprintf(" (%s - %s)", ptrToString(h.StartTime), ptrToString(h.EndTime))
			}
			sb.WriteString(fmt.Sprintf("    * %s [%s]%s\n", h.Title, h.Category, timeStr))
		}
	}
	if len(scheduledHabits) == 0 {
		sb.WriteString("  (Tidak ada rutinitas khusus yang dijadwalkan hari ini)\n")
	}

	// Section B: Workspace Tasks
	sb.WriteString(fmt.Sprintf("\n💻 [2. WORKSPACE / TARGET KERJA LAPTOP HARI INI]\n"))
	sb.WriteString(fmt.Sprintf("- Total Task: %d | Tuntas: %d | Tertunda: %d\n", len(allTasks), len(doneTasks), len(pendingTasks)))
	for _, t := range doneTasks {
		sb.WriteString(fmt.Sprintf("  * [SELESAI] %s\n", t.Title))
	}
	for _, t := range pendingTasks {
		sb.WriteString(fmt.Sprintf("  * [TERTUNDA] %s (Status: %s)\n", t.Title, t.Status))
	}
	if len(allTasks) == 0 {
		sb.WriteString("  (Tidak ada target task spesifik di luar rutinitas)\n")
	}

	// Section C: Deep Work Session Logs
	sb.WriteString(fmt.Sprintf("\n⏱️ [3. SESI DEEP WORK LAPTOP]\n"))
	sb.WriteString(fmt.Sprintf("- Total Durasi Fokus: %d menit | Sesi Tuntas: %d | Sesi Batal/Menyerah: %t\n", totalFocusMinutes, completedSessions, gaveUp))
	for _, l := range logs {
		sb.WriteString(fmt.Sprintf("  * Topik: %s | Durasi: %d menit | Status: %s\n", l.Topic, l.ActualMinutes, l.Status))
	}
	if len(logs) == 0 {
		sb.WriteString("  (Belum ada sesi laptop summon tercatat)\n")
	}

	metricsText := sb.String()

	// 5. Request evaluasi ke Gemini Flash API
	apiKey := os.Getenv("GEMINI_API_KEY")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Gagal inisialisasi Gemini Client: %v", err)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-3.5-flash")

	prompt := fmt.Sprintf(`Bertindaklah sebagai AI Accountability Partner dan evaluator produktivitas yang objektif, jujur, tegas, suportif namun anti-basa-basi.
Evaluasilah performa harian pengguna berdasarkan data lengkap berikut:
%s

Tolong tulis evaluasi jurnal harian komprehensif dalam format Markdown dengan struktur:
# %s — Daily Reality & Accountability
### 1. The Reality (Fakta Angka & Perbandingan)
- Paparkan perbandingan nyata antara habit yang berhasil diselesaikan vs habit yang terlewat / tidak dikerjakan.
- Sebutkan angka durasi deep work dan status penyelesaian task laptop.

### 2. Honest Diagnosis (Penyebab & Pola Prokrastinasi)
- Soroti secara spesifik habit atau task mana yang diabaikan/terlewat (misal: olahraga, belajar, atau target kerja tertentu).
- Berikan diagnosis jujur mengapa habit tersebut dilewati (misal: friksi awal, manajemen energi, atau distraksi).

### 3. Tomorrow's Single Target (Misi 5 Menit Penyelamat)
- Tentukan TEPAT 1 aksi mikro berdurasi 5 menit pertama untuk besok pagi guna mengembalikan momentum habit yang sempat terlewat.`, metricsText, today)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		// Fallback to gemini-3.5-flash-lite or gemini-3.1-flash-lite if needed
		modelFallback := client.GenerativeModel("gemini-3.5-flash-lite")
		resp, err = modelFallback.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			modelFallback2 := client.GenerativeModel("gemini-3.1-flash-lite")
			resp, err = modelFallback2.GenerateContent(ctx, genai.Text(prompt))
			if err != nil {
				log.Printf("❌ Gagal generate diary via Gemini: %v", err)
				return
			}
		}
	}

	var diaryMarkdown string
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				diaryMarkdown += string(textPart)
			} else {
				diaryMarkdown += fmt.Sprintf("%v", part)
			}
		}
	}

	if diaryMarkdown == "" {
		log.Println("⚠️ Hasil evaluasi Gemini kosong")
		return
	}

	// 6. Simpan langsung ke database Supabase tabel `diaries`
	entry := DiaryEntry{
		EntryDate:         today,
		TotalFocusMinutes: totalFocusMinutes,
		CompletedSessions: completedSessions,
		DistractionCount:  len(missedHabits), // Catat jumlah habit yang terlewat sebagai metrik distraksi/loss
		GaveUp:            gaveUp,
		ContentMarkdown:   diaryMarkdown,
	}

	_, _, err = supabaseClient.From("diaries").Upsert(entry, "entry_date", "", "").Execute()
	if err != nil {
		log.Printf("Gagal simpan diary ke Supabase: %v", err)
	} else {
		fmt.Printf("☁️ Diary Komprehensif (%s) berhasil disimpan ke database Supabase!\n", today)
	}
}

func openVSCode(path string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "code", path)
	} else {
		cmd = exec.Command("code", path)
	}
	return cmd.Start()
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func ptrToString(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}