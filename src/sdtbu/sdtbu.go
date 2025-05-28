package sdtbu

import (
	"CourseTool/des" // 假設 des 套件用於加密
	"bytes"
	"encoding/json" // 導入 json 套件，用於處理 JSON 數據
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url" // 導入 url 套件，用於構建表單數據
	"sort"
	"strings"
	"time" // 導入 time 套件，用於處理時間
	"unicode/utf8"

	"golang.org/x/net/html"
)

// ANSI 顏色代碼常量
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Purple = "\033[35m" // 新增紫色
)

// SemesterStartDate 定義學期第一週的第一天
const SemesterStartDate = "2025.02.24" // 您可以根據實際情況修改此日期

// Init 函數，用於初始化
func Init() {
	fmt.Println(Green + "SDTBU: Initializing..." + Reset)
}

// LoginParams 結構體用於儲存從登入頁面提取的參數
type LoginParams struct {
	Lt        string
	Execution string
	EventId   string
}

// ClientSession 結構體用於儲存 HTTP 客戶端和 cookie jar，以便在不同函數間共用
type ClientSession struct {
	Client *http.Client
	Jar    *cookiejar.Jar
	// 您可以在這裡添加其他會話相關的資訊，例如 User-Agent
	UserAgent string
	reqURL    string // 用於存儲請求的 URL
	// 您可以在這裡添加其他需要的字段，例如請求后獲得的部分信息
	CalssListUserInfoString string // 用於存儲課程列表的字符串
	ClassListbyTimeString   string // 用於存儲本周課程時間列表的字符串
}

// ClassSchedule 結構體定義了每節課的開始和結束時間
type ClassSchedule struct {
	Lesson int    // 節次，例如 1 代表第一節課
	Start  string // 開始時間，例如 "08:00"
	End    string // 結束時間，例如 "08:45"
}

// 全局課程時間表，您可以根據實際情況修改
var classTimetable = []ClassSchedule{
	{Lesson: 1, Start: "08:00", End: "09:30"},
	{Lesson: 2, Start: "08:45", End: "09:30"},
	{Lesson: 3, Start: "09:50", End: "11:20"},
	{Lesson: 4, Start: "10:35", End: "11:20"},
	{Lesson: 5, Start: "14:00", End: "15:30"},
	{Lesson: 6, Start: "14:45", End: "15:30"},
	{Lesson: 7, Start: "15:50", End: "17:20"},
	{Lesson: 8, Start: "16:35", End: "17:20"},
	{Lesson: 9, Start: "19:00", End: "20:30"},
	{Lesson: 10, Start: "19:45", End: "20:30"},
	{Lesson: 11, Start: "20:50", End: "21:35"},
}

// GetFormattedClassTime 根據節次返回格式化的課程開始和結束時間。
// lessonNumber: 課程的節次。
// 返回格式如 "08:00-08:45" 的時間字符串，如果找不到則返回錯誤。
func GetFormattedClassTime(lessonNumber int) (string, error) {
	for _, schedule := range classTimetable {
		if schedule.Lesson == lessonNumber {
			return fmt.Sprintf("%s-%s", schedule.Start, schedule.End), nil
		}
	}
	return "", fmt.Errorf(Yellow+"SDTBU: 未找到節次 %d 對應的時間表資訊。"+Reset, lessonNumber)
}

// goWeekdayToApiSkxq 將 Go 的 time.Weekday 轉換為系統使用的 SKXQ (1-7, 1=Mon, 7=Sun)
func goWeekdayToApiSkxq(wd time.Weekday) int {
	if wd == time.Sunday {
		return 7 // 系統中星期日是 7
	}
	return int(wd) // Monday is 1, ..., Saturday is 6
}

// extractIntFromClassMap 安全地從 map[string]interface{} 中提取指定鍵的整數值。
// 處理 JSON 數字可能解析為 float64 的情況。
func extractIntFromClassMap(classMap map[string]interface{}, key string) (int, bool) {
	val, ok := classMap[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}

// NewClientSession 函數用於創建並初始化一個新的 ClientSession
func NewClientSession() (*ClientSession, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf(Red+"SDTBU: Failed to create cookie jar: %v"+Reset, err)
	}

	client := &http.Client{
		Jar: jar, // 為客戶端設定 cookie jar
	}

	return &ClientSession{
		Client:    client,
		Jar:       jar,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0",
	}, nil
}

// NextClass 函數用於與當前時間對比並返回下一節課程的資訊。
// classListJSON 是一個包含課程資訊的列表，每個課程資訊是一個 map。
// 假設傳入的 classListJSON 已經是排序好的，並且在篩選出今天的課程後，
// 這些課程也保持了按節次排序的特性。
func (cs *ClientSession) NextClass(classListJSON []map[string]interface{}) (map[string]interface{}, error) {
	if len(classListJSON) == 0 {
		return nil, fmt.Errorf(Yellow + "SDTBU: 沒有課程資訊可供判斷下一節課。" + Reset)
	}

	// 獲取當前時間和星期
	now := time.Now()
	currentTimeStr := now.Format("15:04") // 格式化為 HH:MM
	currentSystemSkxq := goWeekdayToApiSkxq(now.Weekday())

	// fmt.Printf(Green+"SDTBU: 當前時間: %s, 今天星期 (系統): %d\n"+Reset, currentTimeStr, currentSystemSkxq)

	// --- 第一部分: 檢查今天的下一節課 ---
	var todayClasses []map[string]interface{}
	for _, class := range classListJSON {
		skxq, ok := extractIntFromClassMap(class, "SKXQ")
		if !ok {
			// fmt.Printf(Red+"警告: 課程 SKXQ 資訊無效或缺失: %v\n"+Reset, class)
			continue
		}
		if skxq == currentSystemSkxq {
			todayClasses = append(todayClasses, class)
		}
	}

	if len(todayClasses) > 0 {
		for _, class := range todayClasses {
			skjc, ok := extractIntFromClassMap(class, "SKJC")
			if !ok {
				// fmt.Printf(Yellow+"警告: 課程 SKJC 資訊無效或缺失: %v\n"+Reset, class)
				continue
			}

			var classStart, classEnd string
			for _, schedule := range classTimetable {
				if schedule.Lesson == skjc {
					classStart = schedule.Start
					classEnd = schedule.End
					break
				}
			}

			if classStart == "" {
				// fmt.Printf(Yellow+"警告: 未找到今天課程 (節次 %d) 的時間表資訊。\n"+Reset, skjc)
				continue
			}

			layout := "15:04"
			currentT, errCurrent := time.Parse(layout, currentTimeStr)
			classStartT, errStart := time.Parse(layout, classStart)
			classEndT, errEnd := time.Parse(layout, classEnd)

			if errCurrent != nil || errStart != nil || errEnd != nil {
				// fmt.Printf(Red+"錯誤: 解析時間時出錯 for class %v - currentT: %v, classStartT: %v, classEndT: %v\n"+Reset, class, errCurrent, errStart, errEnd)
				continue // 跳過此課程，如果時間解析失敗
			}

			if currentT.Before(classStartT) || (currentT.After(classStartT) && currentT.Before(classEndT)) {
				_, _ = class["KCMC"].(string) // kcmcStr was used in a commented-out fmt.Printf
				// fmt.Printf(Green+"SDTBU: 今天的下一節課是: %s (星期: %d, 節次: %d, 時間: %s-%s)\n"+Reset, kcmcStr, currentSystemSkxq, skjc, classStart, classEnd)
				return class, nil // 直接返回今天的下一節課
			}
		}
	}

	// --- 第二部分: 如果今天沒有更多課程，查找明天的第一節課 ---
	// fmt.Println(Yellow + "SDTBU: 今天沒有更多課程了，正在查找明天的課程..." + Reset)

	goTomorrowWd := time.Weekday((int(now.Weekday()) + 1) % 7) // 計算明天的 Go Weekday
	tomorrowSystemSkxq := goWeekdayToApiSkxq(goTomorrowWd)     // 轉換為系統的 SKXQ

	// fmt.Printf(Green+"SDTBU: 查找明天 (星期 %d) 的課程...\n"+Reset, tomorrowSystemSkxq)

	var tomorrowClasses []map[string]interface{}
	for _, class := range classListJSON { // classListJSON 已經是排序好的
		skxq, ok := extractIntFromClassMap(class, "SKXQ")
		if !ok {
			continue
		}
		if skxq == tomorrowSystemSkxq {
			tomorrowClasses = append(tomorrowClasses, class)
		}
	}

	if len(tomorrowClasses) > 0 {
		firstClassTomorrow := tomorrowClasses[0] // 由於已排序，第一個就是明天的第一節課

		skjcTomorrow, ok := extractIntFromClassMap(firstClassTomorrow, "SKJC")
		if !ok {
			return nil, fmt.Errorf(Yellow + "SDTBU: 明天第一節課的節次(SKJC)資訊無效。" + Reset)
		}

		var classStartTomorrow string
		for _, schedule := range classTimetable {
			if schedule.Lesson == skjcTomorrow {
				classStartTomorrow = schedule.Start
				// classEndTomorrow = schedule.End // classEndTomorrow was used in a commented-out fmt.Printf
				break
			}
		}

		if classStartTomorrow == "" {
			return nil, fmt.Errorf(Yellow+"SDTBU: 未找到明天第一節課 (節次 %d) 的時間表資訊。"+Reset, skjcTomorrow)
		}

		_, _ = firstClassTomorrow["KCMC"].(string) // kcmcTomorrowStr was used in a commented-out fmt.Printf
		// fmt.Printf(Green+"SDTBU: 明天的首節課程是: %s (星期: %d, 節次: %d, 時間: %s-%s)\n"+Reset,
		// 	kcmcTomorrowStr, tomorrowSystemSkxq, skjcTomorrow, classStartTomorrow, classEndTomorrow)

		// 創建副本以添加 Remark
		resultClass := make(map[string]interface{})
		for k, v := range firstClassTomorrow {
			resultClass[k] = v
		}
		resultClass["Remark"] = "明天的首節課程" // 添加說明

		return resultClass, nil
	}

	// 如果今天和明天都沒有課程
	return nil, fmt.Errorf(Yellow + "SDTBU: 今天和明天都沒有課程了。" + Reset)
}

// SortClass 根據課程列表對課程進行排序，並返回排序後的課程列表和一個訊息字符串。
// classListJSON 是一個包含課程資訊的列表，每個課程資訊是一個 map。
// "SKXQ" (上課星期) 和 "SKJC" (上課節次) 預期為整數型，但會處理可能來自 JSON 的 float64 類型。
// 排序規則：首先按 SKXQ (上課星期) 升序排序，然後按 SKJC (上課節次) 升序排序。
func (cs *ClientSession) SortClass(classListJSON []map[string]interface{}) ([]map[string]interface{}, string) {
	if len(classListJSON) == 0 {
		return nil, "沒有課程資訊"
	}

	// 使用 sort.Slice 對 classListJSON 進行排序
	sort.Slice(classListJSON, func(i, j int) bool {
		// 獲取第 i 個課程的 SKXQ 和 SKJC，並進行類型斷言
		// 由於 JSON 解析可能將數字解析為 float64，我們首先斷言為 float64，然後轉換為 int
		skxq_i_float, ok_i_skxq := classListJSON[i]["SKXQ"].(float64)
		skxq_i := 0
		if ok_i_skxq {
			skxq_i = int(skxq_i_float)
		} else {
			// 錯誤處理：如果類型不匹配，打印錯誤並假設一個值以避免崩潰
			fmt.Printf(Red+"錯誤: classListJSON[%d][\"SKXQ\"] 不是 float64，實際類型為 %T。\n"+Reset, i, classListJSON[i]["SKXQ"])
		}

		skjc_i_float, ok_i_skjc := classListJSON[i]["SKJC"].(float64)
		skjc_i := 0
		if ok_i_skjc {
			skjc_i = int(skjc_i_float)
		} else {
			fmt.Printf(Red+"錯誤: classListJSON[%d][\"SKJC\"] 不是 float64，實際類型為 %T。\n"+Reset, i, classListJSON[i]["SKJC"])
		}

		// 獲取第 j 個課程的 SKXQ 和 SKJC，並進行類型斷言
		skxq_j_float, ok_j_skxq := classListJSON[j]["SKXQ"].(float64)
		skxq_j := 0
		if ok_j_skxq {
			skxq_j = int(skxq_j_float)
		} else {
			fmt.Printf(Red+"錯誤: classListJSON[%d][\"SKXQ\"] 不是 float64，實際類型為 %T。\n"+Reset, j, classListJSON[j]["SKXQ"])
		}

		skjc_j_float, ok_j_skjc := classListJSON[j]["SKJC"].(float64)
		skjc_j := 0
		if ok_j_skjc {
			skjc_j = int(skjc_j_float)
		} else {
			fmt.Printf(Red+"錯誤: classListJSON[%d][\"SKJC\"] 不是 float64，實際類型為 %T。\n"+Reset, j, classListJSON[j]["SKJC"])
		}

		// 首先比較星期 (SKXQ)
		if skxq_i != skxq_j {
			return skxq_i < skxq_j
		}
		// 如果星期相同，則比較節次 (SKJC)
		return skjc_i < skjc_j
	})

	// 打印排序後的 classList 內容 (用於調試)
	// fmt.Println(Green + "已排序的 classList 內容:" + Reset)
	// for i, class := range classListJSON {
	// 	// 確保 KCMC 是字符串類型
	// 	kcmc, ok := class["KCMC"].(string)
	// 	if !ok {
	// 		kcmc = "未知課程名稱" // 如果 KCMC 不是字符串，使用預設值
	// 	}
	// 	// 由於 SKXQ 和 SKJC 可能仍然是 float64，我們再次嘗試斷言為 float64 並轉換為 int 進行顯示
	// 	skxq_float, _ := class["SKXQ"].(float64)
	// 	skxq := int(skxq_float)
	// 	skjc_float, _ := class["SKJC"].(float64)
	// 	skjc := int(skjc_float)
	// 	fmt.Printf("課程 %d: %s (星期: %d, 節次: %d)\n", i+1, kcmc, skxq, skjc)
	// }

	return classListJSON, "課程列表已排序，但沒有可顯示的第一節課。"
}

// ParseClassList 函數用於解析 GetClassbyTime 的 JSON 響應結構
// 解析 GetClassbyTime 的 JSON 響應結構
// 這裡假設 classList 是一個包含課程資訊的列表，每個課程資訊是一個物件
func (cs *ClientSession) ParseClassList(jsonData string) ([]map[string]interface{}, error) {
	// 聲明一個 Go 切片變量，用於存儲解析後的 classList 內容
	var classList []map[string]interface{}
	// 使用 json.Unmarshal 將字符串變量解析到 Go 切片中
	err := json.Unmarshal([]byte(jsonData), &classList)
	if err != nil {
		fmt.Println(Red+"Error unmarshalling classList string:", err, Reset)
		return nil, fmt.Errorf(Red+"SDTBU: Error unmarshalling classList string: %v"+Reset, err)
	}
	return classList, nil
}

// GetClassbyTime 函數用於發送 POST 請求獲取用戶的本周課程資訊
func (cs *ClientSession) GetClassbyTime() error {
	fmt.Println(Blue + "SDTBU: Fetching class information by time..." + Reset)

	// 請求 URL
	var requestURL string

	//判斷是否使用webVPN
	if cs.reqURL != "https://zhss.sdtbu.edu.cn/tp_up/" {
		requestURL = cs.reqURL + "up/widgets/getClassbyTime?vpn-12-o2-zhss.sdtbu.edu.cn"
	} else {
		requestURL = cs.reqURL + "up/widgets/getClassbyTime"
	}

	// 聲明一個 Go 切片變量，用於存儲解析後的 classList 內容
	// 這裡我們將 classList 內的物件鍵值改為 interface{}，以適應可能包含數字或其他類型的 JSON 值
	var classListContent []map[string]interface{}

	// 使用 json.Unmarshal 將字符串變量解析到 Go 切片中
	err := json.Unmarshal([]byte(cs.CalssListUserInfoString), &classListContent)
	if err != nil {
		fmt.Println(Red+"Error unmarshalling classList string:", err, Reset)
		return fmt.Errorf(Red+"SDTBU: Error unmarshalling classListUserInfoString: %v"+Reset, err)
	}

	// 計算當前教學週
	startDate, err := time.Parse("2006.01.02", SemesterStartDate)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error parsing semester start date: %v"+Reset, err)
	}
	now := time.Now()

	// 確保只比較日期部分，忽略時間
	daysSinceStart := int(now.Truncate(24*time.Hour).Sub(startDate.Truncate(24*time.Hour)).Hours() / 24)

	currentLearnWeek := 1 // 默認為第一周
	if daysSinceStart >= 0 {
		currentLearnWeek = (daysSinceStart / 7) + 1
	} else {
		// 如果當前日期早於開學日期，可以根據需求處理，這裡默認為第1周或報錯
		fmt.Println(Yellow + "SDTBU: Current date is before the semester start date. Defaulting learnWeek to 1." + Reset)
	}

	//測試工具，打印當前周
	// fmt.Printf(Green+"SDTBU: Current learn week: %d\n"+Reset, currentLearnWeek)

	// 構建請求體數據
	requestBody := map[string]interface{}{
		"schoolYear": "2024-2025",
		"semester":   "2",
		// "learnWeek":  "1",
		"learnWeek": fmt.Sprintf("%d", currentLearnWeek), // 使用計算出的當前周
		"classList": classListContent,                    // 使用之前獲取的課程列表
	}

	// 將請求體數據編碼為 JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error marshalling request body to JSON: %v"+Reset, err)
	}

	// 創建 POST 請求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating POST request for GetClassbyTime: %v"+Reset, err)
	}

	// 設定請求標頭
	req.Header.Set("Content-Type", "application/json") // 設定內容類型為 JSON
	req.Header.Set("User-Agent", cs.UserAgent)         // 設定 User-Agent

	// 發送請求
	resp, err := cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error sending POST request to GetClassbyTime: %v"+Reset, err)
	}
	defer resp.Body.Close() // 確保響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: POST request to %s status: %s\n"+Reset, requestURL, resp.Status)

	// 讀取響應主體
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error reading GetClassbyTime response body: %v"+Reset, err)
	}

	// 記錄Class内容
	cs.ClassListbyTimeString = string(bodyBytes)

	// fmt.Println(Green + "SDTBU: Class information by time fetched successfully." + Reset)
	// fmt.Println(Cyan + "SDTBU: Class List by Time String: " + Reset + cs.ClassListbyTimeString)

	return nil
}

// GetClassbyUserInfo 函數用於發送 POST 請求獲取用戶的課程資訊
func (cs *ClientSession) GetClassbyUserInfo() error {
	fmt.Println(Blue + "SDTBU: Fetching class information..." + Reset)

	// 請求 URL
	var requestURL string

	//判斷是否使用webVPN
	if cs.reqURL != "https://zhss.sdtbu.edu.cn/tp_up/" {
		requestURL = cs.reqURL + "up/widgets/getClassbyUserInfo?vpn-12-o2-zhss.sdtbu.edu.cn"
	} else {
		requestURL = cs.reqURL + "up/widgets/getClassbyUserInfo"
	}

	// 構建請求體數據
	requestBody := map[string]string{
		"schoolYear": "2024-2025",
		"semester":   "2",
		"learnWeek":  "14",
	}

	// 將請求體數據編碼為 JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error marshalling request body to JSON: %v"+Reset, err)
	}

	// 創建 POST 請求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating POST request for getClassbyUserInfo: %v"+Reset, err)
	}

	// 設定請求標頭
	req.Header.Set("Content-Type", "application/json") // 設定內容類型為 JSON
	req.Header.Set("User-Agent", cs.UserAgent)         // 設定 User-Agent

	// 發送請求
	resp, err := cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error sending POST request to getClassbyUserInfo: %v"+Reset, err)
	}
	defer resp.Body.Close() // 確保響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: POST request to %s status: %s\n"+Reset, requestURL, resp.Status)

	// 讀取響應主體
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error reading getClassbyUserInfo response body: %v"+Reset, err)
	}

	// 記錄Class内容
	cs.CalssListUserInfoString = string(bodyBytes)

	return nil
}

// Login 函數，作為 ClientSession 的方法，用於登入 SDTBU 系統
// 該函數模擬了瀏覽器行為，先進行 GET 請求獲取登入頁面，
// 然後解析頁面以提取必要的參數（例如 lt, execution, _eventId 值），
// 最後構建 POST 請求並發送登入資訊。
func (cs *ClientSession) Login(username, password string) error {
	fmt.Printf(Green+"SDTBU: Logging in with username: %s\n"+Reset, username)

	// --- 1. 執行 GET 請求以獲取登入頁面和相關參數 ---
	// 宣告 req 變數，以便在後續的 GET 和 POST 請求中重複使用
	var req *http.Request
	var err error
	var resp *http.Response

	getReqURL := "https://zhss.sdtbu.edu.cn/tp_up/"
	req, err = http.NewRequest("GET", getReqURL, nil)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating GET request: %v"+Reset, err)
	}
	req.Header.Set("User-Agent", cs.UserAgent)

	// 使用共用的客戶端傳送 GET 請求
	resp, err = cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error making GET request to %s: %v"+Reset, getReqURL, err)
	}
	defer resp.Body.Close() // 確保 GET 請求的響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: GET request to %s status: %s\n"+Reset, getReqURL, resp.Status)

	// 讀取響應主體以提取登入表單的 HTML 內容
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error reading GET response body: %v"+Reset, err)
	}
	htmlBody := string(bodyBytes)

	// POST 請求的 URL 將是提供登入表單的 URL。
	// 在 GET 請求（以及任何重定向）之後，這在 resp.Request.URL 中可用。
	postTargetURL := resp.Request.URL.String()
	fmt.Println(Yellow+"SDTBU: Login form URL (target for POST): "+Reset, postTargetURL)

	// 2. 從 HTML 內容中提取登入參數 (lt, execution, _eventId)
	// 這些參數通常是隱藏欄位，用於維持會話狀態或防止 CSRF 攻擊。
	loginParams := ExtractLoginParameters(htmlBody)
	if loginParams.Lt == "" || loginParams.Execution == "" || loginParams.EventId == "" {
		return fmt.Errorf(Red+"SDTBU: Failed to extract all required login parameters. Lt: '%s', Execution: '%s', EventId: '%s'"+Reset,
			loginParams.Lt, loginParams.Execution, loginParams.EventId)
	}
	//fmt.Println(Yellow+"SDTBU: Extracted login parameters:", loginParams, Reset)

	// 3. 準備 POST 請求的表單資料
	// 根據原代碼邏輯，rsa 值由用戶名、密碼和 lt 值拼接而成，然後進行加密。
	rsa := fmt.Sprintf("%s%s%s", username, password, loginParams.Lt)
	//fmt.Println("SDTBU: Prepared RSA value (before encryption):", rsa)

	// 加密 RSA 值，這裡假設 des.StrEnc 函數可用於加密
	encryptedRSA := des.StrEnc(rsa, "1", "2", "3")
	//fmt.Println("SDTBU: Encrypted RSA value:", encryptedRSA)

	// 計算用戶名和密碼的長度，用於 POST 請求中的 ul 和 pl 參數
	ul := utf8.RuneCountInString(username)
	pl := utf8.RuneCountInString(password)

	// 構建 POST 請求的資料，使用 url.Values 更規範地處理表單數據
	formData := url.Values{}
	formData.Set("rsa", encryptedRSA)
	formData.Set("ul", fmt.Sprintf("%d", ul))
	formData.Set("pl", fmt.Sprintf("%d", pl))
	formData.Set("lt", loginParams.Lt)
	formData.Set("execution", loginParams.Execution)
	formData.Set("_eventId", loginParams.EventId)

	postDataString := formData.Encode() // 將表單數據編碼為 URL 查詢字符串格式
	//fmt.Println("SDTBU: POST data:", postDataString)

	// 將 POST 資料字串轉換為 io.Reader
	postDataReader := strings.NewReader(postDataString)

	// --- 4. 執行 POST 請求以提交登入資訊 ---
	// 這裡重新賦值 req，而不是重新宣告
	req, err = http.NewRequest("POST", postTargetURL, postDataReader)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating POST request: %v"+Reset, err)
	}

	// 設定請求標頭
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", cs.UserAgent) // 保持 User-Agent 一致

	// 傳送 POST 請求
	resp, err = cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error sending POST request: %v"+Reset, err)
	}
	defer resp.Body.Close() // 確保 POST 響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: POST request to %s status: %s\n"+Reset, req.URL, resp.Status)
	fmt.Println(Yellow+"SDTBU: Current URL after POST:", resp.Request.URL.String()+Reset) // 列印請求的最終 URL

	cs.reqURL = resp.Request.URL.String() // 儲存最終請求的 URL

	// 在這裡可以添加進一步的邏輯來檢查登入是否成功，
	// 例如：檢查響應狀態碼、重定向的 URL、或響應主體內容。
	// 如果登入成功，通常會重定向到一個受保護的頁面。

	// 請求用戶主頁或儀表板頁面
	// 這裡重新賦值 req，而不是重新宣告
	req, err = http.NewRequest("GET", resp.Request.URL.String()+"view?m=up", nil)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating GET request for dashboard: %v"+Reset, err)
	}
	req.Header.Set("User-Agent", cs.UserAgent) // 保持 User-Agent 一致
	resp, err = cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error fetching dashboard page: %v"+Reset, err)
	}
	defer resp.Body.Close()

	return nil
}

// ExtractLoginParameters 從 HTML 內容中提取登入參數 (lt, execution, _eventId) 的值
// 該函數會遍歷 HTML 節點，查找具有特定 id 或 name 屬性的 input 標籤，
// 並提取其 value 值。
func ExtractLoginParameters(htmlbody string) LoginParams {
	doc, err := html.Parse(strings.NewReader(htmlbody))
	if err != nil {
		log.Printf(Red+"SDTBU: Error parsing HTML: %v"+Reset, err)
		return LoginParams{} // 返回一個空的 LoginParams
	}

	params := LoginParams{}
	var findParams func(*html.Node)
	findParams = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var id, name, value string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "id":
					id = attr.Val
				case "name":
					name = attr.Val
				case "value":
					value = attr.Val
				}
			}

			// 根據 id 或 name 提取對應的參數值
			if id == "lt" {
				params.Lt = value
			}
			if name == "execution" {
				params.Execution = value
			}
			if name == "_eventId" {
				params.EventId = value
			}
		}
		// 遞歸遍歷子節點
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findParams(c)
			// 如果所有參數都已找到，則停止遍歷，提高效率
			if params.Lt != "" && params.Execution != "" && params.EventId != "" {
				return
			}
		}
	}

	findParams(doc)
	return params
}
