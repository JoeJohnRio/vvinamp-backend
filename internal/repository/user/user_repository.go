package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	database "spotify-clone/internal/pkg/db/migrations/mysql"

	"golang.org/x/crypto/bcrypt"

	"log"
)

type UserRepository interface {
	GetUserQuickPicks(ctx context.Context, userID string) ([]QuickPick, error)
}

// NewAlbumRepository returns a new instance of the albumRepository (constructor)
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

type NullString struct {
	String string
	Valid  bool
}
type userRepository struct {
	db *sql.DB
}

func (r *userRepository) GetUserQuickPicks(ctx context.Context, userID string) ([]QuickPick, error) {
	query := `
        SELECT t.id, t.title, a.name, COUNT(lh.track_id) AS play_count, al.cover_art
        FROM listening_history lh
        JOIN tracks t ON lh.track_id = t.id
        JOIN track_artist ta ON ta.track_id = t.id
        JOIN artists a ON ta.artist_id = a.id
        JOIN albums al ON t.album_id = al.id
        WHERE lh.user_id = $1 AND lh.played_at >= NOW() - INTERVAL '30 days'
        GROUP BY t.id, t.title, a.name, al.cover_art
        ORDER BY play_count DESC
        LIMIT 5;
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query quick picks: %w", err)
	}
	defer rows.Close()

	var picks []QuickPick
	for rows.Next() {
		var q QuickPick
		err := rows.Scan(&q.TrackID, &q.Title, &q.ArtistName, &q.PlayCount, &q.CoverArt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quick pick: %w", err)
		}
		picks = append(picks, q)
	}

	return picks, nil
}

// type User struct {
// 	ID       string `json:"id"`
// 	Username string `json:"name"`
// 	Password string `json:"password"`
// }

type User struct {
	ID               string
	Username         string
	Email            string
	PasswordHash     string
	ProfilePicture   string
	JoinDate         time.Time
	SubscriptionType string
	LastLogin        *time.Time
}

func (user *User) Authenticate() bool {
	statement, err := database.Db.Prepare("select Password from Users WHERE Username = ?")
	if err != nil {
		log.Fatal(err)
	}
	row := statement.QueryRow(user.Username)

	var hashedPassword string
	err = row.Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Fatal(err)
		}
	}

	return CheckPasswordHash(user.PasswordHash, hashedPassword)
}

// CheckPassword hash compares raw password with it's hashed values
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (user *User) Create() {
	statement, err := database.Db.Prepare(`
        INSERT INTO users (username, email, password_hash, profile_picture, subscription_type)
        VALUES (?, ?, ?, ?, ?)
    `)
	print(statement)
	if err != nil {
		log.Fatal(err)
	}

	hashedPassword, err := HashPassword(user.PasswordHash)
	_, err = statement.Exec(user.Username, user.Email, hashedPassword, user.ProfilePicture, user.SubscriptionType)
	if err != nil {
		log.Fatal(err)
	}
}

// HashPassword hashes given password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// GetUserIdByUsername check if a user exists in database by given username
func GetUserIdByUsername(username string) (int, error) {
	statement, err := database.Db.Prepare("select ID from Users WHERE Username = ?")
	if err != nil {
		log.Fatal(err)
	}
	row := statement.QueryRow(username)

	var Id int
	err = row.Scan(&Id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return 0, err
	}

	return Id, nil
}
