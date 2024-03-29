package console

import (
	"errors"
	"net/url"
	"strings"

	"github.com/murtaza-u/ellipsis/view/partial/console"
)

type AppValidator struct {
	console.AppParams
}

func newAppValidator(p console.AppParams) AppValidator {
	return AppValidator{AppParams: p}
}

func (v AppValidator) Validate() (*console.AppParams, map[string]error) {
	errMap := make(map[string]error)
	if err := v.validateName(); err != nil {
		errMap["name"] = err
	}
	if err := v.validateLogoURL(); err != nil {
		errMap["logo_url"] = err
	}
	if err := v.validateBackchannelLogoutURL(); err != nil {
		errMap["backchannel_logout_url"] = err
	}
	if err := v.validateAuthCallbackURLs(); err != nil {
		errMap["auth_callback_urls"] = err
	}
	if err := v.validateLogoutCallbackURLs(); err != nil {
		errMap["logout_callback_urls"] = err
	}
	if err := v.validateIDTokenExpiration(); err != nil {
		errMap["id_token_expiration"] = err
	}
	return &v.AppParams, errMap
}

func (v *AppValidator) validateName() error {
	v.Name = strings.TrimSpace(v.Name)
	if len(v.Name) < 2 || len(v.Name) > 50 {
		return errors.New("name must be between 2 and 50 characters")
	}
	return nil
}

func validateURL(s string) error {
	if len(s) > 100 {
		return errors.New("url too long")
	}
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return errors.New("invalid URL")
	}
	return nil
}

func (v *AppValidator) validateLogoURL() error {
	if v.LogoURL == "" {
		return nil
	}
	v.LogoURL = strings.TrimSpace(v.LogoURL)
	v.LogoURL = strings.TrimSuffix(v.LogoURL, "/")
	return validateURL(v.LogoURL)
}

func (v *AppValidator) validateBackchannelLogoutURL() error {
	if v.BackchannelLogoutURL == "" {
		return nil
	}
	v.BackchannelLogoutURL = strings.TrimSpace(v.BackchannelLogoutURL)
	v.BackchannelLogoutURL = strings.TrimSuffix(v.BackchannelLogoutURL, "/")
	return validateURL(v.BackchannelLogoutURL)
}

func validateAndTransformCallbackURLs(urls string) (string, error) {
	if len(urls) > 1000 {
		return "", errors.New("value too long")
	}

	var callbacks []string

	for _, callback := range strings.Split(urls, ",") {
		callback = strings.TrimSpace(callback)
		callback = strings.TrimSuffix(callback, "/")
		_, err := url.ParseRequestURI(callback)
		if err != nil {
			return "", errors.New("one or more invalid URL")
		}
		callbacks = append(callbacks, callback)
	}

	return strings.Join(callbacks, ","), nil
}

func (v *AppValidator) validateAuthCallbackURLs() error {
	if v.AuthCallbackURLs == "" {
		return errors.New("missing auth callback URL")
	}

	urls, err := validateAndTransformCallbackURLs(v.AuthCallbackURLs)
	if err != nil {
		return err
	}

	v.AuthCallbackURLs = urls
	return nil
}

func (v *AppValidator) validateLogoutCallbackURLs() error {
	if v.LogoutCallbackURLs == "" {
		return errors.New("missing logout callback URL")
	}

	urls, err := validateAndTransformCallbackURLs(v.LogoutCallbackURLs)
	if err != nil {
		return err
	}

	v.LogoutCallbackURLs = urls
	return nil
}

func (v AppValidator) validateIDTokenExpiration() error {
	if v.IDTokenExpiration < 300 || v.IDTokenExpiration > 86400 {
		return errors.New("id token expiration must be between 300s to 86400s")
	}
	return nil
}
