package api

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/internal/sqlc"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/a-h/templ"
	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
	pswdValidator "github.com/wagslane/go-password-validator"
)

func (Server) SignUpPage(c echo.Context) error {
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Sign Up | Account",
			view.SignUp(view.Credentials{}, map[string]error{}),
		),
	})
}

func (s Server) SignUp(c echo.Context) error {
	form := new(view.Credentials)
	if err := c.Bind(form); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Account",
				view.Error("Failed to parse form", http.StatusBadRequest),
			),
			Status: http.StatusBadRequest,
		})
	}

	errMap := make(map[string]error)

	if err := validateEmail(form.Email); err != nil {
		errMap["email"] = err
	}
	if err := validatePassword(form.Password); err != nil {
		errMap["password"] = err
	}
	if form.Password != form.ConfirmPassword {
		errMap["password"] = errors.New("passwords do not match")
		errMap["confirm_password"] = errMap["password"]
	}

	if len(errMap) != 0 {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Account",
				view.SignUp(*form, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	hash, err := argon2id.CreateHash(form.Password, argon2id.DefaultParams)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Account",
				view.Error("Failed to hash password", http.StatusInternalServerError),
			),
			Status: http.StatusInternalServerError,
		})
	}

	_, err = s.queries.CreateUser(c.Request().Context(), sqlc.CreateUserParams{
		Email:          form.Email,
		HashedPassword: sql.NullString{String: hash, Valid: true},
	})
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Account",
				view.Error("Database operation failed", http.StatusInternalServerError),
			),
			Status: http.StatusInternalServerError,
		})
	}

	isBoosted := c.Request().Header.Get("HX-Boosted") != ""
	if !isBoosted {
		c.Redirect(http.StatusFound, "/login")
	}

	// redirect to "/login"
	r := c.Response()
	r.Header().Set("HX-Redirect", "/login")

	// render empty template
	h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
	return h.Component.Render(c.Request().Context(), r)
}

func validateEmail(email string) error {
	if len(email) > 50 {
		return errors.New("too long")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid E-Mail")
	}
	return nil
}

func validatePassword(pswd string) error {
	if len(pswd) < 8 || len(pswd) > 70 {
		return errors.New("password must be between 8 and 70 characters")
	}
	err := pswdValidator.Validate(pswd, 60)
	if err != nil {
		return err
	}
	return nil
}
