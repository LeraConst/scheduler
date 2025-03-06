package repeat

import (
	"fmt"
	"log"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// RulesNextDate задает правила повтроения задач
func RulesNextDate(now time.Time, dateStr string, repeat string) (string, error) {
	// Проверка на корректность даты
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	repeat = strings.Join(strings.Fields(repeat), " ") // лишние пробелы
	// Если repeat пустой — задачи больше нет
	if repeat == "" {
		return "", nil
	}

	// Разделяем правило на слайс строк (буква, цифра)
	rules := strings.Split(repeat, " ")

	switch {
	case rules[0] == "y":
		// Добавляем год, если дата уже >= now
		if !date.Before(now) {
			date = date.AddDate(1, 0, 0)
		} else {
			for date.Before(now) {
				date = date.AddDate(1, 0, 0)
			}
		}

		log.Println("Определили правило y и возвращаем", date.Format("02.01.2006"))
		return date.Format("20060102"), nil

	case rules[0] == "d" && len(rules) == 2:
		days, err := strconv.Atoi(rules[1])
		// Проверка на некорректные данные
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("некорректное правило repeat: %s", repeat)
		}
		// Добавляем дни, пока дата не станет >= now
		if !date.Before(now) {
			date = date.AddDate(0, 0, days)
		} else {
			for date.Before(now) {
				date = date.AddDate(0, 0, days)
			}
		}
		for !date.After(now) {
			date = date.AddDate(0, 0, days)
		}

		log.Printf("Определили правило d %s и возвращаем %v", rules[1], date.Format("02.01.2006"))
		return date.Format("20060102"), nil

	case rules[0] == "w" && len(rules) == 2:
		weekDays := strings.Split(rules[1], ",") // разделяем числа (дни недели)
		var validDays []time.Weekday             // для записи корректных дней недели

		for _, w := range weekDays {
			day, err := strconv.Atoi(w)
			// Проверка на некорректные данные
			if err != nil || day < 1 || day > 7 {
				return "", fmt.Errorf("некорректное правило repeat: %s", repeat)
			}
			validDays = append(validDays, time.Weekday(day%7))
		}

		for {
			if (slices.Contains(validDays, date.Weekday())) && date.After(now) {
				log.Printf("Определили правило w %s и возвращаем %v", rules[1], date.Format("02.01.2006"))
				return date.Format("20060102"), nil
			}
			date = date.AddDate(0, 0, 1)
		}

	case rules[0] == "m" && len(rules) <= 3 && len(rules) >= 2:
		// Обрабатываем правила m <дни месяца> [<месяцы>]
		daysStr := strings.Split(rules[1], ",")
		var validDays []int

		for _, d := range daysStr {
			dTrimmed := strings.TrimLeft(d, "0") // Убираем ведущие нули
			if dTrimmed == "" {
				dTrimmed = "0" // Если число было "0", оно должно остаться "0"
			}
			day, err := strconv.Atoi(dTrimmed)
			if err != nil || day == 0 || day < -2 || day > 31 {
				return "", fmt.Errorf("некорректное правило repeat: %s", repeat)
			}
			validDays = append(validDays, day)
		}

		var validMonths map[int]bool
		if len(rules) == 3 {
			validMonths = make(map[int]bool)
			monthsStr := strings.Split(rules[2], ",")
			for _, m := range monthsStr {
				mTrimmed := strings.TrimLeft(m, "0") // Убираем ведущие нули у месяцев
				if mTrimmed == "" {
					mTrimmed = "0"
				}
				month, err := strconv.Atoi(mTrimmed)
				if err != nil || month < 1 || month > 12 {
					return "", fmt.Errorf("некорректное правило repeat: %s", repeat)
				}
				validMonths[month] = true
			}
		}

		possibleDates := []time.Time{}
		// Если date в прошлом – сдвигаем его на сейчас
		if date.Before(now) {
			date = now
		}
		currentDate := date

		for i := 0; i < 12; i++ { // Проверяем ближайшие 12 месяцев
			lastDay := time.Date(currentDate.Year(), currentDate.Month()+1, 0, 0, 0, 0, 0, currentDate.Location()).Day()

			for _, d := range validDays {
				targetDay := d
				if d == -1 {
					targetDay = lastDay
				} else if d == -2 {
					targetDay = lastDay - 1
				}

				if targetDay > 0 && targetDay <= lastDay {
					candidateDate := time.Date(currentDate.Year(), currentDate.Month(), targetDay, 0, 0, 0, 0, currentDate.Location())

					// Проверяем, что дата в будущем и соответствует допустимым месяцам
					if candidateDate.After(now) && (validMonths == nil || validMonths[int(candidateDate.Month())]) {
						possibleDates = append(possibleDates, candidateDate)
					}
				}
			}
			currentDate = currentDate.AddDate(0, 1, 0) // Переход на следующий месяц
		}

		// Сортируем и выбираем ближайшую подходящую дату
		sort.Slice(possibleDates, func(i, j int) bool { return possibleDates[i].Before(possibleDates[j]) })

		if len(possibleDates) > 0 {
			log.Printf("Определили правило %s и возвращаем %v", repeat, possibleDates[0].Format("02.01.2006"))
			return possibleDates[0].Format("20060102"), nil
		}

		return "", fmt.Errorf("не удалось найти подходящую дату для repeat: %s", repeat)

	default:
		return "", fmt.Errorf("неизвестное правило repeat: %s", repeat)
	}

}
