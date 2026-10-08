package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

type HabitSeed struct {
	Title      string
	Category   string
	StartTime  string
	EndTime    string
	RepeatDays []int64
}

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		_ = godotenv.Load(".env")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	fmt.Println("📦 Connecting directly to Postgres...")
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	// 1. Ensure column repeat_days
	_, err = db.Exec("ALTER TABLE daily_habits ADD COLUMN IF NOT EXISTS repeat_days INT[] DEFAULT '{1,2,3,4,5,6,7}';")
	if err != nil {
		log.Printf("Alter table warning: %v", err)
	}

	// 2. Reload PostgREST schema cache
	_, _ = db.Exec("NOTIFY pgrst, 'reload schema';")

	// 1 = Senin, 2 = Selasa, 3 = Rabu, 4 = Kamis, 5 = Jumat, 6 = Sabtu, 7 = Minggu
	allDays := []int64{1, 2, 3, 4, 5, 6, 7}
	monWedFri := []int64{1, 3, 5}
	tueThu := []int64{2, 4}
	tueThuSat := []int64{2, 4, 6}
	monToSat := []int64{1, 2, 3, 4, 5, 6}
	satOnly := []int64{6}
	sunOnly := []int64{7}
	monToFriAndSun := []int64{1, 2, 3, 4, 5, 7}

	habits := []HabitSeed{
		// 1. Pagi Rutin (Semua Hari)
		{
			Title:      "Bangun & Sholat Subuh",
			Category:   "Spiritual & Wellness",
			StartTime:  "04:30",
			EndTime:    "05:15",
			RepeatDays: allDays,
		},
		{
			Title:      "Olahraga di Rumah",
			Category:   "Health & Fitness",
			StartTime:  "05:15",
			EndTime:    "06:00",
			RepeatDays: allDays,
		},
		{
			Title:      "Pekerjaan Rumah & Rapikan Kamar Mandi",
			Category:   "Life & Chores",
			StartTime:  "06:00",
			EndTime:    "07:15",
			RepeatDays: allDays,
		},
		{
			Title:      "Mandi & Sarapan Pagi",
			Category:   "Health & Wellness",
			StartTime:  "07:15",
			EndTime:    "08:00",
			RepeatDays: allDays,
		},

		// 2. Sesi Pagi 08:00 - 11:45
		{
			Title:      "Belajar IELTS & TOEFL",
			Category:   "Learning & Course",
			StartTime:  "08:00",
			EndTime:    "11:45",
			RepeatDays: monWedFri,
		},
		{
			Title:      "Belajar Tes Logika & Psikotes",
			Category:   "Learning & Course",
			StartTime:  "08:00",
			EndTime:    "11:45",
			RepeatDays: tueThuSat,
		},
		{
			Title:      "Evaluasi Mingguan / Santai",
			Category:   "Personal Review",
			StartTime:  "08:00",
			EndTime:    "11:45",
			RepeatDays: sunOnly,
		},

		// 3. Dzuhur & Siang
		{
			Title:      "Sholat Dzuhur & Makan Siang",
			Category:   "Spiritual & Wellness",
			StartTime:  "11:45",
			EndTime:    "13:00",
			RepeatDays: allDays,
		},
		{
			Title:      "Power Nap / Istirahat Siang",
			Category:   "Health & Rest",
			StartTime:  "13:00",
			EndTime:    "13:30",
			RepeatDays: monToSat,
		},
		{
			Title:      "Free Time / Jalan-jalan Santai",
			Category:   "Leisure & Refreshing",
			StartTime:  "13:00",
			EndTime:    "18:00",
			RepeatDays: sunOnly,
		},

		// 4. Sesi Siang-Sore 13:30 - 15:15
		{
			Title:      "Ngoding & Build Project (Sesi 1)",
			Category:   "Coding & Project",
			StartTime:  "13:30",
			EndTime:    "15:15",
			RepeatDays: monWedFri,
		},
		{
			Title:      "Belajar Public Speaking & Hal Baru",
			Category:   "Personal Development",
			StartTime:  "13:30",
			EndTime:    "15:00",
			RepeatDays: tueThuSat,
		},

		// 5. Ashar 15:00 - 15:30
		{
			Title:      "Sholat Ashar & Break Sore",
			Category:   "Spiritual & Wellness",
			StartTime:  "15:00",
			EndTime:    "15:30",
			RepeatDays: monToSat,
		},

		// 6. Sesi Sore 15:30 - 17:00
		{
			Title:      "Ngoding & Build Project (Sesi 2)",
			Category:   "Coding & Project",
			StartTime:  "15:30",
			EndTime:    "16:45",
			RepeatDays: monWedFri,
		},
		{
			Title:      "Santai / Lanjutan Tugas Ringan",
			Category:   "Deep Focus",
			StartTime:  "15:30",
			EndTime:    "17:00",
			RepeatDays: tueThu,
		},
		{
			Title:      "Main / Nongkrong / Jalan ke Cafe",
			Category:   "Social & Leisure",
			StartTime:  "15:30",
			EndTime:    "21:30",
			RepeatDays: satOnly,
		},

		// 7. Senja & Malam (Non-Sabtu)
		{
			Title:      "Beres Rumah Ringan & Mandi Sore",
			Category:   "Life & Chores",
			StartTime:  "17:00",
			EndTime:    "18:00",
			RepeatDays: monToFriAndSun,
		},
		{
			Title:      "Sholat Maghrib, Makan Malam, Sholat Isya",
			Category:   "Spiritual & Wellness",
			StartTime:  "18:00",
			EndTime:    "19:30",
			RepeatDays: monToFriAndSun,
		},
		{
			Title:      "Review Ringan & Baca Santai",
			Category:   "Reading & Review",
			StartTime:  "19:30",
			EndTime:    "20:30",
			RepeatDays: monToFriAndSun,
		},
		{
			Title:      "Free Time / Persiapan Tidur",
			Category:   "Rest & Relaxation",
			StartTime:  "20:30",
			EndTime:    "22:00",
			RepeatDays: monToFriAndSun,
		},

		// 8. Khusus Sabtu Malam
		{
			Title:      "Pulang, Bersih-bersih & Istirahat",
			Category:   "Rest & Relaxation",
			StartTime:  "21:30",
			EndTime:    "22:00",
			RepeatDays: satOnly,
		},

		// 9. Tidur Malam (Semua Hari)
		{
			Title:      "Tidur Malam",
			Category:   "Health & Sleep",
			StartTime:  "22:00",
			EndTime:    "04:30",
			RepeatDays: allDays,
		},
	}

	// 1. Delete all old daily habits
	fmt.Println("🧹 Cleaning old dummy daily_habits...")
	_, err = db.Exec("DELETE FROM daily_habits;")
	if err != nil {
		log.Fatalf("Delete error: %v", err)
	}
	fmt.Println("✅ Old dummy habits deleted.")

	// 2. Insert all real habits
	fmt.Printf("🌱 Inserting %d real weekly habits...\n", len(habits))
	stmt, err := db.Prepare("INSERT INTO daily_habits (title, category, start_time, end_time, repeat_days, completed_dates) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Fatalf("Prepare error: %v", err)
	}
	defer stmt.Close()

	for _, h := range habits {
		_, err := stmt.Exec(h.Title, h.Category, h.StartTime, h.EndTime, pq.Array(h.RepeatDays), pq.Array([]string{}))
		if err != nil {
			log.Fatalf("Insert habit '%s' error: %v", h.Title, err)
		}
	}

	fmt.Println("🎉 All 22 real weekly habits successfully seeded to Supabase!")
}
