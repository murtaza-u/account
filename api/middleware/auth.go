package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/view"

	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	db *sqlc.Queries
}

func NewAuthMiddleware(db *sqlc.Queries) AuthMiddleware {
	return AuthMiddleware{
		db: db,
	}
}

type CtxWithAuthInfo struct {
	echo.Context
	SessionID string
	UserID    string
	AvatarURL string
	Email     string
}

func (m AuthMiddleware) AuthInfo(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("auth_session")
		if err != nil {
			return next(c)
		}
		sess, err := m.db.GetSessionWithUser(c.Request().Context(), cookie.Value)
		if err != nil {
			return next(c)
		}
		return next(CtxWithAuthInfo{
			Context:   c,
			SessionID: sess.ID,
			UserID:    sess.UserID,
			Email:     sess.Email,
			AvatarURL: sess.AvatarUrl.String,
		})
	}
}

func (m AuthMiddleware) Required(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		returnTo := url.QueryEscape(c.Request().URL.RequestURI())
		redirectTo := "/login?return_to=" + returnTo
		cookie, err := c.Cookie("auth_session")
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, redirectTo)
		}
		sess, err := m.db.GetSession(c.Request().Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Redirect(http.StatusTemporaryRedirect, redirectTo)
			}
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
				Status: http.StatusInternalServerError,
			})
		}
		if time.Until(sess.ExpiresAt) <= 0 {
			return c.Redirect(http.StatusTemporaryRedirect, redirectTo)
		}
		return next(c)
	}
}

func (m AuthMiddleware) AlreadyAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("auth_session")
		if err != nil {
			return next(c)
		}
		sess, err := m.db.GetSession(c.Request().Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return next(c)
			}
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
				Status: http.StatusInternalServerError,
			})
		}
		if time.Until(sess.ExpiresAt) <= 0 {
			return next(c)
		}
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
