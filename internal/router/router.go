package router

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ishantswami13-crypto/vantro-backend/internal/admin"
	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
	handlers "github.com/ishantswami13-crypto/vantro-backend/internal/http"
	"github.com/ishantswami13-crypto/vantro-backend/internal/income"
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
	BizHandler          *handlers.BusinessHandler
	AdminHandler        *admin.Handler
	OnboardingHandler   *handlers.OnboardingHandler
	AuthMW              fiber.Handler
}

func (r *Router) RegisterRoutes(app *fiber.App) {
	if r.AuthHandler != nil {
		app.Post("/api/auth/signup", r.AuthHandler.Signup)
		app.Post("/api/auth/login", r.AuthHandler.Login)
		app.Get("/api/me", r.AuthMW, r.AuthHandler.Me)

		if strings.EqualFold(os.Getenv("DEBUG"), "true") {
			app.Get("/api/debug/users", r.AuthHandler.DebugUsers)
		}
	}

	if r.IncomeHandler != nil {
		if r.AuthMW != nil {
			app.Post("/api/incomes", r.AuthMW, r.IncomeHandler.CreateIncome)
			app.Get("/api/incomes", r.AuthMW, r.IncomeHandler.ListIncomes)
		} else {
			app.Post("/api/incomes", r.IncomeHandler.CreateIncome)
			app.Get("/api/incomes", r.IncomeHandler.ListIncomes)
		}
	}

	if r.ExpenseHandler != nil {
		if r.AuthMW != nil {
			app.Post("/api/expenses", r.AuthMW, r.ExpenseHandler.CreateExpense)
			app.Get("/api/expenses", r.AuthMW, r.ExpenseHandler.ListExpenses)
		} else {
			app.Post("/api/expenses", r.ExpenseHandler.CreateExpense)
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
			app.Post("/api/transactions", r.AuthMW, r.TxHandler.Create)
			app.Get("/api/transactions/summary", r.AuthMW, r.TxHandler.Summary)
			app.Get("/api/transactions", r.AuthMW, r.TxHandler.List)
		} else {
			app.Post("/api/transactions", r.TxHandler.Create)
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
		} else {
			app.Get("/api/transactions", r.TransactionsHandler.ListLatest)
			app.Get("/api/transactions/summary", r.TransactionsHandler.GetSummary)
			app.Get("/api/export/transactions.csv", r.TransactionsHandler.ExportCSV)
			app.Delete("/api/transactions/:type/:id", r.TransactionsHandler.Delete)
		}
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
}
