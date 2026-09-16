package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"ficha-tracker/internal/domain"
	"ficha-tracker/internal/service"

	"github.com/google/uuid"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails binding layer exposing the services to the frontend.
type App struct {
	ctx      context.Context
	fichaSv  *service.FichaService
	acsSv    *service.ACSService
	authSv   *service.AuthService
	backupSv *service.BackupService
}

func NewApp(fichaSv *service.FichaService, acsSv *service.ACSService, authSv *service.AuthService, backupSv *service.BackupService) *App {
	return &App{fichaSv: fichaSv, acsSv: acsSv, authSv: authSv, backupSv: backupSv}
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

// RequestTypeStatView is the wire format for domain.RequestTypeStat.
type RequestTypeStatView struct {
	RequestType string `json:"requestType"`
	Count       int    `json:"count"`
}

// ACSStatView is the wire format for domain.ACSStat.
type ACSStatView struct {
	ACSID   string `json:"acsId"`
	ACSName string `json:"acsName"`
	Count   int    `json:"count"`
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

// ChangePassword verifies the current credentials and replaces the
// account's password.
func (a *App) ChangePassword(username, oldPassword, newPassword string) error {
	return a.authSv.ChangePassword(username, oldPassword, newPassword)
}

// saveFile prompts the user for a destination file via the native save
// dialog. It returns an empty path (and no error) if the user cancels.
func (a *App) saveFile(title, defaultFilename, filterName, filterPattern string) (string, error) {
	return wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
		Filters:         []wailsruntime.FileFilter{{DisplayName: filterName, Pattern: filterPattern}},
	})
}

// BackupDatabase prompts the user for a destination and writes a consistent
// snapshot of the database there. Returns the chosen path, or an empty
// string if the user cancelled the dialog.
func (a *App) BackupDatabase() (string, error) {
	defaultName := fmt.Sprintf("ficha-tracker-backup-%s.db", time.Now().Format("2006-01-02"))
	path, err := a.saveFile("Salvar backup do banco de dados", defaultName, "Banco de dados (*.db)", "*.db")
	if err != nil || path == "" {
		return "", err
	}
	if err := a.backupSv.Backup(path); err != nil {
		return "", err
	}
	return path, nil
}

// ExportFichasCSV exports fichas matching the given filters to a CSV file
// chosen by the user. Returns the chosen path, or an empty string if the
// user cancelled the dialog.
func (a *App) ExportFichasCSV(name, month string) (string, error) {
	fichas, err := a.fichaSv.ListFichas(name, month)
	if err != nil {
		return "", err
	}

	defaultName := fmt.Sprintf("fichas-%s.csv", time.Now().Format("2006-01-02"))
	path, err := a.saveFile("Exportar fichas em CSV", defaultName, "CSV (*.csv)", "*.csv")
	if err != nil || path == "" {
		return "", err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Nome", "Tipo de solicitação", "ACS", "Telefone", "Avisado", "Registrado em"})
	for _, f := range fichas {
		notified := "Não"
		if f.Notified {
			notified = "Sim"
		}
		_ = w.Write([]string{f.FullName, f.RequestType, f.ACSName, f.Phone, notified, f.CreatedAt.Format(time.RFC3339)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("building csv: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("writing csv: %w", err)
	}
	return path, nil
}

// ExportStatsCSV exports the request-type and ACS analytics for the given
// window to a CSV file chosen by the user. Returns the chosen path, or an
// empty string if the user cancelled the dialog.
func (a *App) ExportStatsCSV(start, end string) (string, error) {
	requestTypeStats, err := a.fichaSv.StatsByRequestType(start, end)
	if err != nil {
		return "", err
	}
	acsStats, err := a.fichaSv.StatsByACS(start, end)
	if err != nil {
		return "", err
	}

	defaultName := fmt.Sprintf("analises-%s.csv", time.Now().Format("2006-01-02"))
	path, err := a.saveFile("Exportar análises em CSV", defaultName, "CSV (*.csv)", "*.csv")
	if err != nil || path == "" {
		return "", err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Exames por tipo de solicitação"})
	_ = w.Write([]string{"Tipo de solicitação", "Quantidade"})
	for _, s := range requestTypeStats {
		_ = w.Write([]string{s.RequestType, fmt.Sprintf("%d", s.Count)})
	}
	_ = w.Write([]string{})
	_ = w.Write([]string{"Ranking de ACS"})
	_ = w.Write([]string{"ACS", "Quantidade"})
	for _, s := range acsStats {
		_ = w.Write([]string{s.ACSName, fmt.Sprintf("%d", s.Count)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("building csv: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("writing csv: %w", err)
	}
	return path, nil
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

// StatsByRequestType returns exam counts per request type for the given
// analysis window ("YYYY-MM-DD" dates; pass an empty string to leave a
// bound open).
func (a *App) StatsByRequestType(start, end string) ([]RequestTypeStatView, error) {
	stats, err := a.fichaSv.StatsByRequestType(start, end)
	if err != nil {
		return nil, err
	}
	views := make([]RequestTypeStatView, 0, len(stats))
	for _, s := range stats {
		views = append(views, RequestTypeStatView{RequestType: s.RequestType, Count: s.Count})
	}
	return views, nil
}

// StatsByACS returns exam counts per ACS for the given analysis window
// ("YYYY-MM-DD" dates; pass an empty string to leave a bound open).
func (a *App) StatsByACS(start, end string) ([]ACSStatView, error) {
	stats, err := a.fichaSv.StatsByACS(start, end)
	if err != nil {
		return nil, err
	}
	views := make([]ACSStatView, 0, len(stats))
	for _, s := range stats {
		views = append(views, ACSStatView{ACSID: s.ACSID.String(), ACSName: s.ACSName, Count: s.Count})
	}
	return views, nil
}
