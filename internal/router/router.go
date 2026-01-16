package router

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ishantswami13-crypto/vantro-backend/internal/admin"
	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
	handlers "github.com/ishantswami13-crypto/vantro-backend/internal/http"
	"github.com/ishantswami13-crypto/vantro-backend/internal/income"
	"github.com/ishantswami13-crypto/vantro-backend/internal/points"
	"github.com/ishantswami13-crypto/vantro-backend/internal/reports"
	"github.com/ishantswami13-crypto/vantro-backend/internal/summary"
	"github.com/ishantswami13-crypto/vantro-backend/internal/transactions"
)

type Router struct {
	AuthHandler         *handlers.AuthHandler
	IncomeHandler       *income.Handler
	ExpenseHandler      *expense.Handler
	SummaryHandler      *summary.Handler
	TxHandler           *handlers.TransactionsHandler
	TransactionsHandler *transactions.Handler
	SimpleTxHandler     *transactions.SimpleHandler
	BizHandler          *handlers.BusinessHandler
	AdminHandler        *admin.Handler
	OnboardingHandler   *handlers.OnboardingHandler
	ReportsHandler      *reports.Handler
	PointsHandler       *points.Handler
	AuthMW              fiber.Handler
}

func (r *Router) RegisterRoutes(app *fiber.App) {
	authLimiter := RateLimitAuth()
	writeLimiter := RateLimitWrite()

	if r.AuthHandler != nil {
		app.Post("/api/auth/signup", authLimiter, r.AuthHandler.Signup)
		app.Post("/api/auth/login", authLimiter, r.AuthHandler.Login)
		app.Post("/auth/demo", authLimiter, r.AuthHandler.Demo)
		app.Post("/api/auth/demo", authLimiter, r.AuthHandler.Demo)
		app.Get("/api/me", r.AuthMW, r.AuthHandler.Me)

		if strings.EqualFold(os.Getenv("DEBUG"), "true") {
			app.Get("/api/debug/users", r.AuthHandler.DebugUsers)
		}
	}

	if r.IncomeHandler != nil {
		if r.AuthMW != nil {
			app.Post("/api/incomes", r.AuthMW, writeLimiter, r.IncomeHandler.CreateIncome)
			app.Get("/api/incomes", r.AuthMW, r.IncomeHandler.ListIncomes)
		} else {
			app.Post("/api/incomes", writeLimiter, r.IncomeHandler.CreateIncome)
			app.Get("/api/incomes", r.IncomeHandler.ListIncomes)
		}
	}

	if r.ExpenseHandler != nil {
		if r.AuthMW != nil {
			app.Post("/api/expenses", r.AuthMW, writeLimiter, r.ExpenseHandler.CreateExpense)
			app.Get("/api/expenses", r.AuthMW, r.ExpenseHandler.ListExpenses)
		} else {
			app.Post("/api/expenses", writeLimiter, r.ExpenseHandler.CreateExpense)
			app.Get("/api/expenses", r.ExpenseHandler.ListExpenses)
		}
	}

	if r.SummaryHandler != nil {
		if r.AuthMW != nil {
			app.Get("/api/summary", r.AuthMW, r.SummaryHandler.GetSummary)
		} else {
			app.Get("/api/summary", r.SummaryHandler.GetSummary)
		}
	}

	if r.TxHandler != nil {
		if r.AuthMW != nil {
			app.Post("/api/transactions", r.AuthMW, writeLimiter, r.TxHandler.Create)
			app.Get("/api/transactions/summary", r.AuthMW, r.TxHandler.Summary)
			app.Get("/api/transactions", r.AuthMW, r.TxHandler.List)
		} else {
			app.Post("/api/transactions", writeLimiter, r.TxHandler.Create)
			app.Get("/api/transactions/summary", r.TxHandler.Summary)
			app.Get("/api/transactions", r.TxHandler.List)
		}
	}

	if r.BizHandler != nil {
		if r.AuthMW != nil {
			app.Get("/api/businesses", r.AuthMW, r.BizHandler.List)
			app.Post("/api/businesses", r.AuthMW, r.BizHandler.Create)
		} else {
			app.Get("/api/businesses", r.BizHandler.List)
			app.Post("/api/businesses", r.BizHandler.Create)
		}
	}

	if r.TransactionsHandler != nil {
		if r.AuthMW != nil {
			app.Get("/api/transactions", r.AuthMW, r.TransactionsHandler.ListLatest)
			app.Get("/api/transactions/summary", r.AuthMW, r.TransactionsHandler.GetSummary)
			app.Get("/api/export/transactions.csv", r.AuthMW, r.TransactionsHandler.ExportCSV)
			app.Delete("/api/transactions/:type/:id", r.AuthMW, r.TransactionsHandler.Delete)
			app.Post("/api/transactions/:type/:id/undo", r.AuthMW, r.TransactionsHandler.Undo)
		} else {
			app.Get("/api/transactions", r.TransactionsHandler.ListLatest)
			app.Get("/api/transactions/summary", r.TransactionsHandler.GetSummary)
			app.Get("/api/export/transactions.csv", r.TransactionsHandler.ExportCSV)
			app.Delete("/api/transactions/:type/:id", r.TransactionsHandler.Delete)
			app.Post("/api/transactions/:type/:id/undo", r.TransactionsHandler.Undo)
		}
	}

	// New unified transactions endpoints
	if r.SimpleTxHandler != nil && r.AuthMW != nil {
		app.Post("/transactions", r.AuthMW, writeLimiter, r.SimpleTxHandler.Create)
		app.Get("/me/transactions", r.AuthMW, r.SimpleTxHandler.List)
	}

	if r.AdminHandler != nil {
		app.Get("/api/admin/overview", r.AdminHandler.Overview)
	}

	if r.OnboardingHandler != nil {
		if r.AuthMW != nil {
			app.Post("/api/onboarding/step", r.AuthMW, r.OnboardingHandler.UpdateStep)
		} else {
			app.Post("/api/onboarding/step", r.OnboardingHandler.UpdateStep)
		}
	}

	if r.ReportsHandler != nil && r.AuthMW != nil {
		app.Get("/api/reports", r.AuthMW, r.ReportsHandler.Get)
		app.Get("/api/reports/categories", r.AuthMW, r.ReportsHandler.Categories)
		app.Get("/api/reports/statement", r.AuthMW, r.ReportsHandler.Statement)
		app.Get("/api/reports/statement.pdf", r.AuthMW, r.ReportsHandler.StatementPDF)
	}

	if r.PointsHandler != nil && r.AuthMW != nil {
		app.Get("/me/points", r.AuthMW, r.PointsHandler.PointsSummary)
		app.Get("/me/points/ledger", r.AuthMW, r.PointsHandler.PointsLedger)
		app.Get("/rewards", r.AuthMW, r.PointsHandler.Rewards)
		app.Post("/redeem", r.AuthMW, writeLimiter, r.PointsHandler.Redeem)
	}
}
