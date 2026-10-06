# Desk Summon Backend 🚀

Desk Summon Backend adalah daemon service berbasis Go yang mendengarkan perubahan status fokus dari database Supabase (yang dikendalikan via aplikasi mobile/HP) dan secara otomatis menyiapkan workspace di PC/Laptop (membuka VS Code, browser dokumentasi, dll).

## ✨ Fitur
- **Supabase Polling Listener**: Memonitor state `current_session` secara realtime/polling periodik.
- **Auto Workspace Launch**: Membuka project path di VS Code secara otomatis saat sesi fokus dimulai (`FOCUSING`).
- **Auto Browser Trigger**: Membuka tautan dokumentasi atau referensi terkait di browser default.

## 🛠️ Prasyarat
- Go 1.20+
- Akun & Proyek Supabase
- VS Code CLI (`code` terdaftar di PATH environment variable)

## 🚀 Setup & Menjalankan

1. **Clone repository:**
   ```bash
   git clone https://github.com/FarhanTZ/desk-summon-backend.git
   cd desk-summon-backend
   ```

2. **Buat file `.env`:**
   ```env
   SUPABASE_URL=https://your-project.supabase.co
   SUPABASE_ANON_KEY=your-supabase-anon-key
   GEMINI_API_KEY=your-gemini-api-key
   ```

3. **Install dependencies:**
   ```bash
   go mod tidy
   ```

4. **Jalankan daemon:**
   ```bash
   go run main.go
   ```
