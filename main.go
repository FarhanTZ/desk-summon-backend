package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
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

func main() {
	// 1. Load file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca environment bawaan")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("Error: SUPABASE_URL atau SUPABASE_ANON_KEY belum diisi di .env")
	}

	// 2. Inisialisasi Supabase Client
	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		log.Fatalf("Gagal inisialisasi Supabase: %v", err)
	}

	fmt.Println("🚀 Daemon Desk Summon aktif...")
	fmt.Println("Menunggu perintah dari HP (current_session)...")

	var lastHandledState string = ""
	var lastHandledTopic string = ""

	// 3. Polling loop (cek setiap 2 detik)
	for {
		data, _, err := client.From("current_session").Select("*", "exact", false).Eq("id", "1").Single().Execute()
		if err != nil {
			log.Printf("Gagal membaca database: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		var session CurrentSession
		if err := json.Unmarshal(data, &session); err != nil {
			log.Printf("Gagal parse data session: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		currentTopic := ptrToString(session.Topic)

		// Deteksi jika state FOCUSING (baik baru berubah ke FOCUSING, atau topik baru dikirim saat FOCUSING)
		if session.State == "FOCUSING" {
			if lastHandledState != "FOCUSING" || (currentTopic != "-" && currentTopic != lastHandledTopic) {
				fmt.Printf("\n⚡ Perintah diterima! Topik: %s\n", currentTopic)

				// Buka VS Code hanya jika ada project path
				if session.ProjectPath != nil && *session.ProjectPath != "" {
					targetPath := *session.ProjectPath
					if err := openVSCode(targetPath); err != nil {
						log.Printf("Gagal membuka VS Code: %v", err)
					} else {
						fmt.Printf("✅ VS Code berhasil dibuka pada: %s\n", targetPath)
					}
				}

				// Buka URL / Browser jika ada (bisa multiple URL dipisahkan koma)
				if session.DocURL != nil && *session.DocURL != "" {
					rawURLs := strings.Split(*session.DocURL, ",")
					for _, u := range rawURLs {
						cleanURL := strings.TrimSpace(u)
						if cleanURL != "" {
							if err := openBrowser(cleanURL); err != nil {
								log.Printf("Gagal membuka browser (%s): %v", cleanURL, err)
							} else {
								fmt.Printf("🌐 Browser dibuka: %s\n", cleanURL)
							}
							time.Sleep(300 * time.Millisecond)
						}
					}
				}

				lastHandledState = "FOCUSING"
				lastHandledTopic = currentTopic
			}
		} else if session.State == "IDLE" || session.State == "SURRENDERED" {
			if lastHandledState == "FOCUSING" {
				fmt.Printf("🛑 Sesi selesai/berubah menjadi: %s\n", session.State)
			}
			lastHandledState = session.State
			lastHandledTopic = ""
		}

		time.Sleep(2 * time.Second)
	}
}

// Helper membuka VS Code
func openVSCode(path string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "code", path)
	} else {
		cmd = exec.Command("code", path)
	}
	return cmd.Start()
}

// Helper membuka Browser
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