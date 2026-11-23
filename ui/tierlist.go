package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"tierlist-app/clipboardimage"
	"time"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/wierdling/gofiledialog"

	"tierlist-app/editor"
	"tierlist-app/models"
	appstorage "tierlist-app/storage"
)

type TierListScreen struct {
	window          fyne.Window
	categories      *[]models.Category
	catIdx          int
	tlIdx           int
	tierList        models.TierList
	tierRows        []*TierRow
	tiersBox        *fyne.Container
	topBar          *fyne.Container
	poolBox         *fyne.Container
	poolSection     fyne.CanvasObject
	editMode        bool
	editBtn         *widget.Button
	deleteAllBtn    *widget.Button
	content         fyne.CanvasObject
	pasteShortcut   fyne.Shortcut
	dragPlaceholder fyne.CanvasObject
	dragDestination int
	dragIndex       int
}

func NewTierListScreen(win fyne.Window, categories *[]models.Category, catIdx, tlIdx int, tl models.TierList) fyne.CanvasObject {
	ts := &TierListScreen{
		window: win, categories: categories,
		catIdx: catIdx, tlIdx: tlIdx, tierList: tl,
	}
	ts.tierList.RepairIDs()
	ts.buildUI()
	ts.setupDropHandler()
	ts.setupPasteShortcut()
	ts.setupCloseHandler()
	return ts.content
}

func (ts *TierListScreen) buildUI() {
	backBtn := widget.NewButtonWithIcon("Назад", theme.NavigateBackIcon(), func() {
		ts.window.SetOnDropped(nil)
		ts.removePasteShortcut()
		ts.window.SetContent(NewCategoryScreen(ts.window, ts.categories, ts.catIdx, (*ts.categories)[ts.catIdx]))
	})
	saveBtn := widget.NewButtonWithIcon("Сохранить", theme.DocumentSaveIcon(), func() {
		if ts.save() {
			dialog.ShowInformation("Сохранено", "Изменения сохранены", ts.window)
		}
	})
	ts.editBtn = widget.NewButtonWithIcon("Редактировать тиры", theme.DocumentCreateIcon(), func() { ts.toggleEditMode() })
	initializingDisplay := true
	labelsToggle := widget.NewCheck("Подписи", func(checked bool) {
		if initializingDisplay {
			return
		}
		ts.change(func(list *models.TierList) error { list.ShowLabels = &checked; return nil })
	})
	labelsToggle.SetChecked(ts.tierList.LabelsVisible())
	fitSelect := widget.NewSelect([]string{"Заполнить", "Вместить"}, func(value string) {
		if initializingDisplay {
			return
		}
		fit := "cover"
		if value == "Вместить" {
			fit = "contain"
		}
		ts.change(func(list *models.TierList) error { list.ImageFit = fit; return nil })
	})
	if ts.tierList.ImageFillMode() == "contain" {
		fitSelect.SetSelected("Вместить")
	} else {
		fitSelect.SetSelected("Заполнить")
	}
	initializingDisplay = false
	ts.topBar = container.NewHBox(backBtn, saveBtn, ts.editBtn, labelsToggle, fitSelect)

	title := widget.NewLabelWithStyle(ts.tierList.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewVBox(ts.topBar, container.NewPadded(title))

	ts.tiersBox = container.NewVBox()
	ts.tierRows = nil
	for i := range ts.tierList.Tiers {
		row := NewTierRow(&ts.tierList.Tiers[i], ts)
		ts.tierRows = append(ts.tierRows, row)
		ts.tiersBox.Add(row)
	}

	ts.poolBox = container.NewGridWrap(fyne.NewSize(112, 128))
	ts.rebuildPool()
	fileBtn := widget.NewButtonWithIcon("Добавить файлы", theme.FolderOpenIcon(), ts.addImageToPool)
	pasteBtn := widget.NewButtonWithIcon("Вставить (Ctrl+V)", theme.ContentPasteIcon(), ts.pasteImageFromClipboard)
	pasteBtn.Importance = widget.HighImportance
	cleanupBtn := widget.NewButtonWithIcon("Очистить файлы", theme.DeleteIcon(), ts.confirmImageCleanup)
	ts.deleteAllBtn = widget.NewButtonWithIcon("Очистить пул", theme.DeleteIcon(), ts.deleteAllPoolItems)
	ts.deleteAllBtn.Importance = widget.DangerImportance
	ts.updateDeleteAllVisibility()

	poolHeader := container.NewBorder(nil, nil,
		widget.NewLabelWithStyle("Изображения · перетащите в тир", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(fileBtn, pasteBtn, cleanupBtn, ts.deleteAllBtn), nil)
	poolScroll := container.NewVScroll(ts.poolBox)
	poolScroll.SetMinSize(fyne.NewSize(280, 150))
	poolContent := container.NewBorder(container.NewPadded(poolHeader), nil, nil, nil, poolScroll)
	poolBG := canvas.NewRectangle(color.NRGBA{R: 37, G: 39, B: 45, A: 255})
	ts.poolSection = container.NewStack(poolBG, container.NewPadded(poolContent))

	tierScroll := container.NewVScroll(container.NewPadded(ts.tiersBox))
	ts.content = container.NewBorder(header, ts.poolSection, nil, nil, tierScroll)
}

func (ts *TierListScreen) toggleEditMode() {
	ts.editMode = !ts.editMode
	if ts.editMode {
		ts.editBtn.SetText("Готово")
	} else {
		ts.editBtn.SetText("Редактировать тиры")
	}
	for _, row := range ts.tierRows {
		row.Refresh()
	}
	ts.rebuildPool()
}

func (ts *TierListScreen) rebuildPool() {
	ts.poolBox.Objects = nil
	for i := range ts.tierList.Pool {
		item := ts.tierList.Pool[i]
		tile := NewTierItemWidget(item, true, false,
			func() { ts.deleteFromPool(item.ID) },
			func() { ts.renameInPool(item.ID) },
		)
		tile.SetDisplay(ts.tierList.LabelsVisible(), ts.tierList.ImageFillMode())
		ts.poolBox.Add(NewDraggableItem(tile, item, nil, ts))
	}
	if len(ts.tierList.Pool) == 0 {
		ts.poolBox.Add(widget.NewLabel("Здесь будут новые и возвращённые изображения"))
	}
	ts.poolBox.Refresh()
	ts.updateDeleteAllVisibility()
}

func (ts *TierListScreen) deleteFromPool(id string) {
	var imagePath string
	if ts.change(func(list *models.TierList) error {
		for _, item := range list.Pool {
			if item.ID == id {
				imagePath = item.ImagePath
				break
			}
		}
		_, err := editor.Remove(list, id)
		return err
	}) {
		ts.deleteImageIfUnused(imagePath)
	}
}

func (ts *TierListScreen) renameInPool(id string) {
	var current *models.TierItem
	for i := range ts.tierList.Pool {
		if ts.tierList.Pool[i].ID == id {
			current = &ts.tierList.Pool[i]
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
			ts.change(func(list *models.TierList) error { return editor.RenameItem(list, id, entry.Text) })
		}, ts.window,
	)
}

func (ts *TierListScreen) addImageToPool() {
	err := gofiledialog.ShowOpen(func(paths []string, err error) {
		if err != nil {
			ts.showError(err)
			return
		}
		ts.importDroppedImages(paths)
	}, ts.window,
		gofiledialog.WithTitle("Выберите картинки"),
		gofiledialog.WithFilters(
			gofiledialog.Filter{Name: "Изображения", Extensions: []string{".png", ".jpg", ".jpeg", ".webp"}},
			gofiledialog.Filter{Name: "Все файлы"},
		),
		gofiledialog.WithMultiSelect(true),
	)
	if err != nil {
		ts.showError(err)
	}
}

func (ts *TierListScreen) save() bool {
	previous := (*ts.categories)[ts.catIdx].TierLists[ts.tlIdx]
	(*ts.categories)[ts.catIdx].TierLists[ts.tlIdx] = ts.tierList
	if err := appstorage.SaveCategories(*ts.categories); err != nil {
		(*ts.categories)[ts.catIdx].TierLists[ts.tlIdx] = previous
		dialog.ShowError(err, ts.window)
		return false
	}
	return true
}

func insideObject(pos fyne.Position, obj fyne.CanvasObject) bool {
	if obj == nil || !obj.Visible() {
		return false
	}
	origin := fyne.CurrentApp().Driver().AbsolutePositionForObject(obj)
	size := obj.Size()
	return pos.X >= origin.X && pos.X < origin.X+size.Width && pos.Y >= origin.Y && pos.Y < origin.Y+size.Height
}

func (ts *TierListScreen) destinationAt(pos fyne.Position) (int, bool) {
	if insideObject(pos, ts.poolSection) {
		return -1, true
	}
	for i, row := range ts.tierRows {
		if insideObject(pos, row) {
			return i, true
		}
	}
	return 0, false
}

// previewImage is a translucent in-layout copy of the dragged image.
func previewImage(path string) fyne.CanvasObject {
	img := createThumbnail(path, 96, 96)
	if img == nil {
		img = canvas.NewImageFromResource(loadImageResource(path))
	}
	img.FillMode = canvas.ImageFillContain
	img.Translucency = 0.48
	return container.NewGridWrap(fyne.NewSize(112, 128), img)
}

func (ts *TierListScreen) clearDropHighlight() {
	for _, row := range ts.tierRows {
		row.ClearInsertionPreview()
	}
	if ts.dragPlaceholder != nil {
		objects := make([]fyne.CanvasObject, 0, len(ts.poolBox.Objects))
		for _, object := range ts.poolBox.Objects {
			if object != ts.dragPlaceholder {
				objects = append(objects, object)
			}
		}
		ts.poolBox.Objects = objects
		ts.poolBox.Refresh()
	}
	ts.dragPlaceholder = nil
	ts.dragDestination = -2
	ts.dragIndex = -1
}

func (ts *TierListScreen) highlightDropTarget(item models.TierItem, pos fyne.Position) {
	destination, ok := ts.destinationAt(pos)
	if !ok {
		ts.clearDropHighlight()
		return
	}
	var inx int
	if destination == -1 {
		inx = ts.poolInsertionIndex(pos)
	} else {
		inx = ts.tierRows[destination].InsertionIndex(pos)
	}
	if ts.dragPlaceholder != nil && ts.dragDestination == destination && ts.dragIndex == inx {
		return
	}
	ts.clearDropHighlight()
	placeholder := previewImage(item.ImagePath)
	ts.dragPlaceholder = placeholder
	ts.dragDestination = destination
	ts.dragIndex = inx
	if destination == -1 {
		objects := append([]fyne.CanvasObject(nil), ts.poolBox.Objects...)
		if inx > len(objects) {
			inx = len(objects)
		}
		objects = append(objects, nil)
		copy(objects[inx+1:], objects[inx:])
		objects[inx] = placeholder
		ts.poolBox.Objects = objects
		ts.poolBox.Refresh()
	} else {
		ts.tierRows[destination].ShowInsertionPreview(placeholder, inx)
	}
}

func (ts *TierListScreen) poolInsertionIndex(pos fyne.Position) int {
	index := 0
	for _, object := range ts.poolBox.Objects {
		if object == ts.dragPlaceholder {
			continue
		}
		origin := fyne.CurrentApp().Driver().AbsolutePositionForObject(object)
		if pos.Y < origin.Y+object.Size().Height/2 {
			return index
		}
		if pos.Y < origin.Y+object.Size().Height && pos.X < origin.X+object.Size().Width/2 {
			return index
		}
		index++
	}
	return index
}

func (ts *TierListScreen) handleDrop(d *DraggableItem, pos fyne.Position) {
	destination, ok := ts.destinationAt(pos)
	index := -1
	if ok {
		if destination == -1 {
			index = ts.poolInsertionIndex(pos)
		} else {
			index = ts.tierRows[destination].InsertionIndex(pos)
		}
	}
	ts.clearDropHighlight()
	if !ok {
		return
	}
	// The preview index is measured before removing the dragged item.
	// Correct it for reordering within the same destination.
	if destination == -1 {
		for i, item := range ts.tierList.Pool {
			if item.ID == d.item.ID && i < index {
				index--
				break
			}
		}
	} else {
		for i, item := range ts.tierList.Tiers[destination].Items {
			if item.ID == d.item.ID && i < index {
				index--
				break
			}
		}
	}
	ts.change(func(list *models.TierList) error { return editor.MoveAt(list, d.item.ID, destination, index) })
}

func (ts *TierListScreen) confirmImageCleanup() {
	report, err := appstorage.CleanupImages(*ts.categories, true)
	if err != nil {
		ts.showError(err)
		return
	}
	if report.Orphaned == 0 {
		dialog.ShowInformation("Хранилище", "Неиспользуемых изображений нет", ts.window)
		return
	}
	dialog.ShowConfirm("Очистить хранилище?", fmt.Sprintf("Найдено %d неиспользуемых изображений. Удалить их без возможности восстановления?", report.Orphaned), func(ok bool) {
		if !ok {
			return
		}
		cleaned, cleanupErr := appstorage.CleanupImages(*ts.categories, false)
		if cleanupErr != nil {
			ts.showError(cleanupErr)
			return
		}
		dialog.ShowCustom("Очистка завершена", "Закрыть", container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Удалено изображений: %d", cleaned.Deleted)),
			widget.NewLabel(fmt.Sprintf("Освобождено: %s", formatDataSize(cleaned.BytesReclaimed))),
		), ts.window)
	}, ts.window)
}

func (ts *TierListScreen) updateDeleteAllVisibility() {
	if ts.deleteAllBtn == nil {
		return
	}
	if len(ts.tierList.Pool) == 0 {
		ts.deleteAllBtn.Hide()
	} else {
		ts.deleteAllBtn.Show()
	}
}

func (ts *TierListScreen) deleteAllPoolItems() {
	if len(ts.tierList.Pool) == 0 {
		return
	}
	count := len(ts.tierList.Pool)
	dialog.ShowConfirm(
		"Удалить все непривязанные картинки?",
		fmt.Sprintf(
			"Будет удалено %d картинок из панели.\n"+
				"Картинки, уже размещённые в тирах, останутся.",
			count,
		),
		func(ok bool) {
			if !ok {
				return
			}
			oldPool := append([]models.TierItem(nil), ts.tierList.Pool...)
			if ts.change(func(list *models.TierList) error { list.Pool = nil; return nil }) {
				for _, it := range oldPool {
					ts.deleteImageIfUnused(it.ImagePath)
				}
			}
		}, ts.window,
	)
}

func (ts *TierListScreen) setupDropHandler() {
	ts.window.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		paths := imageDropPaths(uris)
		if len(paths) == 0 {
			return
		}
		fyne.Do(func() { ts.importDroppedImages(paths) })
	})
}

func imageDropPaths(uris []fyne.URI) []string {
	paths := make([]string, 0, len(uris))
	for _, uri := range uris {
		if uri == nil || uri.Scheme() != "file" {
			continue
		}
		path := uri.Path()
		switch strings.ToLower(filepath.Ext(path)) {
		case ".png", ".jpg", ".jpeg", ".webp":
			if runtime.GOOS == "windows" {
				path = filepath.FromSlash(strings.TrimPrefix(path, "/"))
			}
			paths = append(paths, path)
		}
	}
	return paths
}

func (ts *TierListScreen) importDroppedImages(paths []string) {
	if len(paths) == 0 {
		return
	}
	imported := make([]models.TierItem, 0, len(paths))
	for _, path := range paths {
		id := models.NewItemID()
		storedPath, err := appstorage.ImportImage(path, id)
		if err != nil {
			ts.showError(fmt.Errorf("импорт %s: %w", filepath.Base(path), err))
			continue
		}
		imported = append(imported, models.TierItem{ID: id, Label: filepath.Base(path), ImagePath: storedPath})
	}
	if len(imported) == 0 {
		return
	}
	if len(imported) == 1 {
		item := imported[0]
		ts.promptImportedImageName(item.ID, item.ImagePath, item.Label)
		return
	}
	if !ts.change(func(list *models.TierList) error {
		list.Pool = append(list.Pool, imported...)
		return nil
	}) {
		// No document references these files if the save failed.
		for _, item := range imported {
			appstorage.DeleteImage(item.ImagePath)
		}
	}
}

// setupPasteShortcut registers the editor paste action with the window canvas.
// The same import path is used by the Paste button and Ctrl+V / Cmd+V.
func (ts *TierListScreen) setupPasteShortcut() {
	shortcut := &fyne.ShortcutPaste{Clipboard: fyne.CurrentApp().Clipboard()}
	ts.pasteShortcut = shortcut
	ts.window.Canvas().AddShortcut(shortcut, func(fyne.Shortcut) {
		ts.pasteImageFromClipboard()
	})
}

func (ts *TierListScreen) removePasteShortcut() {
	if ts.pasteShortcut != nil {
		ts.window.Canvas().RemoveShortcut(ts.pasteShortcut)
		ts.pasteShortcut = nil
	}
}

func (ts *TierListScreen) setupCloseHandler() {
	ts.window.SetCloseIntercept(func() {
		fyne.CurrentApp().Quit()
	})
}

// deleteImageIfUnused removes an image only when no saved list references it.
func (ts *TierListScreen) deleteImageIfUnused(path string) {
	if path == "" || models.ImageInUse(*ts.categories, path) {
		return
	}
	if err := appstorage.DeleteImageChecked(path); err != nil {
		ts.showError(err)
	}
}

func (ts *TierListScreen) showError(err error) { dialog.ShowError(err, ts.window) }

// pasteImageFromClipboard reads bytes off the UI thread and asks for a label
// before the image becomes part of the document.
func (ts *TierListScreen) pasteImageFromClipboard() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		data, err := clipboardimage.ReadImage(ctx)
		if err != nil {
			fyne.Do(func() { ts.showError(err) })
			return
		}
		id := models.NewItemID()
		path, err := appstorage.ImportImageBytes(data, id)
		fyne.Do(func() {
			if err != nil {
				ts.showError(err)
				return
			}
			ts.promptImportedImageName(id, path, "Из буфера")
		})
	}()
}

func (ts *TierListScreen) promptImportedImageName(id, path, suggested string) {
	entry := widget.NewEntry()
	entry.SetText(suggested)
	dialog.ShowForm("Название изображения", "Добавить", "Отмена",
		[]*widget.FormItem{widget.NewFormItem("Название", entry)}, func(ok bool) {
			if !ok {
				appstorage.DeleteImage(path)
				return
			}
			label := strings.TrimSpace(entry.Text)
			if label == "" {
				label = suggested
			}
			if !ts.change(func(list *models.TierList) error {
				list.Pool = append(list.Pool, models.TierItem{ID: id, Label: label, ImagePath: path})
				return nil
			}) {
				appstorage.DeleteImage(path)
			}
		}, ts.window)
}
