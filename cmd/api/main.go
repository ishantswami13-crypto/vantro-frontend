package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ishantswami13-crypto/vantro-backend/internal/admin"
	appapi "github.com/ishantswami13-crypto/vantro-backend/internal/api"
	"github.com/ishantswami13-crypto/vantro-backend/internal/billing"
	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
	apphttp "github.com/ishantswami13-crypto/vantro-backend/internal/http"
	"github.com/ishantswami13-crypto/vantro-backend/internal/income"
	"github.com/ishantswami13-crypto/vantro-backend/internal/points"
	"github.com/ishantswami13-crypto/vantro-backend/internal/reports"
	"github.com/ishantswami13-crypto/vantro-backend/internal/router"
	"github.com/ishantswami13-crypto/vantro-backend/internal/summary"
	"github.com/ishantswami13-crypto/vantro-backend/internal/transactions"
	"github.com/ishantswami13-crypto/vantro-backend/internal/whatsapp"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Ensure JWT_SECRET is set before starting; this is required for all JWT operations.
	_ = mustJWTSecret()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	// Legacy pgxpool for existing handlers
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("error creating pgx pool: %v", err)
	}
	defer pool.Close()

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

	app.Use(router.CorsMiddleware())
	app.Use(requestLogger())
	app.Use(apiKeyMiddleware())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"ok": true,
		})
	})
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"ok": true,
		})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Working")
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Dev token endpoint
	if strings.EqualFold(os.Getenv("ENV"), "dev") {
		app.Get("/dev/token", func(c *fiber.Ctx) error {
			secret := mustJWTSecret()
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"user_id": "11111111-1111-1111-1111-111111111111",
			})
			signed, err := token.SignedString(secret)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.JSON(fiber.Map{"token": signed})
		})
	}

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
	onboardingHandler := &apphttp.OnboardingHandler{DB: pool}
	adminHandler := admin.NewHandler(pool)
	reportsHandler := reports.NewHandler(pool)
	pointsHandler := points.NewHandler(pool)
	simpleTxRepo := transactions.NewSimpleRepo(pool)
	simpleTxHandler := transactions.NewSimpleHandler(simpleTxRepo)
	billingStore := &billing.Store{DB: db}
	razorpayClient := billing.NewRazorpayFromEnv()
	expenseStore := &expense.Store{DB: db}
	repStore := &reports.Store{DB: db}
	twilioClient := whatsapp.NewTwilioFromEnv()
	apiServer := &appapi.Server{DB: db}

	authMiddleware := buildJWTMiddleware(pool)

	// V1 endpoints (JWT only)
	app.Post("/transactions", rateLimitTransactions(), authMiddleware, apiServer.CreateTransaction)
	app.Get("/me/transactions", authMiddleware, apiServer.ListTransactions)
	app.Get("/me/points", authMiddleware, apiServer.PointsSummary)
	app.Get("/me/points/ledger", authMiddleware, apiServer.PointsLedger)
	app.Get("/rewards", apiServer.Rewards) // ok public
	app.Post("/redeem", rateLimitTransactions(), authMiddleware, apiServer.Redeem)

	// Expense v2 endpoints (phone-based)
	app.Post("/v1/expense/add", expense.AddExpenseHandler(expenseStore))
	app.Get("/v1/expense/list", expense.ListExpensesHandler(expenseStore))
	app.Get("/v1/expense/summary", expense.MonthlySummaryHandler(expenseStore))

	// Expense reports (paid)
	app.Get("/v1/expense/report", expense.MonthlyPDFHandler(expenseStore, billingStore))

	// Billing / Razorpay
	app.Post("/v1/billing/create-link", billing.CreatePaymentLinkHandler(billingStore, razorpayClient))
	app.Post("/v1/billing/webhook", billing.RazorpayWebhookHandler(billingStore, expenseStore, repStore, twilioClient))

	// WhatsApp inbound (Twilio webhook)
	app.Post("/v1/whatsapp/inbound", whatsapp.Inbound)

	// Public report download (tokenized)
	app.Get("/r/:token", reports.DownloadHandler(repStore))

	r := &router.Router{
		AuthHandler:         authHandler,
		IncomeHandler:       incomeHandler,
		ExpenseHandler:      expenseHandler,
		SummaryHandler:      summaryHandler,
		TransactionsHandler: txnHandler,
		SimpleTxHandler:     simpleTxHandler,
		BizHandler:          bizHandler,
		AdminHandler:        adminHandler,
		OnboardingHandler:   onboardingHandler,
		ReportsHandler:      reportsHandler,
		PointsHandler:       pointsHandler,
		AuthMW:              authMiddleware,
	}
	r.RegisterRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Listening on port", port)
	log.Fatal(app.Listen(":" + port))
}

func rateLimitTransactions() fiber.Handler {
	max := 60
	if v := strings.TrimSpace(os.Getenv("RATE_LIMIT_TX_MAX")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			max = parsed
		}
	}

	window := time.Minute
	if v := strings.TrimSpace(os.Getenv("RATE_LIMIT_TX_WINDOW_SECONDS")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			window = time.Duration(parsed) * time.Second
		}
	}

	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
	})
}

func requestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		log.Printf("%s %s %d %s", c.Method(), c.Path(), status, time.Since(start))
		return err
	}
}

func apiKeyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		path := strings.ToLower(strings.TrimSuffix(c.Path(), "/"))
		if path == "/health" || path == "/api/health" {
			return c.Next()
		}
		if path == "" {
			path = "/"
		}
		if path == "/" || path == "/healthz" || path == "/api/auth/demo" || path == "/auth/demo" {
			return c.Next()
		}

		env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
		expected := strings.TrimSpace(os.Getenv("API_KEY"))

		authHeader := strings.TrimSpace(c.Get("Authorization"))
		hasBearer := strings.HasPrefix(strings.ToLower(authHeader), "bearer ")

		// In production:
		// - Browser calls should rely on JWT (Authorization: Bearer ...) only.
		// - Non-browser/internal callers (e.g. admin tools) must present a server-side API_KEY.
		if env == "production" {
			if hasBearer {
				// Browser (JWT) flow: do not enforce API_KEY.
				return c.Next()
			}

			// Non-browser/internal: require API_KEY if configured.
			if expected == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing_api_key"})
			}

			key := strings.TrimSpace(c.Get("X-API-Key"))
			if key == "" {
				key = strings.TrimSpace(c.Get("x-api-key"))
			}

			if key == "" || key != expected {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing_api_key"})
			}

			return c.Next()
		}

		// Non-production (dev/stage): keep backward compatibility but allow JWT-only flows.
		if hasBearer {
			// Allow JWT-only calls without requiring API_KEY.
			return c.Next()
		}

		// If no API_KEY is configured in non-production, allow all requests.
		if expected == "" {
			return c.Next()
		}

		// If API_KEY is set, require it for non-Bearer calls to match historical behavior.
		key := strings.TrimSpace(c.Get("X-API-Key"))
		if key == "" {
			key = strings.TrimSpace(c.Get("x-api-key"))
		}

		if key == "" || key != expected {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing_api_key"})
		}

		return c.Next()
	}
}

func buildJWTMiddleware(pool *pgxpool.Pool) fiber.Handler {
	secret := mustJWTSecret()

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

// mustJWTSecret loads JWT_SECRET from the environment or exits the process with a fatal log.
func mustJWTSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	return []byte(secret)
}
