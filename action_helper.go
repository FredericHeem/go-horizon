package horizon

import (
	"net/http"
	"strconv"

	gctx "github.com/goji/context"
	"github.com/stellar/go-horizon/db"
	"github.com/stellar/go-stellar-base"
	"github.com/zenazn/goji/web"
	"golang.org/x/net/context"
)

// ActionHelper wraps the goji context and provides helper functions
// to make defining actions easier.
//
// Notably, this object provides a means of more simply extracting information
// from the Env and URLParams.  Any call to the Get* methods (GetInt, GetString, etc.)
// that fails will populate the Err field and subsequent calls to Get* will be no
// ops.  This allows the simpler pattern:
//
//	ah = &ActionHelper{C:c}
//	id := ah.GetInt("id")
//	order := ah.GetString("order")
//
//	if ah.Err() != nil {
//	  // write an error response here
//	  return
//	}
//
type ActionHelper struct {
	c   web.C
	r   *http.Request
	err error
}

func (a *ActionHelper) Err() error {
	return a.err
}

func (a *ActionHelper) App() *App {
	return a.c.Env["app"].(*App)
}

func (a *ActionHelper) Context() context.Context {
	return gctx.FromC(a.c)
}

// Gets a string from either the URLParams or query string.
// This method prioritizes the URLParams over the Query.
//
// TODO: Add form support, prioritized over query
func (a *ActionHelper) GetString(name string) string {
	if a.err != nil {
		return ""
	}

	fromUrl, ok := a.c.URLParams[name]

	if ok {
		return fromUrl
	}

	return a.r.URL.Query().Get(name)
}

func (a *ActionHelper) GetInt64(name string) int64 {
	if a.err != nil {
		return 0
	}

	asStr := a.GetString(name)

	if asStr == "" {
		return 0
	}

	asI64, err := strconv.ParseInt(asStr, 10, 64)

	if err != nil {
		a.err = err
		return 0
	}

	return asI64
}

func (a *ActionHelper) GetInt32(name string) int32 {
	if a.err != nil {
		return 0
	}

	asStr := a.GetString(name)

	if asStr == "" {
		return 0
	}

	asI64, err := strconv.ParseInt(asStr, 10, 32)

	if err != nil {
		a.err = err
		return 0
	}

	return int32(asI64)
}

func (a *ActionHelper) GetPagingParams() (cursor string, order string, limit int32) {
	if a.err != nil {
		return
	}

	cursor = a.GetString("cursor")
	order = a.GetString("order")
	limit = a.GetInt32("limit")

	if lei := a.r.Header.Get("Last-Event-ID"); lei != "" {
		cursor = lei
	}

	return
}

func (a *ActionHelper) GetPageQuery() db.PageQuery {
	r, err := db.NewPageQuery(a.GetPagingParams())
	a.err = err
	return r
}

// GetAddress reads a base58-check encoded address from the provided parameter
// name.  If the value is not a valid Stellar address, an error is set on the
// ActionHelper.
func (a *ActionHelper) GetAddress(name string) string {
	if a.err != nil {
		return ""
	}

	address := a.GetString(name)

	_, err := stellarbase.DecodeBase58Check(stellarbase.VersionByteAccountID, address)
	a.err = err

	return address
}
