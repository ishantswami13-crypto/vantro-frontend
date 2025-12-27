package reports

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func DownloadHandler(store *Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := strings.TrimSpace(c.Params("token"))
		if token == "" {
			return fiber.ErrNotFound
		}

		path, exp, err := store.GetByToken(c.Context(), token)
		if err != nil || time.Now().After(exp) {
			return fiber.ErrNotFound
		}

		clean := filepath.Clean(path)
		if !strings.Contains(clean, string(filepath.Separator)+"reports"+string(filepath.Separator)) {
			return fiber.ErrNotFound
		}

		f, err := os.Open(clean)
		if err != nil {
			return fiber.ErrNotFound
		}
		defer f.Close()

		stat, _ := f.Stat()

		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", "inline; filename=vantro-report.pdf")
		if stat != nil {
			return c.SendStream(f, int(stat.Size()))
		}
		return c.SendStream(f)
	}
}
