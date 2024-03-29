package oidc

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/internal/sqlc"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type tknParams struct {
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Code         string `form:"code"`
	GrantType    string `form:"grant_type"`
}

type tknResp struct {
	Err       string `json:"error,omitempty"`
	ErrDesc   string `json:"error_description,omitempty"`
	AccessTkn string `json:"access_token,omitempty"`
	TknType   string `json:"token_type,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
	Scope     string `json:"scope,omitempty"`
	IDTkn     string `json:"id_token,omitempty"`
}

func (a API) Token(c echo.Context) error {
	params := new(tknParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "bad_request",
			ErrDesc: "failed to parse form data",
		})
	}
	if params.GrantType != "authorization_code" {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "bad_request",
			ErrDesc: "invalid or unsupported grant_type",
		})
	}

	metadata, err := a.DB.GetAuthzCode(c.Request().Context(), params.Code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusBadRequest, tknResp{
				Err:     "bad_request",
				ErrDesc: "invalid or malformed authorization code",
			})
		}
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_server_error",
			ErrDesc: "database operation failed",
		})
	}

	if metadata.ClientID != params.ClientID {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "unauthorized",
			ErrDesc: "invalid client id or secret",
		})
	}
	client, err := a.DB.GetClient(c.Request().Context(), metadata.ClientID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "unauthorized",
			ErrDesc: "invalid client id or secret",
		})
	}
	match, err := argon2id.ComparePasswordAndHash(params.ClientSecret, client.SecretHash)
	if err != nil || !match {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "unauthorized",
			ErrDesc: "invalid client id or secret",
		})
	}

	accessTkn := jwt.NewWithClaims(jwt.SigningMethodEdDSA, AccessTknClaims{
		UserID: metadata.UserID,
		Scopes: strings.Split(metadata.Scopes, " "),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.BaseURL,
			Subject:   a.BaseURL + "/userinfo",
			Audience:  jwt.ClaimStrings{metadata.ClientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 1800)),
		},
	})
	accessTknStr, err := accessTkn.SignedString(a.Key.Priv)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "failed to generate access token",
		})
	}

	sessionID, err := util.GenerateRandom(25)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "failed to generate auth session id",
		})
	}

	idTknExp := time.Now().Add(time.Second * time.Duration(client.TokenExpiration))
	idTkn := jwt.NewWithClaims(jwt.SigningMethodEdDSA, IDTknClaims{
		SID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.BaseURL,
			Subject:   metadata.ClientID,
			Audience:  jwt.ClaimStrings{metadata.ClientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(idTknExp),
		},
	})
	idTknStr, err := idTkn.SignedString(a.Key.Priv)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "failed to generate id token",
		})
	}

	_, err = a.DB.CreateSession(c.Request().Context(), sqlc.CreateSessionParams{
		ID:        sessionID,
		UserID:    metadata.UserID,
		ClientID:  sql.NullString{String: metadata.ClientID, Valid: true},
		ExpiresAt: idTknExp,
		Os:        metadata.Os,
		Browser:   metadata.Browser,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "database operation failed",
		})
	}

	// invalidate auth code
	a.DB.DeleteAuthzCode(c.Request().Context(), params.Code)

	return c.JSON(http.StatusOK, tknResp{
		AccessTkn: accessTknStr,
		TknType:   "Bearer",
		ExpiresIn: 1800,
		Scope:     metadata.Scopes,
		IDTkn:     idTknStr,
	})
}
