package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// EditableRow is a compact, Discord-inspired navigation card.
// Opening the card is its primary action; edit and delete are always available.
type EditableRow struct {
	widget.BaseWidget
	text                      string
	onTap, onDelete, onRename func()
}

func NewEditableRow(text string, _ bool, onTap, onDelete, onRename func()) *EditableRow {
	row := &EditableRow{text: text, onTap: onTap, onDelete: onDelete, onRename: onRename}
	row.ExtendBaseWidget(row)
	return row
}

func (r *EditableRow) CreateRenderer() fyne.WidgetRenderer {
	title := widget.NewButtonWithIcon(r.text, theme.FolderIcon(), r.onTap)
	title.Alignment = widget.ButtonAlignLeading
	title.Importance = widget.LowImportance
	rename := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), r.onRename)
	rename.Importance = widget.LowImportance
	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), r.onDelete)
	deleteButton.Importance = widget.LowImportance
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	bg.CornerRadius = 10
	actions := container.NewHBox(rename, deleteButton)
	content := container.NewBorder(nil, nil, nil, actions, title)
	card := container.NewStack(bg, container.NewPadded(content))
	return widget.NewSimpleRenderer(card)
}
