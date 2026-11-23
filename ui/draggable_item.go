package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"tierlist-app/models"
)

// DraggableItem tracks the cursor in canvas coordinates. The preview is moved
// before it is shown, so it never flashes at the canvas origin.
type DraggableItem struct {
	widget.BaseWidget
	inner   fyne.CanvasObject
	item    models.TierItem
	screen  *TierListScreen
	lastPos fyne.Position
	ghost   *widget.PopUp
}

func NewDraggableItem(inner fyne.CanvasObject, item models.TierItem, _ *models.Tier, screen *TierListScreen) *DraggableItem {
	d := &DraggableItem{inner: inner, item: item, screen: screen}
	d.ExtendBaseWidget(d)
	return d
}

func (d *DraggableItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.inner)
}

func (d *DraggableItem) MinSize() fyne.Size { return fyne.NewSize(112, 128) }

func (d *DraggableItem) Dragged(ev *fyne.DragEvent) {
	d.lastPos = ev.AbsolutePosition
	if d.ghost == nil {
		img := createThumbnail(d.item.ImagePath, 96, 96)
		if img == nil {
			img = canvas.NewImageFromResource(loadImageResource(d.item.ImagePath))
		}
		img.FillMode = canvas.ImageFillCover
		img.SetMinSize(fyne.NewSize(96, 96))
		img.Translucency = 0.25
		d.ghost = widget.NewPopUp(img, d.screen.window.Canvas())
		d.ghost.Resize(fyne.NewSize(104, 104))
		d.ghost.Move(fyne.NewPos(ev.AbsolutePosition.X-52, ev.AbsolutePosition.Y-52))
		d.ghost.Show()
	} else {
		d.ghost.Move(fyne.NewPos(ev.AbsolutePosition.X-52, ev.AbsolutePosition.Y-52))
	}
	d.screen.highlightDropTarget(d.item, ev.AbsolutePosition)
}

func (d *DraggableItem) DragEnd() {
	if d.ghost != nil {
		d.ghost.Hide()
		d.ghost = nil
	}
	if d.screen != nil {
		d.screen.handleDrop(d, d.lastPos)
	}
}
