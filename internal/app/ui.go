// Code generated by ifacemaker; DO NOT EDIT.

package app

import (
	"context"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/ErikKalkoken/evebuddy/internal/github"
)

// UI ...
type UI interface {
	App() fyne.App
	ClearAllCaches()
	CharacterService() CharacterService
	ESIStatusService() ESIStatusService
	EveImageService() EveImageService
	EveUniverseService() EveUniverseService
	MemCache() CacheService
	StatusCacheService() StatusCacheService
	DataPaths() map[string]string
	MainWindow() fyne.Window
	IsDeveloperMode() bool
	IsOffline() bool
	// Init initialized the app.
	// It is meant for initialization logic that requires the UI to be fully created.
	// It should be called directly after the UI was created and before the Fyne loop is started.
	Init()
	IsDesktop() bool
	IsMobile() bool
	Settings() Settings
	// ShowAndRun shows the UI and runs the Fyne loop (blocking),
	ShowAndRun()
	// CurrentCharacterID returns the ID of the current character or 0 if non it set.
	CurrentCharacterID() int32
	CurrentCharacter() *Character
	HasCharacter() bool
	LoadCharacter(id int32) error
	// UpdateStatus refreshed all status information pages.
	UpdateStatus()
	// UpdateCrossPages refreshed all pages that contain information about multiple characters.
	UpdateCrossPages()
	SetAnyCharacter() error
	UpdateAvatar(id int32, setIcon func(fyne.Resource))
	UpdateMailIndicator()
	MakeCharacterSwitchMenu(refresh func()) []*fyne.MenuItem
	UpdateGeneralSectionsAndRefreshIfNeeded(forceUpdate bool)
	UpdateGeneralSectionAndRefreshIfNeeded(ctx context.Context, section GeneralSection, forceUpdate bool)
	// UpdateCharacterAndRefreshIfNeeded runs update for all sections of a character if needed
	// and refreshes the UI accordingly.
	UpdateCharacterAndRefreshIfNeeded(ctx context.Context, characterID int32, forceUpdate bool)
	// UpdateCharacterSectionAndRefreshIfNeeded runs update for a character section if needed
	// and refreshes the UI accordingly.
	//
	// All UI areas showing data based on character sections needs to be included
	// to make sure they are refreshed when data changes.
	UpdateCharacterSectionAndRefreshIfNeeded(ctx context.Context, characterID int32, s CharacterSection, forceUpdate bool)
	AvailableUpdate() (github.VersionInfo, error)
	ShowInformationDialog(title, message string, parent fyne.Window)
	ShowConfirmDialog(title, message, confirm string, callback func(bool), parent fyne.Window)
	NewErrorDialog(message string, err error, parent fyne.Window) dialog.Dialog
	ShowErrorDialog(message string, err error, parent fyne.Window)
	// ModifyShortcutsForDialog modifies the shortcuts for a dialog.
	ModifyShortcutsForDialog(d dialog.Dialog, w fyne.Window)
	ShowUpdateStatusWindow()
	ShowLocationInfoWindow(id int64)
	ShowTypeInfoWindow(id int32)
	ShowEveEntityInfoWindow(o *EveEntity)
	ShowInfoWindow(c EveEntityCategory, id int32)
	ShowSnackbar(text string)
	WebsiteRootURL() *url.URL
	MakeWindowTitle(subTitle string) string
}
