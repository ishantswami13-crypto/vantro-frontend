package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ishantswami13-crypto/vantro-backend/internal/admin"
	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
	apphttp "github.com/ishantswami13-crypto/vantro-backend/internal/http"
	"github.com/ishantswami13-crypto/vantro-backend/internal/income"
	"github.com/ishantswami13-crypto/vantro-backend/internal/router"
	"github.com/ishantswami13-crypto/vantro-backend/internal/summary"
	"github.com/ishantswami13-crypto/vantro-backend/internal/transactions"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("error creating pgx pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "internal server error"

			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
				message = fiberErr.Message
			}

			return c.Status(code).JSON(fiber.Map{"error": message})
		},
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"ok": true,
		})
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,https://vantro-frontend.onrender.com",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Working")
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	authHandler := &apphttp.AuthHandler{DB: pool}
	incomeRepo := income.NewRepository(pool)
	incomeHandler := income.NewHandler(incomeRepo)
	expenseRepo := expense.NewRepository(pool)
	expenseHandler := expense.NewHandler(expenseRepo)
	summaryRepo := summary.Repo{DB: pool}
	summaryHandler := &summary.Handler{Repo: summaryRepo}
	bizHandler := apphttp.NewBusinessHandler(pool)
	txnRepo := transactions.NewRepo(pool)
	txnHandler := transactions.NewHandler(txnRepo)
	adminHandler := admin.NewHandler(pool)
	adminMW := admin.RequireAdminAPIKey()

	authMiddleware := buildJWTMiddleware(pool)

	r := &router.Router{
		AuthHandler:         authHandler,
		IncomeHandler:       incomeHandler,
		ExpenseHandler:      expenseHandler,
		SummaryHandler:      summaryHandler,
		TransactionsHandler: txnHandler,
		BizHandler:          bizHandler,
		AdminHandler:        adminHandler,
		AuthMW:              authMiddleware,
		AdminMW:             adminMW,
	}
	r.RegisterRoutes(app)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func buildJWTMiddleware(pool *pgxpool.Pool) fiber.Handler {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = []byte("supersecretapikey")
	}

	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid token")
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		userIDVal, ok := claims["user_id"].(string)
		if !ok || strings.TrimSpace(userIDVal) == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		c.Locals("user_id", userIDVal)
		c.Locals("userID", userIDVal)

		// Update last_seen_at (best-effort, do not block request)
		go func(uid string) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, _ = pool.Exec(ctx, `UPDATE users SET last_seen_at = NOW() WHERE id = $1::uuid`, uid)
		}(userIDVal)

		return c.Next()
	}
}
