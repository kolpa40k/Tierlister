package ui

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xdraw "golang.org/x/image/draw"

	"tierlist-app/models"
	"tierlist-app/storage"
)

const (
	itemCardWidth     float32 = 112
	itemCardHeight    float32 = 128
	itemCaptionHeight float32 = 24
)

var cardBackground = color.NRGBA{R: 48, G: 51, B: 58, A: 255}

func loadImageResource(path string) fyne.Resource {
	if path == "" {
		return nil
	}
	tryPaths := []string{path}
	if path[0] == '/' || path[0] == '\\' {
		tryPaths = append(tryPaths, path[1:])
	}
	tryPaths = append(tryPaths, filepath.FromSlash(path))
	for _, p := range tryPaths {
		if data, err := storage.ReadImageFile(p); err == nil && len(data) > 0 {
			return fyne.NewStaticResource(filepath.Base(p), data)
		}
	}
	fyne.LogError("image could not be loaded", fmt.Errorf("invalid or missing image: %q", path))
	return nil
}

type TierItemWidget struct {
	widget.BaseWidget
	item         models.TierItem
	image        *canvas.Image
	editMode     bool
	onDelete     func()
	onRename     func()
	returnToPool bool
	showLabels   bool
	fillMode     string
	caption      *widget.Label
	captionBG    *canvas.Rectangle
}

var _ fyne.SecondaryTappable = (*TierItemWidget)(nil)

func NewTierItemWidget(item models.TierItem, editMode bool, returnToPool bool, onDelete, onRename func()) *TierItemWidget {
	w := &TierItemWidget{item: item, editMode: editMode, onDelete: onDelete, onRename: onRename, returnToPool: returnToPool, showLabels: true, fillMode: "cover"}
	w.image = canvas.NewImageFromResource(theme.BrokenImageIcon())
	w.image.FillMode = canvas.ImageFillStretch
	w.caption = widget.NewLabel(item.Label)
	w.caption.Alignment = fyne.TextAlignCenter
	w.caption.TextStyle = fyne.TextStyle{Bold: true}
	w.caption.Truncation = fyne.TextTruncateEllipsis
	w.captionBG = canvas.NewRectangle(color.NRGBA{R: 29, G: 30, B: 35, A: 255})
	w.ExtendBaseWidget(w)
	w.SetDisplay(true, "cover")
	return w
}

func (w *TierItemWidget) MinSize() fyne.Size { return fyne.NewSize(itemCardWidth, itemCardHeight) }

func (w *TierItemWidget) SetDisplay(showLabels bool, fit string) {
	if fit != "contain" {
		fit = "cover"
	}
	w.showLabels, w.fillMode = showLabels, fit
	height := int(itemCardHeight)
	if showLabels {
		height -= int(itemCaptionHeight)
	}
	updated := createThumbnailWithFit(w.item.ImagePath, int(itemCardWidth), height, fit)
	if updated == nil {
		updated = canvas.NewImageFromResource(theme.BrokenImageIcon())
	}
	w.image.Image, w.image.Resource = updated.Image, updated.Resource
	w.image.FillMode = canvas.ImageFillStretch
	w.image.Refresh()
	w.Refresh()
}

// cardLayout places the image and caption in exact slots. Unlike VBox/Border,
// it adds no theme spacing and never allows an image's intrinsic size to resize a row.
type cardLayout struct{ card *TierItemWidget }

func (l cardLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(itemCardWidth, itemCardHeight)
}
func (l cardLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	imageHeight := size.Height
	if l.card.showLabels {
		imageHeight -= itemCaptionHeight
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, imageHeight))
	for _, object := range objects[1:] {
		if !l.card.showLabels {
			object.Hide()
			continue
		}
		object.Show()
		object.Move(fyne.NewPos(0, imageHeight))
		object.Resize(fyne.NewSize(size.Width, itemCaptionHeight))
	}
}

func (w *TierItemWidget) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(cardBackground)
	content := container.New(cardLayout{card: w}, container.NewStack(bg, w.image), w.captionBG, w.caption)
	return widget.NewSimpleRenderer(content)
}

func (w *TierItemWidget) TappedSecondary(event *fyne.PointEvent) {
	if !w.editMode {
		return
	}
	items := make([]*fyne.MenuItem, 0, 2)
	if w.onRename != nil {
		items = append(items, fyne.NewMenuItem("Переименовать", w.onRename))
	}
	if w.onDelete != nil {
		title := "Удалить изображение"
		if w.returnToPool {
			title = "Вернуть в пул"
		}
		items = append(items, fyne.NewMenuItem(title, w.onDelete))
	}
	if len(items) == 0 {
		return
	}
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	cv := app.Driver().CanvasForObject(w)
	if cv == nil {
		return
	}
	menu := widget.NewPopUpMenu(fyne.NewMenu("Изображение", items...), cv)
	menu.ShowAtPosition(event.AbsolutePosition)
}

func createThumbnail(path string, width, height int) *canvas.Image {
	return createThumbnailWithFit(path, width, height, "cover")
}

func createThumbnailWithFit(path string, width, height int, fit string) *canvas.Image {
	file, err := storage.OpenFileAtPath(path)
	if err != nil {
		fyne.LogError("failed to open image file", err)
		return nil
	}
	defer func() {
		if err := file.Close(); err != nil {
			fyne.LogError("failed to close image file", err)
		}
	}()
	src, _, err := image.Decode(file)
	if err != nil {
		fyne.LogError("failed to decode image", err)
		return nil
	}
	return thumbnailFromImage(src, width, height, fit)
}

// thumbnailFromImage keeps the original alpha channel. The background is a
// separate UI object, not baked into the image. Cropping only removes clearly
// uniform, opaque near-white horizontal bars at the outer edges.
func thumbnailFromImage(src image.Image, width, height int, fit string) *canvas.Image {
	if src == nil || width <= 0 || height <= 0 {
		return nil
	}
	bounds := src.Bounds()
	if bounds.Empty() {
		return nil
	}
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	sx := float64(width) / float64(bounds.Dx())
	sy := float64(height) / float64(bounds.Dy())
	scale := max(sx, sy)
	if fit == "contain" {
		scale = min(sx, sy)
	}
	tw := max(1, int(float64(bounds.Dx())*scale+0.5))
	th := max(1, int(float64(bounds.Dy())*scale+0.5))
	x, y := (width-tw)/2, (height-th)/2
	xdraw.CatmullRom.Scale(dst, image.Rect(x, y, x+tw, y+th), src, bounds, xdraw.Src, nil)
	result := canvas.NewImageFromImage(dst)
	result.FillMode = canvas.ImageFillStretch
	return result
}
