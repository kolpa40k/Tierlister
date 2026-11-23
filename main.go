package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"tierlist-app/storage"
	"tierlist-app/ui"
)

func main() {
	myApp := app.New()

	app.SetMetadata(fyne.AppMetadata{
		Name:    "Tierlist",
		Version: "1.1.0",
		ID:      "com.kolpa40k.tierlist",
	})

	settings, err := storage.LoadSettings()
	if err != nil {
		log.Printf("не удалось загрузить настройки: %v", err)
	}
	if settings == nil || settings.ThemeName == "" {
		settings = &storage.Settings{ThemeName: "Discord"}
	}
	ui.ApplyPreset(myApp, settings.ThemeName)

	myWindow := myApp.NewWindow("Tier List Maker")
	myWindow.Resize(fyne.NewSize(900, 600))

	categories, loadErr := storage.LoadCategories()
	if loadErr != nil {
		log.Printf("не удалось загрузить категории: %v", loadErr)
	}

	if loadErr == nil && len(categories) > 0 {
		changed, migErr := storage.MigrateAllImages(categories)
		if migErr != nil {
			log.Printf("ошибка миграции картинок: %v", migErr)
		}
		if changed {
			if err := storage.SaveCategories(categories); err != nil {
				log.Printf("не удалось сохранить после миграции: %v", err)
			}
		}
	}

	if loadErr == nil {
		myWindow.SetContent(ui.NewHomeScreen(myWindow, categories))
	} else {
		myWindow.SetContent(widget.NewLabel("Не удалось загрузить данные. Исправьте файл категорий перед повторным запуском."))
	}

	if loadErr != nil {
		myApp.Lifecycle().SetOnStarted(func() {
			dialog.ShowError(
				fmt.Errorf(
					"не удалось загрузить сохранённые категории.\n"+
						"Сохранение отключено, чтобы не перезаписать повреждённый файл.\n\n%v",
					loadErr,
				),
				myWindow,
			)
		})
	}

	myWindow.ShowAndRun()
}
