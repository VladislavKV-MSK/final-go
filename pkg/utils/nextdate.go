// Package utils предоставляет утилиты для работы с датами, включая расчет следующих дат
// выполнения задач на основе правил повторения и поддержку стандартного формата дат.
//
// Основные возможности:
//   - Расчет следующей даты для задач с повторением (ежедневно, еженедельно, ежемесячно, ежегодно).
//   - Парсинг и валидация строковых правил повторения.
//   - Работа с особыми днями месяца (последний, предпоследний день).
//   - Конвертация между строковыми и числовыми представлениями дат.
//
// Форматы:
//   - Дата: "YYYYMMDD".
//   - Правила повторения:
//   - "y"       — ежегодно.
//   - "d N"     — каждые N дней (1 ≤ N ≤ 400).
//   - "w D1,D2" — по дням недели (1-7, где 1-понедельник, 7-воскресенье).
//   - "m D1,D2 [M1,M2]" — по дням месяца (1-31, -1 — последний день, -2 — предпоследний)
//     с опциональным списком месяцев (1-12).
package utils

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// errForamt возвращается при неверном формате входных данных.
var errForamt = fmt.Errorf("error format dstart or repeat")

// Константы для валидации:
const (
	DateFormat = "20060102" // Формат даты (YYYYMMDD)
	max_day    = 400        // Максимальный интервал для ежедневного повтора
	max_wday   = 7          // Максимальное количество дней в неделе
	max_month  = 12         // Максимальное количество месяцев
)

// nextDate рассчитывает следующую дату выполнения задачи на основе правила повтора.
//
// Параметры:
//   - now: текущее время для сравнения
//   - dstart: начальная дата в формате "YYYYMMDD"
//   - repeat: правило повтора в формате:
//   - "y" - ежегодно
//   - "d N" - каждые N дней (1 ≤ N ≤ 400)
//   - "w D1,D2,..." - по дням недели (1-7, где 1-понедельник, 7-воскресенье)
//   - "m D1,D2,... [M1,M2,...]" - по дням месяца (1-31, -1 - последний день, -2 - предпоследний)
//     с опциональным списком месяцев (1-12)
//
// Возвращает:
//   - следующую дату в формате "YYYYMMDD"
//   - ошибку при неверном формате входных данных или пустую строку для разовых задач
func NextDate(now time.Time, dstart string, repeat string) (string, error) {

	if repeat == "" { // разовая задача,  будет удалена после
		return "", nil
	}

	// париснг repeat, dstart
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", errForamt
	}

	rule := strings.Split(repeat, " ")
	ruleLen := len(rule)

	switch rule[0] {
	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				return date.Format(DateFormat), nil
			}
		}
	case "d":
		if ruleLen < 2 {
			return "", errForamt
		}
		interval, err := strconv.Atoi(rule[1])
		if err != nil || interval < 0 || interval > max_day {
			return "", errForamt
		}

		for {
			date = date.AddDate(0, 0, interval)
			if afterNow(date, now) {
				return date.Format(DateFormat), nil
			}
		}
	case "w":
		if ruleLen < 2 {
			return "", errForamt
		}
		dmap, err := parseWeek(rule[1])
		if err != nil {
			return "", err
		}
		for {
			date = date.AddDate(0, 0, 1)
			weekday := int(date.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			if dmap[weekday] && afterNow(date, now) {
				return date.Format(DateFormat), nil
			}
		}
	case "m":
		if ruleLen < 2 {
			return "", errForamt
		}
		return findMonthDay(now, date, rule[1], rule[2:]...)
	default:
		return "", errForamt
	}
}

// afterNow проверяет, что дата находится после текущего времени.
func afterNow(date, now time.Time) bool {
	return date.After(now)
}

// lastDayOfMonth возвращает последний день указанного месяца.
func lastDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location())
}

// parseWeek парсит строку с днями недели в map[int]bool.
//
// Формат: "D1,D2,..." где 1 ≤ D ≤ 7.
// Возвращает ошибку при неверном формате.
func parseWeek(s string) (map[int]bool, error) {
	days := strings.Split(s, ",")
	dmap := make(map[int]bool)
	for _, dayStr := range days {
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > max_wday {
			return nil, errForamt
		}
		dmap[day] = true
	}
	return dmap, nil
}

// parseMonth парсит строку с месяцами в map[int]bool.
//
// Формат: "M1,M2,..." где 1 ≤ M ≤ 12.
// Если список пуст, используются все месяцы.
// Возвращает ошибку при неверном формате.
func parseMonth(months []string) (map[int]bool, error) {
	monthMap := make(map[int]bool)
	if len(months) > 0 {
		monthStrs := strings.Split(months[0], ",")
		for _, monthStr := range monthStrs {
			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > max_month {
				return monthMap, errForamt
			}
			monthMap[month] = true
		}
	} else {
		for i := 1; i < 13; i++ {
			monthMap[i] = true
		}
	}
	return monthMap, nil

}

// arrangeSpecialDays упорядочивает дни месяца, помещая специальные дни (-1, -2) в конец.
func arrangeSpecialDays(days []int) []int {
	var regularDays, minusTwo, minusOne []int

	// Разделяем дни на три группы
	for _, day := range days {
		switch day {
		case -2:
			minusTwo = append(minusTwo, day)
		case -1:
			minusOne = append(minusOne, day)
		default:
			regularDays = append(regularDays, day)
		}
	}

	// Собираем результат в нужном порядке
	result := regularDays
	if len(minusTwo) > 0 {
		result = append(result, minusTwo...)
	}
	if len(minusOne) > 0 {
		result = append(result, minusOne...)
	}

	return result
}

// parseDays парсит строку с днями месяца в отсортированный слайс.
//
// Формат: "D1,D2,..." где -2 ≤ D ≤ 31 (0 исключен).
// Специальные значения:
//
//	-1 - последний день месяца
//	-2 - предпоследний день месяца
//
// Возвращает ошибку при неверном формате.
func parseDays(s string) ([]int, error) {
	dayStr := strings.Split(s, ",")
	dayInt, err := stringsToInts(dayStr) // [случайный порядок]
	slices.Sort(dayInt)                  //[-1, -2, 15]
	result := arrangeSpecialDays(dayInt) // [15, -2, -1] верный порядок
	return result, err
}

// StringsToInts конвертирует слайс строк в слайс целых чисел с валидацией
func stringsToInts(strs []string) ([]int, error) {
	ints := make([]int, 0, len(strs))
	for _, s := range strs {
		num, err := strconv.Atoi(s)
		if err != nil || num < -2 || num > 32 || num == 0 {
			return nil, errForamt
		}
		ints = append(ints, num)
	}
	return ints, nil
}

// findMonthDay находит следующую дату для месячного правила повтора.
//
// Алгоритм:
// 1. Проверяет доступные месяцы
// 2. Для каждого месяца проверяет указанные дни
// 3. Возвращает первую дату после now
func findMonthDay(now, date time.Time, daysStr string, months ...string) (string, error) {

	month, err := parseMonth(months)
	if err != nil {
		return "", err
	}

	days, err := parseDays(daysStr)
	if err != nil {
		return "", err
	}

	// Начинаем с первого дня следующего месяца, если день начала уже прошел в текущем месяце
	// напрмиер  Текущая дата: 15 января, а дата задачи 10 январи
	// и повтор каждый месяц, мы сразу перейдем на месяц вперед
	if !afterNow(date, now) {
		date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).AddDate(0, 1, 0)
	}

	for {
		currentMonth := int(date.Month())
		// проверяем доступность месяца
		if !month[currentMonth] {
			date = date.AddDate(0, 1, 0)
			continue
		}

		// Проверяем все дни в текущем месяце, начиная с 1, так ка слайс отсортирован
		for _, day := range days {
			var target time.Time
			switch {
			case day == -1:
				target = lastDayOfMonth(date)
			case day == -2:
				last := lastDayOfMonth(date)
				target = last.AddDate(0, 0, -1)
			default:
				// Пропускаем дни, которых нет в месяце (для февраля или 30-дневных)
				lastDay := lastDayOfMonth(date).Day()
				if day > lastDay {
					continue
				}
				target = time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, date.Location())
			}

			if target.After(now) {
				return target.Format(DateFormat), nil
			}
		}

		// Переход к следующему месяцу
		date = date.AddDate(0, 1, 0)
	}
}
