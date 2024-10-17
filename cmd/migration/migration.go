package migration

import (
	"database/sql"
	"fmt"
	"time"

	"mlvt/internal/infra/reason"
	"mlvt/internal/infra/zap-logging/log"

	"golang.org/x/crypto/bcrypt"
)

// Migration defines a database migration.
type Migration struct {
	ID   int
	Name string
	SQL  string
}

// Migrate runs all pending migrations and inserts sample data.
func Migrate(db *sql.DB) error {
	// Ensure the migrations table exists
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	// Define migrations
	migrations := []Migration{
		{
			ID:   1,
			Name: "create_users_table",
			SQL: `
                CREATE TABLE IF NOT EXISTS users (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    first_name TEXT NOT NULL,
                    last_name TEXT NOT NULL,
                    username TEXT NOT NULL UNIQUE,
                    email TEXT NOT NULL UNIQUE,
                    password TEXT NOT NULL,
                    status INTEGER NOT NULL,
                    premium BOOLEAN NOT NULL DEFAULT FALSE,
                    role TEXT NOT NULL DEFAULT 'User',
                    avatar TEXT NOT NULL DEFAULT '',
                    avatar_folder TEXT NOT NULL DEFAULT '',
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			ID:   2,
			Name: "create_videos_table",
			SQL: `
                CREATE TABLE IF NOT EXISTS videos (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    user_id INTEGER NOT NULL,
                    title TEXT NOT NULL,
                    duration INTEGER NOT NULL,
                    description TEXT,
                    file_name TEXT NOT NULL,
                    folder TEXT NOT NULL,
                    image TEXT NOT NULL,
                    status TEXT NOT NULL DEFAULT 'raw',
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
                );`,
		},
		{
			ID:   3,
			Name: "create_transcriptions_table",
			SQL: `
                CREATE TABLE IF NOT EXISTS transcriptions (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    video_id INTEGER NOT NULL,
                    user_id INTEGER NOT NULL,
                    text TEXT NOT NULL,
                    lang TEXT NOT NULL,
                    folder TEXT NOT NULL,
                    file_name TEXT NOT NULL,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE,
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
                );`,
		},
		{
			ID:   4,
			Name: "create_transaction_logs_table",
			SQL: `
                CREATE TABLE IF NOT EXISTS transaction_logs (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    order_id TEXT NOT NULL UNIQUE,
                    payment_method TEXT NOT NULL,
                    action TEXT NOT NULL,
                    status TEXT NOT NULL,
                    details TEXT,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );`,
		},
		{
			ID:   5,
			Name: "create_frames_table",
			SQL: `
                CREATE TABLE IF NOT EXISTS frames (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    video_id INTEGER NOT NULL,
                    link TEXT NOT NULL,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
                );`,
		},
		{
			ID:   6,
			Name: "create_audios_table",
			SQL: `
                CREATE TABLE IF NOT EXISTS audios (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    video_id INTEGER NOT NULL,
                    user_id INTEGER NOT NULL,
                    duration INTEGER NOT NULL,
                    lang TEXT NOT NULL,
                    folder TEXT NOT NULL,
                    file_name TEXT NOT NULL,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE,
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
                );`,
		},
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if err := applyMigration(db, migration); err != nil {
			return err
		}
	}

	// Insert sample data
	if err := insertSampleData(db); err != nil {
		return err
	}

	return nil
}

// insertSampleData inserts sample data into the database for testing purposes.
func insertSampleData(db *sql.DB) error {
	// Define sample users with plaintext passwords
	users := []struct {
		FirstName    string
		LastName     string
		UserName     string
		Email        string
		Password     string
		Status       int
		Premium      bool
		Role         string
		Avatar       string
		AvatarFolder string
	}{
		{"John", "Doe", "johndoe", "john@example.com", "SecureP@ssw0rd!", 1, false, "User", "avatar1.png", "avatars/"},
		{"Jane", "Smith", "janesmith", "jane@example.com", "AnotherP@ssw0rd!", 1, true, "Admin", "avatar2.png", "avatars/"},
		// Add more sample users as needed
	}

	for _, user := range users {
		// Hash the plaintext password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for '%s': %v", user.Email, err)
		}

		// Insert user with hashed password
		query := `
            INSERT INTO users (first_name, last_name, username, email, password, status, premium, role, avatar, avatar_folder, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(email) DO NOTHING;` // Prevent duplicate email entries
		_, err = db.Exec(query, user.FirstName, user.LastName, user.UserName, user.Email, string(hashedPassword), user.Status, user.Premium, user.Role, user.Avatar, user.AvatarFolder, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert user '%s': %v", user.Email, err)
		}
	}

	// Insert sample videos
	baseTitles := []string{
		"Introduction to Go", "Learning AWS S3 Basics", "Cooking Masterclass",
		"Travel Vlog: Paris", "Yoga for Beginners", "Machine Learning 101",
		"Advanced Golang Techniques", "Docker Basics", "ReactJS Tutorial",
		"Kubernetes Deployment", "Cloud Computing", "Database Optimization",
		"Web Development with Flask", "Microservices Architecture", "Python for Data Science",
		"Getting Started with TypeScript", "Mastering GraphQL", "DevOps Best Practices",
		"Building APIs with Node.js", "Serverless Computing Overview",
	}
	userIDs := []uint64{1, 2} // Assuming the sample users have IDs 1 and 2
	durationBase := 300       // Base duration in seconds

	for i := 1; i <= 20; i++ {
		title := fmt.Sprintf("%s - Part %d", baseTitles[i%len(baseTitles)], i)
		userID := userIDs[i%len(userIDs)]
		duration := durationBase + (i * 10)
		description := fmt.Sprintf("Description for %s", title)
		fileName := fmt.Sprintf("video%d.mp4", i)
		folder := "videos/"
		image := fmt.Sprintf("https://example.com/video%d-thumbnail.jpg", i)
		status := "raw"

		query := `
            INSERT INTO videos (user_id, title, duration, description, file_name, folder, image, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
		_, err := db.Exec(query, userID, title, duration, description, fileName, folder, image, status, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert video '%s': %v", title, err)
		}
	}

	// Insert more sample data for other tables as needed
	// For example, inserting sample transcriptions, audios, frames, and transaction_logs

	log.Info(reason.InsertSampleDataSuccess.Message())

	return nil
}

// ensureMigrationsTable ensures that the migrations table exists.
func ensureMigrationsTable(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS migrations (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    `
	_, err := db.Exec(createTableSQL)
	return err
}

// applyMigration applies a single migration if it hasn't been applied yet.
func applyMigration(db *sql.DB, migration Migration) error {
	// Check if the migration has already been applied
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", migration.Name).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Migration already applied
		log.Infof("%s %s %s", reason.Migration.Message(), migration.Name, reason.AlreadyApplied.Message())
		return nil
	}

	// Apply the migration
	log.Infof("%s %s", reason.ApplyingMigration.Message(), migration.Name)
	_, err = db.Exec(migration.SQL)
	if err != nil {
		return fmt.Errorf("failed to apply migration '%s': %v", migration.Name, err)
	}

	// Record the migration as applied
	_, err = db.Exec("INSERT INTO migrations (name) VALUES (?)", migration.Name)
	if err != nil {
		return fmt.Errorf("failed to record migration '%s': %v", migration.Name, err)
	}

	return nil
}
