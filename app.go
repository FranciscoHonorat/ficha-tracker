package main

import (
	"context"
	"time"

	"ficha-tracker/internal/domain"
	"ficha-tracker/internal/service"

	"github.com/google/uuid"
)

// App is the Wails binding layer exposing the services to the frontend.
type App struct {
	ctx     context.Context
	fichaSv *service.FichaService
	acsSv   *service.ACSService
	authSv  *service.AuthService
}

func NewApp(fichaSv *service.FichaService, acsSv *service.ACSService, authSv *service.AuthService) *App {
	return &App{fichaSv: fichaSv, acsSv: acsSv, authSv: authSv}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// FichaView is the wire format exposed to the frontend, kept separate from
// domain.Ficha so IPC-friendly types (plain strings) don't leak domain concerns.
type FichaView struct {
	ID          string `json:"id"`
	FullName    string `json:"fullName"`
	RequestType string `json:"requestType"`
	ACSID       string `json:"acsId"`
	ACSName     string `json:"acsName"`
	Phone       string `json:"phone"`
	Notified    bool   `json:"notified"`
	CreatedAt   string `json:"createdAt"`
}

// ACSView is the wire format exposed to the frontend for domain.ACS.
type ACSView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"createdAt"`
}

func toFichaView(f *domain.Ficha) FichaView {
	return FichaView{
		ID:          f.ID.String(),
		FullName:    f.FullName,
		RequestType: f.RequestType,
		ACSID:       f.ACSID.String(),
		ACSName:     f.ACSName,
		Phone:       f.Phone,
		Notified:    f.Notified,
		CreatedAt:   f.CreatedAt.Format(time.RFC3339),
	}
}

func toFichaViews(fichas []*domain.Ficha) []FichaView {
	views := make([]FichaView, 0, len(fichas))
	for _, f := range fichas {
		views = append(views, toFichaView(f))
	}
	return views
}

func toACSView(a *domain.ACS) ACSView {
	return ACSView{
		ID:        a.ID.String(),
		Name:      a.Name,
		Phone:     a.Phone,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}

func toACSViews(list []*domain.ACS) []ACSView {
	views := make([]ACSView, 0, len(list))
	for _, a := range list {
		views = append(views, toACSView(a))
	}
	return views
}

// HasAccount reports whether any account has been created yet.
func (a *App) HasAccount() (bool, error) {
	return a.authSv.HasAccount()
}

// CreateAccount creates the local login account used to gate the app.
func (a *App) CreateAccount(username, password string) error {
	_, err := a.authSv.CreateAccount(username, password)
	return err
}

// Login verifies the given credentials against the stored account.
func (a *App) Login(username, password string) error {
	return a.authSv.Login(username, password)
}

// ListRequestTypes returns the canonical list of ficha request types.
func (a *App) ListRequestTypes() []string {
	return domain.RequestTypes
}

// RegisterACS validates and persists a new ACS.
func (a *App) RegisterACS(name, phone string) (ACSView, error) {
	acs, err := a.acsSv.RegisterACS(name, phone)
	if err != nil {
		return ACSView{}, err
	}
	return toACSView(acs), nil
}

// UpdateACS validates and persists changes to an existing ACS.
func (a *App) UpdateACS(id, name, phone string) (ACSView, error) {
	acsID, err := uuid.Parse(id)
	if err != nil {
		return ACSView{}, domain.ErrInvalidID
	}
	acs, err := a.acsSv.UpdateACS(acsID, name, phone)
	if err != nil {
		return ACSView{}, err
	}
	return toACSView(acs), nil
}

// DeleteACS removes an ACS, refusing to do so while fichas still reference it.
func (a *App) DeleteACS(id string) error {
	acsID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidID
	}
	return a.acsSv.DeleteACS(acsID)
}

// ListACS returns every registered ACS.
func (a *App) ListACS() ([]ACSView, error) {
	list, err := a.acsSv.ListACS()
	if err != nil {
		return nil, err
	}
	return toACSViews(list), nil
}

// RegisterFicha validates and persists a new ficha print record.
func (a *App) RegisterFicha(fullName, requestType, acsID, phone string, notified bool) (FichaView, error) {
	acsUUID, err := uuid.Parse(acsID)
	if err != nil {
		return FichaView{}, domain.ErrInvalidACSID
	}
	ficha, err := a.fichaSv.RegisterFicha(fullName, requestType, acsUUID, phone, notified)
	if err != nil {
		return FichaView{}, err
	}
	return toFichaView(ficha), nil
}

// UpdateFicha validates and persists changes to an existing ficha.
func (a *App) UpdateFicha(id, fullName, requestType, acsID, phone string, notified bool) (FichaView, error) {
	fichaID, err := uuid.Parse(id)
	if err != nil {
		return FichaView{}, domain.ErrInvalidID
	}
	acsUUID, err := uuid.Parse(acsID)
	if err != nil {
		return FichaView{}, domain.ErrInvalidACSID
	}
	ficha, err := a.fichaSv.UpdateFicha(fichaID, fullName, requestType, acsUUID, phone, notified)
	if err != nil {
		return FichaView{}, err
	}
	return toFichaView(ficha), nil
}

// DeleteFicha removes a ficha by id.
func (a *App) DeleteFicha(id string) error {
	fichaID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidID
	}
	return a.fichaSv.DeleteFicha(fichaID)
}

// ListFichas returns fichas matching an optional name filter and an
// optional month filter ("YYYY-MM"), most recent first. Pass an empty
// string to skip a filter.
func (a *App) ListFichas(name, month string) ([]FichaView, error) {
	fichas, err := a.fichaSv.ListFichas(name, month)
	if err != nil {
		return nil, err
	}
	return toFichaViews(fichas), nil
}

// ListFichasByACS returns every ficha registered for the given ACS, most
// recent first.
func (a *App) ListFichasByACS(acsID string) ([]FichaView, error) {
	acsUUID, err := uuid.Parse(acsID)
	if err != nil {
		return nil, domain.ErrInvalidACSID
	}
	fichas, err := a.fichaSv.ListFichasByACS(acsUUID)
	if err != nil {
		return nil, err
	}
	return toFichaViews(fichas), nil
}

// ListAvailableMonths returns every month ("YYYY-MM") that has at least one
// ficha registered, most recent first.
func (a *App) ListAvailableMonths() ([]string, error) {
	return a.fichaSv.ListAvailableMonths()
}
