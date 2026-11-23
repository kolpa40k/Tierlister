package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Tappable = (*TierNameButton)(nil)

type TierNameButton struct {
	widget.BaseWidget

	text     string
	editMode bool
	onTap    func()

	label *widget.Label
	bg    *canvas.Rectangle
}

func NewTierNameButton(
	text string,
	editMode bool,
	onTap func(),
) *TierNameButton {
	b := &TierNameButton{
		text:     text,
		editMode: editMode,
		onTap:    onTap,
	}

	b.label = widget.NewLabel(text)
	b.label.Wrapping = fyne.TextWrapWord
	b.label.Alignment = fyne.TextAlignCenter
	b.label.TextStyle = fyne.TextStyle{Bold: true}

	b.bg = canvas.NewRectangle(b.bgColor())
	b.bg.CornerRadius = 6

	b.ExtendBaseWidget(b)

	return b
}

func (b *TierNameButton) Tapped(_ *fyne.PointEvent) {
	if !b.editMode || b.onTap == nil {
		return
	}

	b.onTap()
}

func (b *TierNameButton) SetData(text string, editMode bool) {
	b.text = text
	b.editMode = editMode

	b.label.SetText(text)
	b.bg.FillColor = b.bgColor()
	b.bg.Refresh()
}

func (b *TierNameButton) bgColor() color.Color {
	name := strings.TrimSpace(b.text)
	if name == "" {
		return theme.Color(theme.ColorNamePrimary)
	}

	parts := strings.Fields(name)
	if len(parts) == 0 {
		return theme.Color(theme.ColorNamePrimary)
	}

	switch strings.ToUpper(parts[0]) {
	case "S":
		return color.NRGBA{R: 241, G: 115, B: 119, A: 255}
	case "A":
		return color.NRGBA{R: 248, G: 167, B: 107, A: 255}
	case "B":
		return color.NRGBA{R: 246, G: 205, B: 112, A: 255}
	case "C":
		return color.NRGBA{R: 149, G: 202, B: 139, A: 255}
	case "D":
		return color.NRGBA{R: 122, G: 169, B: 222, A: 255}
	default:
		return theme.Color(theme.ColorNamePrimary)
	}
}

func (b *TierNameButton) CreateRenderer() fyne.WidgetRenderer {
	centered := container.NewVBox(
		layout.NewSpacer(),
		b.label,
		layout.NewSpacer(),
	)

	return widget.NewSimpleRenderer(
		container.NewStack(
			b.bg,
			container.NewPadded(centered),
		),
	)
}
