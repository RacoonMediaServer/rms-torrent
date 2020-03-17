package types

type SearchSession interface {
	// Инициализация сессии поиска, задаем логин пароль и пр
	Setup(settings SessionSettings)

	// Если торрент-трекер запросил капчу при логине, здесь можно указать ее распознанное значение
	SetCaptchaText(captchaText string)

	// Инициировать поиск. Если вернется types.CaptchaRequired - то надо распознать капчу
	Search(text string) ([]Torrent, error)

	// Скачать торрент-файл в указанную директорию
	Download(link, destination string) error
}
