package ui

import (
	"errors"
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"tierlist-app/editor"
	"tierlist-app/models"
)

type TierRow struct {
	widget.BaseWidget
	tier            *models.Tier
	itemsBox        *fyne.Container
	nameBtn         *TierNameButton
	screen          *TierListScreen
	dragPlaceholder fyne.CanvasObject
	dragIndex       int
}

func NewTierRow(tier *models.Tier, screen *TierListScreen) *TierRow {
	tr := &TierRow{tier: tier, screen: screen}
	tr.itemsBox = container.NewHBox()
	tr.nameBtn = NewTierNameButton(tier.Name, screen.editMode, func() {
		tr.editName()
	})
	tr.ExtendBaseWidget(tr)
	tr.Refresh()
	return tr
}

func (tr *TierRow) Refresh() {
	tr.itemsBox.Objects = nil
	for i := range tr.tier.Items {
		item := tr.tier.Items[i]
		itemWidget := NewTierItemWidget(
			item,
			true, true,
			func() { tr.returnToPool(item.ID) },
			func() { tr.renameItem(item.ID) },
		)
		itemWidget.SetDisplay(tr.screen.tierList.LabelsVisible(), tr.screen.tierList.ImageFillMode())
		draggable := NewDraggableItem(itemWidget, item, tr.tier, tr.screen)
		tr.itemsBox.Add(draggable)
	}
	tr.itemsBox.Refresh()
	tr.nameBtn.SetData(tr.tier.Name, tr.screen.editMode)
}

func (tr *TierRow) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(
		color.NRGBA{R: 48, G: 51, B: 58, A: 255},
	)
	background.CornerRadius = 8

	nameBox := container.NewGridWrap(
		fyne.NewSize(108, itemCardHeight+16),
		tr.nameBtn,
	)

	itemsScroll := container.NewHScroll(tr.itemsBox)
	// The viewport is taller than the cards, so the overlay scrollbar stays
	// below captions instead of obscuring them.
	itemsScroll.SetMinSize(fyne.NewSize(120, itemCardHeight+16))
	// Reserve a dedicated strip below the cards for the horizontal scrollbar.
	scrollWithGutter := container.NewBorder(nil, container.NewGridWrap(fyne.NewSize(1, 16), canvas.NewRectangle(color.NRGBA{R: 48, G: 51, B: 58, A: 255})), nil, nil, itemsScroll)

	content := container.NewBorder(
		nil,
		nil,
		nameBox,
		nil,
		scrollWithGutter,
	)

	return widget.NewSimpleRenderer(
		container.NewStack(background, content),
	)
}

func (tr *TierRow) returnToPool(id string) {
	tr.screen.change(func(list *models.TierList) error { return editor.Move(list, id, -1) })
}

func (tr *TierRow) renameItem(id string) {
	var current *models.TierItem
	for i := range tr.tier.Items {
		if tr.tier.Items[i].ID == id {
			current = &tr.tier.Items[i]
			break
		}
	}
	if current == nil {
		return
	}

	entry := widget.NewEntry()
	entry.SetText(current.Label)
	entryBox := container.NewGridWrap(fyne.NewSize(360, 40), entry)

	dialog.ShowForm("Переименовать подпись", "Сохранить", "Отмена",
		[]*widget.FormItem{widget.NewFormItem("Подпись", entryBox)},
		func(ok bool) {
			if !ok {
				return
			}
			tr.screen.change(func(list *models.TierList) error { return editor.RenameItem(list, id, entry.Text) })
		}, tr.screen.window,
	)
}

func (tr *TierRow) editName() {
	entry := widget.NewEntry()
	entry.SetText(tr.tier.Name)
	entryBox := container.NewGridWrap(fyne.NewSize(360, 40), entry)

	hint := widget.NewLabel("")
	hint.Alignment = fyne.TextAlignTrailing

	var dlg *dialog.CustomDialog

	saveBtn := widget.NewButton("Сохранить", func() {
		name := strings.TrimSpace(entry.Text)
		if name == "" {
			dialog.ShowError(errors.New("название не может быть пустым"), tr.screen.window)
			return
		}
		if utf8.RuneCountInString(name) > 30 {
			dialog.ShowError(
				fmt.Errorf("название не может быть длиннее 30 символов (сейчас %d)",
					utf8.RuneCountInString(name)),
				tr.screen.window,
			)
			return
		}
		tierIndex := -1
		for i := range tr.screen.tierRows {
			if tr.screen.tierRows[i] == tr {
				tierIndex = i
				break
			}
		}
		if tierIndex < 0 {
			return
		}
		if tr.screen.change(func(list *models.TierList) error { list.Tiers[tierIndex].Name = name; return nil }) {
			dlg.Hide()
		}
	})
	saveBtn.Importance = widget.HighImportance

	updateCounter := func(s string) {
		n := utf8.RuneCountInString(s)
		if n > 30 {
			hint.SetText(fmt.Sprintf("%d / 30 — превышен лимит", n))
			saveBtn.Disable()
		} else {
			hint.SetText(fmt.Sprintf("%d / 30", n))
			saveBtn.Enable()
		}
	}
	entry.OnChanged = updateCounter
	updateCounter(entry.Text)

	content := container.NewVBox(
		widget.NewLabel("Имя тира (максимум 30 символов):"),
		entryBox,
		hint,
		container.NewHBox(layout.NewSpacer(), saveBtn),
	)

	dlg = dialog.NewCustom("Переименовать тир", "Отмена", content, tr.screen.window)
	dlg.Resize(fyne.NewSize(420, 220))
	dlg.Show()
}

// ShowInsertionPreview reserves the destination slot while dragging.
func (tr *TierRow) ShowInsertionPreview(placeholder fyne.CanvasObject, index int) {
	tr.ClearInsertionPreview()
	if placeholder == nil {
		return
	}
	if index < 0 {
		index = 0
	}
	if index > len(tr.itemsBox.Objects) {
		index = len(tr.itemsBox.Objects)
	}
	objects := append([]fyne.CanvasObject(nil), tr.itemsBox.Objects...)
	objects = append(objects, nil)
	copy(objects[index+1:], objects[index:])
	objects[index] = placeholder
	tr.itemsBox.Objects = objects
	tr.dragPlaceholder = placeholder
	tr.dragIndex = index
	tr.itemsBox.Refresh()
}

func (tr *TierRow) ClearInsertionPreview() {
	if tr.dragPlaceholder == nil {
		return
	}
	objects := make([]fyne.CanvasObject, 0, len(tr.itemsBox.Objects)-1)
	for _, object := range tr.itemsBox.Objects {
		if object != tr.dragPlaceholder {
			objects = append(objects, object)
		}
	}
	tr.itemsBox.Objects = objects
	tr.dragPlaceholder = nil
	tr.itemsBox.Refresh()
}

func (tr *TierRow) InsertionIndex(pos fyne.Position) int {
	index := 0
	for _, object := range tr.itemsBox.Objects {
		if object == tr.dragPlaceholder {
			continue
		}
		origin := fyne.CurrentApp().Driver().AbsolutePositionForObject(object)
		if pos.X < origin.X+object.Size().Width/2 {
			return index
		}
		index++
	}
	return index
}
