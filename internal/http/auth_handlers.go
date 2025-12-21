package http

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB *pgxpool.Pool
}

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

type debugUserResponse struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func generateToken(userID string) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = []byte("supersecretapikey")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var body signupRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	if body.Email == "" || body.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(body.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to hash password")
	}

	ctx := userContext(c)

	var userID string
	err = h.DB.QueryRow(
		ctx,
		`INSERT INTO users (email, password_hash, full_name)
         VALUES ($1, $2, $3)
         RETURNING id`,
		body.Email, string(hashedPassword), body.FullName,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fiber.NewError(fiber.StatusConflict, "email already exists")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not create user")
	}

	token, err := generateToken(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create token")
	}

	return c.JSON(authResponse{Token: token})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body loginRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	if body.Email == "" || body.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password required")
	}

	var (
		userID       string
		passwordHash string
	)

	ctx := userContext(c)
	err := h.DB.QueryRow(
		ctx,
		`SELECT id, password_hash FROM users WHERE email = $1`,
		body.Email,
	).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch user")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(body.Password),
	); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	token, err := generateToken(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create token")
	}

	return c.JSON(authResponse{Token: token})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	uid := getUserID(c)
	if uid == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	var step string
	ctx := userContext(c)
	if err := h.DB.QueryRow(ctx, `SELECT onboarding_step FROM users WHERE id = $1`, uid).Scan(&step); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch user")
	}

	return c.JSON(fiber.Map{"user_id": uid, "ok": true, "onboarding_step": step})
}

func (h *AuthHandler) DebugUsers(c *fiber.Ctx) error {
	ctx := userContext(c)

	rows, err := h.DB.Query(
		ctx,
		`SELECT email, created_at FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list users")
	}
	defer rows.Close()

	users := make([]debugUserResponse, 0)
	for rows.Next() {
		var user debugUserResponse
		if err := rows.Scan(&user.Email, &user.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to list users")
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list users")
	}

	return c.JSON(users)
}

func userContext(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}
