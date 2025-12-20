package router

import (
	"github.com/gofiber/fiber/v2"
	"os"
	"strings"

	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
	handlers "github.com/ishantswami13-crypto/vantro-backend/internal/http"
	"github.com/ishantswami13-crypto/vantro-backend/internal/income"
	"github.com/ishantswami13-crypto/vantro-backend/internal/summary"
	"github.com/ishantswami13-crypto/vantro-backend/internal/transactions"
)

type Router struct {
	AuthHandler    *handlers.AuthHandler
	IncomeHandler  *income.Handler
	ExpenseHandler *expense.Handler
	SummaryHandler *summary.Handler
	TxHandler      *handlers.TransactionsHandler
	TxnHandler     *transactions.Handler
	BizHandler     *handlers.BusinessHandler
	AuthMW         fiber.Handler
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
			app.Post("/api/expenses", r.AuthMW, r.ExpenseHandler.AddExpense)
			app.Get("/api/expenses", r.AuthMW, r.ExpenseHandler.ListExpenses)
		} else {
			app.Post("/api/expenses", r.ExpenseHandler.AddExpense)
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

	if r.TxnHandler != nil {
		if r.AuthMW != nil {
			app.Post("/api/transactions", r.AuthMW, r.TxnHandler.Create)
			app.Get("/api/transactions", r.AuthMW, r.TxnHandler.List)
			app.Get("/api/transactions/summary", r.AuthMW, r.TxnHandler.Summary)
		} else {
			app.Post("/api/transactions", r.TxnHandler.Create)
			app.Get("/api/transactions", r.TxnHandler.List)
			app.Get("/api/transactions/summary", r.TxnHandler.Summary)
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
}
