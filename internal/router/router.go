package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
	handlers "github.com/ishantswami13-crypto/vantro-backend/internal/http"
	"github.com/ishantswami13-crypto/vantro-backend/internal/income"
	"github.com/ishantswami13-crypto/vantro-backend/internal/summary"
)

type Router struct {
	AuthHandler    *handlers.AuthHandler
	IncomeHandler  *income.Handler
	ExpenseHandler *expense.Handler
	SummaryHandler *summary.Handler
	AuthMW         fiber.Handler
}

func (r *Router) RegisterRoutes(app *fiber.App) {
	if r.AuthHandler != nil {
		app.Post("/api/auth/signup", r.AuthHandler.Signup)
		app.Post("/api/auth/login", r.AuthHandler.Login)
		app.Get("/api/me", r.AuthMW, r.AuthHandler.Me)
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
}
