package console

import (
	"net/http"

	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"
	"github.com/murtaza-u/ellipsis/view/partial/console"

	"github.com/labstack/echo/v4"
)

func (a API) overviewPage(c echo.Context) error {
	count, err := a.db.GetUserAndClientCount(c.Request().Context())
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Console | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	var avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
	}

	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Console | Ellipsis",
			view.Console(
				"/console", avatarURL,
				console.Overview(
					int(count.ClientCount),
					int(count.UserCount),
				),
			),
		),
	})
}
