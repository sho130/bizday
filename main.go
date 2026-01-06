package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed holidays.yaml
var holidaysYAML []byte

// 祝日の定義を読み込むための構造体
type HolidayList struct {
	Holidays []string `yaml:"holidays"`
}

func main() {
	// 埋め込み済みの祝日一覧を取得
	holidays, err := loadHolidays()
	if err != nil {
		log.Fatalf("祝日ファイルの読み込みに失敗しました: %v", err)
	}

	// 今日の日付
	today := time.Now()
	// tz, _ := time.LoadLocation("Asia/Tokyo")
	// today := time.Date(2025, 4, 1, 0, 0, 0, 0, tz)

	// 今月の開始日と終了日を取得
	start := beginningOfMonth(today)
	end := endOfMonth(today)

	// 今月の開始日から今日までの営業日数
	businessDaysPassed, err := calcBusinessDaysInRange(start, today, holidays)
	if err != nil {
		log.Fatalf("営業日計算中にエラー: %v", err)
	}

	// 今月の開始日から最終日までの営業日数
	businessDaysTotal, err := calcBusinessDaysInRange(start, end, holidays)
	if err != nil {
		log.Fatalf("営業日計算中にエラー: %v", err)
	}

	// 今日が営業日かどうか
	// isTodayBusinessDay := isBusinessDay(today, holidays)

	// 今日が営業日の場合、経過営業日数のカウントが1日分増えるイメージ
	businessDayIndex := businessDaysPassed
	// ただし calcBusinessDaysInRange は「start~today(含む)」なので、すでに今日をカウント済み
	// → そのままでOK

	// 残り営業日 = 今月全営業日数 - これまでの営業日数
	businessDaysLeft := businessDaysTotal - businessDaysPassed
	// もし今日が営業日であっても、既に calcBusinessDaysInRange に含まれているので
	// ここで -1 する必要はない (残りは start~end のうち today を除いた先の日数になる)

	fmt.Printf("今日は今月の %d 営業日目 です\n", businessDayIndex)
	fmt.Printf("今月の残り営業日は %d 日 です\n", businessDaysLeft)
	fmt.Printf("今月の残り想定稼働時間は %d 時間 です\n", businessDaysLeft*8)
	fmt.Printf("%.1f %% 経過しました\n", float64(businessDayIndex)/float64(businessDaysTotal)*100)
}

// loadHolidays は埋め込み済みの YAML から祝日を読み込み、time.Time のスライスにして返す
func loadHolidays() ([]time.Time, error) {
	if len(holidaysYAML) == 0 {
		return nil, fmt.Errorf("holidays.yaml が埋め込まれていません")
	}

	var holidayList HolidayList
	err := yaml.Unmarshal(holidaysYAML, &holidayList)
	if err != nil {
		return nil, err
	}

	var holidays []time.Time
	for _, holidayStr := range holidayList.Holidays {
		t, err := time.Parse("2006-01-02", holidayStr)
		if err != nil {
			return nil, fmt.Errorf("祝日のパースに失敗: %s", holidayStr)
		}
		holidays = append(holidays, t)
	}
	return holidays, nil
}

// isBusinessDay は土日・祝日を除外した“営業日”かどうかを判定
func isBusinessDay(day time.Time, holidays []time.Time) bool {
	// 土日判定
	if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		return false
	}

	// 祝日判定
	for _, h := range holidays {
		if isSameDay(day, h) {
			return false
		}
	}
	return true
}

// isSameDay は、2つの time.Time が同じ年月日かどうかを判定
func isSameDay(day1, day2 time.Time) bool {
	y1, m1, d1 := day1.Date()
	y2, m2, d2 := day2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// calcBusinessDaysInRange は start~end (両端含む) の営業日数を返す
func calcBusinessDaysInRange(start, end time.Time, holidays []time.Time) (int, error) {
	if end.Before(start) {
		return 0, errors.New("end は start より後の日付を指定してください")
	}

	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if isBusinessDay(d, holidays) {
			count++
		}
	}
	return count, nil
}

// beginningOfMonth は与えられた日付の月初 (xx月1日 0:00:00) を返す
func beginningOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// endOfMonth は与えられた日付の月末 (xx月末日 23:59:59) を返す
func endOfMonth(t time.Time) time.Time {
	// 月初を取得
	firstDayOfMonth := beginningOfMonth(t)
	// 次の月に +1 して日数を -1 すると、当月末日
	nextMonth := firstDayOfMonth.AddDate(0, 1, 0)
	endOfThisMonth := nextMonth.AddDate(0, 0, -1)
	// 23:59:59 に設定
	return time.Date(
		endOfThisMonth.Year(),
		endOfThisMonth.Month(),
		endOfThisMonth.Day(),
		23, 59, 59, 0,
		t.Location(),
	)
}
