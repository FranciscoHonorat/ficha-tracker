package main

import (
	"context"
	"time"

	"ficha-tracker/internal/domain"
	"ficha-tracker/internal/service"
)

// App is the Wails binding layer exposing FichaService to the frontend.
type App struct {
	ctx     context.Context
	fichaSv *service.FichaService
}

func NewApp(fichaSv *service.FichaService) *App {
	return &App{fichaSv: fichaSv}
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
	ACS         string `json:"acs"`
	CreatedAt   string `json:"createdAt"`
}

func toFichaView(f *domain.Ficha) FichaView {
	return FichaView{
		ID:          f.ID.String(),
		FullName:    f.FullName,
		RequestType: f.RequestType,
		ACS:         f.ACS,
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

// RegisterFicha validates and persists a new ficha print record.
func (a *App) RegisterFicha(fullName, requestType, acs string) (FichaView, error) {
	ficha, err := a.fichaSv.RegisterFicha(fullName, requestType, acs)
	if err != nil {
		return FichaView{}, err
	}
	return toFichaView(ficha), nil
}

// ListFichas returns every ficha registered so far, most recent first.
func (a *App) ListFichas() ([]FichaView, error) {
	fichas, err := a.fichaSv.ListFichas()
	if err != nil {
		return nil, err
	}
	return toFichaViews(fichas), nil
}

// SearchFichasByName returns every ficha whose name contains the given query.
func (a *App) SearchFichasByName(name string) ([]FichaView, error) {
	fichas, err := a.fichaSv.SearchFichasByName(name)
	if err != nil {
		return nil, err
	}
	return toFichaViews(fichas), nil
}
