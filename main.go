package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	Id          uint
	Title       string
	Description string
}

func main() {

	app := app.New()

	app.Settings().SetTheme(theme.LightTheme())
	window := app.NewWindow("Task Manager")
	window.Resize(fyne.NewSize(500, 600))
	window.CenterOnScreen()

	var tasks []Task
	var createContent *fyne.Container
	var tasksContent *fyne.Container
	var tasksList *widget.List

	DB, _ := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	DB.AutoMigrate(&Task{})
	DB.Find(&tasks)

	noTasksLabel := canvas.NewText("No Tasks", color.Black)

	if len(tasks) != 0 {
		noTasksLabel.Hide()
	}

	// filter

	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter tasks...")

	filterButton := widget.NewButton("Filter", func() {
		var filteredTasks []Task
		filter := filterEntry.Text
		if filter == "" {
			DB.Find(&filteredTasks)
		} else {
			DB.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", "%"+filter+"%", "%"+filter+"%").Find(&filteredTasks)
		}
		tasksList.Refresh() // Обновляем список задач в UI
		if len(filteredTasks) != 0 {
			noTasksLabel.Hide()
		}
		// Обновляем данные в списке
		tasks = filteredTasks // Замените tasks на filteredTasks в списке
		tasksList.Refresh()
	})
	filterBar := container.NewVBox(
		filterEntry,
		filterButton,
	)
	tasksBar := container.NewHBox(
		canvas.NewText("Your Tasks", color.Black),
		layout.NewSpacer(),
		widget.NewButton("+", func() {
			window.SetContent(createContent)
		}),
	)
	tasksList = widget.NewList(
		func() int {
			return len(tasks)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Default")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(tasks[lii].Title)
		},
	)

	tasksList.OnSelected = func(id widget.ListItemID) {
		detailsBar := container.NewHBox(
			canvas.NewText(
				fmt.Sprintf(
					"details about \"%s\"",
					tasks[id].Title,
				),
				color.Black,
			),
			layout.NewSpacer(),
			widget.NewButton("<-", func() {
				window.SetContent(tasksContent)
				tasksList.Unselect(id)
			}),
		)
		taskTitle := widget.NewLabel(tasks[id].Title)
		taskTitle.TextStyle = fyne.TextStyle{Bold: true}

		taskDescription := widget.NewLabel(tasks[id].Description)
		taskDescription.TextStyle = fyne.TextStyle{Italic: true}

		buttonsBox := container.NewHBox(
			widget.NewButton("delete", func() {
				DB.Delete(&Task{}, "tasks.id = ?", tasks[id].Id)
				DB.Find(&tasks)
				if len(tasks) == 0 {
					noTasksLabel.Show()
				} else {
					noTasksLabel.Hide()
				}
				window.SetContent(tasksContent)
			},
			),
			widget.NewButton("edit", func() {
				editBar := container.NewHBox(
					canvas.NewText(
						fmt.Sprintf(
							"editing about \"%s\"",
							tasks[id].Title,
						),
						color.Black,
					),
					layout.NewSpacer(),
					widget.NewButton("<-", func() {
						window.SetContent(tasksContent)
						tasksList.Unselect(id)
					}),
				)
				editTitle := widget.NewEntry()
				editTitle.SetText(tasks[id].Title)

				editDescription := widget.NewMultiLineEntry()
				editDescription.SetText(tasks[id].Description)

				/*editButton := widget.NewButton(

					"save task",

					func() {
						DB.Find(
							&Task{},
							"tasks.id = ?",
							tasks[id].Id,
						).Updates(
							Task{
								Title:       editTitle.Text,
								Description: editDescription.Text,
							},
						)
						DB.Find(&tasks)

						window.SetContent(tasksContent)
						tasksList.UnselectAll()
					},
				)*/
				editButton := widget.NewButton(
					"save task",
					func() {
						DB.Model(&Task{}).Where("id = ?", tasks[id].Id).Updates(
							Task{
								Title:       editTitle.Text,
								Description: editDescription.Text,
							},
						)
						DB.Find(&tasks)

						window.SetContent(tasksContent)
						tasksList.UnselectAll()
					},
				)

				editContent := container.NewVBox(
					editBar,
					canvas.NewLine(color.Black),

					editTitle,
					editDescription,
					editButton,
				)
				window.SetContent(editContent)
			},
			),
		)

		detailsVBox := container.NewVBox(
			detailsBar,
			canvas.NewLine(color.Black),

			taskTitle,
			taskDescription,
			buttonsBox,
		)

		window.SetContent(detailsVBox)
	}

	taskScroll := container.NewScroll(tasksList)
	taskScroll.SetMinSize(fyne.NewSize(500, 500))

	tasksContent = container.NewVBox(
		filterBar,
		canvas.NewLine(color.Black),
		tasksBar,
		canvas.NewLine(color.Black),
		noTasksLabel,
		taskScroll,
	)

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Task title...")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Task description...")

	saveTaskButton := widget.NewButton("save task", func() {
		task := Task{
			Title:       titleEntry.Text,
			Description: descriptionEntry.Text,
		}

		DB.Create(&task)
		DB.Find(&tasks)

		titleEntry.Text = ""
		titleEntry.Refresh()

		descriptionEntry.Text = ""
		descriptionEntry.Refresh()

		window.SetContent(tasksContent)

		tasksList.UnselectAll()

		if len(tasks) == 0 {
			noTasksLabel.Show()
		} else {
			noTasksLabel.Hide()
		}

	})

	createBar := container.NewHBox(
		canvas.NewText("Create new task", color.Black),
		layout.NewSpacer(),
		widget.NewButton("<-", func() {
			titleEntry.Text = ""
			titleEntry.Refresh()
			descriptionEntry.Text = ""
			descriptionEntry.Refresh()

			window.SetContent(tasksContent)

			tasksList.UnselectAll()
		}),
	)

	createContent = container.NewVBox(
		createBar,

		container.NewVBox(
			titleEntry,
			descriptionEntry,
			saveTaskButton,
		),
	)

	window.SetContent(tasksContent)
	window.Show()
	app.Run()
}
