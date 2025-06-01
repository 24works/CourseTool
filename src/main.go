package main

import (
	ASNIColor "CourseTool/asnicolor"
	_ "CourseTool/configloader" // Import for side effect: load .env
	"CourseTool/sdtbu"
	"CourseTool/wxpush"
	"bufio" // 用於讀取用戶輸入
	"fmt"
	"log"
	"os"
	"sort"    // 用於排序時間點
	"strconv" // 用於字串轉數字
	"strings" // 用於字串處理
	"sync"    // 新增：用於併發控制 (互斥鎖)
	"time"    // 用於時間相關操作
)

// PushTime 結構體用於儲存推送時間的小時和分鐘
type PushTime struct {
	Hour   int
	Minute int
}

// SchedulerStatus 結構體用於儲存排程器的當前狀態
type SchedulerStatus struct {
	mu            sync.Mutex    // 保護以下字段的互斥鎖
	NextPushTime  time.Time     // 下一次預計推送的時間
	SleepDuration time.Duration // 距離下一次推送的剩餘時間 (此字段在 runScheduler 中更新，但實時顯示時會重新計算)
	IsRunning     bool          // 排程器是否正在運行
}

// globalSchedulerStatus 是排程器狀態的全局實例
var globalSchedulerStatus = &SchedulerStatus{}

// printBanner 打印應用程式的啟動橫幅
func printBanner() {
	fmt.Println(ASNIColor.BrightCyan + `
=============================================================
   _____                             _______             _ 
  / ____|                           |__   __|           | |
 | |      ___   _   _  _ __  ___   ___ | |  ___    ___  | |
 | |     / _ \ | | | || '__|/ __| / _ \| | / _ \  / _ \ | |
 | |____| (_) || |_| || |   \__ \|  __/| || (_) || (_) || |
  \_____|\___/  \__,_||_|   |___/ \___||_| \___/  \___/ |_|
                                                           
=============================================================
作者：Richard Miku
版本：v1.0.0
說明：基於Golang的課程工具，提供課程提醒功能。
使用：請確保已正確配置環境變數，然後運行此程序。
網址：https://www.ric.moe
GitHub：https://github.com/RichardMiku/CourseTool
=============================================================
	` + ASNIColor.Reset)
}

// initializeSession 初始化 SDTBU 客戶端會話並執行登入
func initializeSession() (*sdtbu.ClientSession, error) {
	sdtbu.Init() // 初始化您的套件

	session, err := sdtbu.NewClientSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create client session: %v", err)
	}

	username := os.Getenv("SDTBU_USERNAME")
	password := os.Getenv("SDTBU_PASSWORD")

	if username == "" || password == "" {
		return nil, fmt.Errorf(ASNIColor.Red + "錯誤: 環境變數 SDTBU_USERNAME 或 SDTBU_PASSWORD 未設定。" + ASNIColor.Reset)
	}

	err = session.Login(username, password)
	if err != nil {
		return nil, fmt.Errorf(ASNIColor.Red+"登入失敗: %v"+ASNIColor.Reset, err)
	}

	return session, nil
}

// fetchAndProcessClassData 獲取並處理課程數據
func fetchAndProcessClassData(session *sdtbu.ClientSession) (map[string]interface{}, error) {
	// 注意：session 對象在多個 Goroutine 中被訪問。
	// 如果 sdtbu.ClientSession 的方法（如 GetClassbyUserInfo, GetClassbyTime, ParseClassList, SortClass, NextClass）
	// 修改了其內部狀態且不是併發安全的，則需要在此處或 sdtbu 內部添加互斥鎖。
	// 目前假設這些方法是讀取為主或內部已處理併發。
	session.GetClassbyUserInfo()
	err := session.GetClassbyTime()
	if err != nil {
		return nil, fmt.Errorf(ASNIColor.Red+"獲取本周課程失敗: %v"+ASNIColor.Reset, err)
	}

	classList, err := session.ParseClassList(session.ClassListbyTimeString)
	if err != nil {
		return nil, fmt.Errorf(ASNIColor.Red+"解析課程列表失敗: %v"+ASNIColor.Reset, err)
	}

	sortedClassList, sortMsg := session.SortClass(classList)
	if len(sortedClassList) == 0 {
		log.Println(ASNIColor.Yellow + "沒有課程可供排序或顯示: " + sortMsg + ASNIColor.Reset)
		return nil, nil // 返回 nil 表示沒有課程，但不是錯誤
	}

	nextClassInfo, err := session.NextClass(sortedClassList)
	if err != nil {
		log.Printf(ASNIColor.Yellow+"獲取下一節課失敗: %v"+ASNIColor.Reset, err)
		return nil, nil // 返回 nil 表示沒有下一節課，但不是錯誤
	}

	if nextClassInfo == nil {
		log.Println(ASNIColor.Yellow + "今天沒有更多課程了，或未找到下一節課資訊。" + ASNIColor.Reset)
		return nil, nil // 返回 nil 表示沒有下一節課
	}

	return nextClassInfo, nil
}

// extractClassInfo 從課程資訊 map 中安全地提取數據
func extractClassInfo(classInfo map[string]interface{}) (courseName, teacherName, location, timeNumber string) {
	var ok bool

	courseName, ok = classInfo["KCMC"].(string)
	if !ok {
		log.Println(ASNIColor.Yellow + "警告: 課程名稱 (KCMC) 未在 CLINFO 中找到或其類型非字符串。" + ASNIColor.Reset)
		courseName = "未知課程"
	}

	teacherName, ok = classInfo["JSXM"].(string)
	if !ok {
		teacherName, ok = classInfo["JSMC"].(string) // 嘗試備用鍵
		if !ok {
			log.Println(ASNIColor.Yellow + "警告: 教師姓名 (JSXM/JSMC) 未在 CLINFO 中找到或其類型非字符串。" + ASNIColor.Reset)
			teacherName = "未知教師"
		}
	}

	location, ok = classInfo["JXDD"].(string)
	if !ok {
		location, ok = classInfo["JASMC"].(string) // 嘗試備用鍵
		if !ok {
			log.Println(ASNIColor.Yellow + "警告: 上課地點 (JKDD/JASMC) 未在 CLINFO 中找到或其類型非字符串。" + ASNIColor.Reset)
			location = "未知地點"
		}
	}

	var skjcVal interface{}
	skjcVal, ok = classInfo["SKJC"]
	if !ok {
		log.Println(ASNIColor.Yellow + "警告: 上課節次 (SKJC) 未在 CLINFO 中找到。" + ASNIColor.Reset)
		timeNumber = "未知時間"
	} else {
		skjcFloat, ok := skjcVal.(float64)
		if !ok {
			log.Printf(ASNIColor.Yellow+"警告: 上課節次 (SKJC) 類型不是 float64，實際為 %T。"+ASNIColor.Reset, skjcVal)
			timeNumber = "未知時間"
		} else {
			var err error
			timeNumber, err = sdtbu.GetFormattedClassTime(int(skjcFloat))
			if err != nil {
				log.Printf(ASNIColor.Yellow+"警告: 獲取格式化課程時間失敗: %v"+ASNIColor.Reset, err)
				timeNumber = "未知時間"
			}
		}
	}
	return
}

// sendWxPushNotification 檢查環境變數並發送微信推送
func sendWxPushNotification(courseName, teacherName, location, timeNumber string) {
	wxAppID := os.Getenv("WXPUSH_APP_ID")
	wxAppSecret := os.Getenv("WXPUSH_APP_SECRET")
	wxToUser := os.Getenv("WXPUSH_OPEN_ID")
	wxTemplateID := os.Getenv("WXPUSH_COURSE_TEMPLATE_ID")

	if wxAppID == "" || wxAppSecret == "" || wxToUser == "" || wxTemplateID == "" {
		log.Println(ASNIColor.Yellow + "警告: 微信推送所需的一個或多個環境變數 (WXPUSH_APP_ID, WXPUSH_APP_SECRET, WXPUSH_OPEN_ID, WXPUSH_COURSE_TEMPLATE_ID) 未設定。將跳過微信推送功能。" + ASNIColor.Reset)
		fmt.Println(ASNIColor.BrightYellow + "下一節課程資訊：" + ASNIColor.Reset)
		fmt.Printf("課程名稱: %s\n", courseName)
		fmt.Printf("教師姓名: %s\n", teacherName)
		fmt.Printf("上課地點: %s\n", location)
		fmt.Printf("上課節次: %s\n", timeNumber)
		fmt.Printf("額外備註: %s\n", "N/A")
		return
	}

	accessToken, err := wxpush.GetAccessToken()
	if err != nil {
		// 這裡改為 log.Printf 而不是 log.Fatalf，以便排程器可以繼續運行
		log.Printf(ASNIColor.Red+"錯誤: 獲取微信 Access Token 失敗: %v"+ASNIColor.Reset, err)
		return // 如果獲取 Access Token 失敗，則不繼續發送
	}

	courseData := wxpush.CourseReminderData{
		CourseName:     courseName,
		TeacherName:    teacherName,
		CourseLocation: location,
		TimeNumber:     timeNumber,
		Note:           "記得帶上好心情去上課哦！",
	}

	err = wxpush.SendCourseReminder(accessToken, courseData)
	if err != nil {
		fmt.Printf(ASNIColor.Red+"發送課程提醒失敗: %v"+ASNIColor.Reset+"\n", err)
	} else {
		fmt.Println(ASNIColor.BrightGreen + "課程提醒已成功發送！" + ASNIColor.Reset)
	}
}

// parsePushTimeTable 從環境變數中讀取時間表並解析為 PushTime 結構體切片
func parsePushTimeTable() ([]PushTime, error) {
	timeTableStr := os.Getenv("PUSH_TIME_TABLE")
	if timeTableStr == "" {
		return []PushTime{}, nil // 如果未設定，返回空切片
	}

	timeStrings := strings.Split(timeTableStr, "|")
	var pushTimes []PushTime

	for _, ts := range timeStrings {
		parts := strings.Split(ts, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("無效的時間格式 '%s'。預期格式為 HH:MM", ts)
		}
		hour, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("無效的小時數 '%s': %v", parts[0], err)
		}
		minute, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("無效的分鐘數 '%s': %v", parts[1], err)
		}

		if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
			return nil, fmt.Errorf("時間 '%s' 超出有效範圍 (00:00-23:59)", ts)
		}
		pushTimes = append(pushTimes, PushTime{Hour: hour, Minute: minute})
	}

	// 排序推送時間，以便於查找下一個時間點
	sort.Slice(pushTimes, func(i, j int) bool {
		if pushTimes[i].Hour != pushTimes[j].Hour {
			return pushTimes[i].Hour < pushTimes[j].Hour
		}
		return pushTimes[i].Minute < pushTimes[j].Minute
	})

	return pushTimes, nil
}

// runScheduler 負責排程並觸發消息推送
func runScheduler(session *sdtbu.ClientSession) {
	pushTimes, err := parsePushTimeTable()
	if err != nil {
		log.Fatalf(ASNIColor.Red+"錯誤: 解析 PUSH_TIME_TABLE 失敗: %v"+ASNIColor.Reset, err)
	}
	if len(pushTimes) == 0 {
		log.Println(ASNIColor.Yellow + "警告: PUSH_TIME_TABLE 未設定或沒有有效時間，排程器將不會觸發推送。" + ASNIColor.Reset)
		return
	}

	// 標記排程器正在運行
	globalSchedulerStatus.mu.Lock()
	globalSchedulerStatus.IsRunning = true
	globalSchedulerStatus.mu.Unlock()

	// pushedToday 追蹤當天哪些時間點已經推送過，防止重複推送
	var pushedToday = make(map[string]bool)
	var lastCheckedDay = time.Now().Day() // 記錄上次檢查的日期，用於每日重置

	for {
		now := time.Now()

		// 在午夜時重置 pushedToday map
		if now.Day() != lastCheckedDay {
			pushedToday = make(map[string]bool)
			lastCheckedDay = now.Day()
			log.Println(ASNIColor.BrightGreen + "已重置每日推送狀態。" + ASNIColor.Reset)
		}

		var nextPushTime time.Time
		foundNext := false

		// 尋找今天尚未推送且在當前時間之後的下一個排程時間
		for _, pt := range pushTimes {
			scheduledTime := time.Date(now.Year(), now.Month(), now.Day(), pt.Hour, pt.Minute, 0, 0, now.Location())
			timeStr := fmt.Sprintf("%02d:%02d", pt.Hour, pt.Minute)

			// 如果排程時間在未來並且今天尚未推送
			if scheduledTime.After(now) && !pushedToday[timeStr] {
				nextPushTime = scheduledTime
				foundNext = true
				break // 找到最早的下一個時間點，退出循環
			}
		}

		if foundNext {
			sleepDuration := nextPushTime.Sub(now)

			// 更新全局排程器狀態
			globalSchedulerStatus.mu.Lock()
			globalSchedulerStatus.NextPushTime = nextPushTime
			globalSchedulerStatus.SleepDuration = sleepDuration // 這裡仍然更新，但 /status 將重新計算
			globalSchedulerStatus.mu.Unlock()

			log.Printf(ASNIColor.BrightBlue+"下一次推送將在 %s 觸發 (剩餘 %s)。"+ASNIColor.Reset, nextPushTime.Format("15:04:05"), sleepDuration)
			time.Sleep(sleepDuration) // 休眠直到下一個推送時間

			// 喚醒後再次檢查時間，以處理輕微延遲或系統時鐘變化
			currentCheckTime := time.Now()
			timeStr := fmt.Sprintf("%02d:%02d", nextPushTime.Hour(), nextPushTime.Minute())

			// 檢查當前時間是否在預定時間附近 (例如 +/- 1 分鐘) 且尚未推送
			if currentCheckTime.After(nextPushTime.Add(-1*time.Minute)) && currentCheckTime.Before(nextPushTime.Add(1*time.Minute)) && !pushedToday[timeStr] {
				log.Println(ASNIColor.BrightGreen + "觸發課程推送！" + ASNIColor.Reset)
				// 在推送前獲取最新的課程資訊
				classInfo, err := fetchAndProcessClassData(session)
				if err != nil {
					log.Printf(ASNIColor.Red+"錯誤: 獲取課程資訊失敗: %v"+ASNIColor.Reset, err)
				} else if classInfo != nil {
					courseName, teacherName, location, timeNumber := extractClassInfo(classInfo)
					sendWxPushNotification(courseName, teacherName, location, timeNumber)
				} else {
					log.Println(ASNIColor.Yellow + "沒有找到下一節課資訊，跳過推送。" + ASNIColor.Reset)
				}
				pushedToday[timeStr] = true // 標記為已推送
			} else {
				log.Printf(ASNIColor.Yellow+"警告: 已過預定推送時間 %s 或已推送，跳過本次觸發。", nextPushTime.Format("15:04")+ASNIColor.Reset)
			}

			// 短暫休眠，避免在多個時間點非常接近時導致忙碌等待
			time.Sleep(1 * time.Second)

		} else {
			// 今天所有排程時間都已過或已推送。休眠直到明天的第一個排程時間。
			tomorrow := now.Add(24 * time.Hour)
			firstPushTimeTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), pushTimes[0].Hour, pushTimes[0].Minute, 0, 0, now.Location())
			sleepDuration := firstPushTimeTomorrow.Sub(now)

			// 更新全局排程器狀態
			globalSchedulerStatus.mu.Lock()
			globalSchedulerStatus.NextPushTime = firstPushTimeTomorrow
			globalSchedulerStatus.SleepDuration = sleepDuration // 這裡仍然更新，但 /status 將重新計算
			globalSchedulerStatus.mu.Unlock()

			log.Printf(ASNIColor.BrightBlue+"今天所有推送已完成。將在 %s 重新開始排程 (剩餘 %s)。"+ASNIColor.Reset, firstPushTimeTomorrow.Format("2006-01-02 15:04:05"), sleepDuration)
			time.Sleep(sleepDuration) // 休眠直到明天
		}
	}
}

// handleUserInput 處理用戶在控制台的輸入
func handleUserInput(session *sdtbu.ClientSession) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(ASNIColor.BrightGreen + "排程器已啟動。輸入 /nextcourse 查看下一節課，或輸入 /status 檢查狀態，輸入 /clear 清除控制台。" + ASNIColor.Reset)
	fmt.Print(ASNIColor.BrightBlue + "> " + ASNIColor.Reset) // 初始提示符

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf(ASNIColor.Red+"讀取輸入錯誤: %v\n"+ASNIColor.Reset, err)
			continue
		}

		command := strings.TrimSpace(input)

		switch strings.ToLower(command) {
		case "/nextcourse":
			fmt.Println(ASNIColor.BrightCyan + "正在獲取下一節課程資訊..." + ASNIColor.Reset)
			classInfo, err := fetchAndProcessClassData(session)
			if err != nil {
				fmt.Printf(ASNIColor.Red+"錯誤: 無法獲取課程資訊: %v\n"+ASNIColor.Reset, err)
			} else if classInfo != nil {
				courseName, teacherName, location, timeNumber := extractClassInfo(classInfo)
				fmt.Println(ASNIColor.BrightYellow + "下一節課程資訊：" + ASNIColor.Reset)
				fmt.Printf("課程名稱: %s\n", courseName)
				fmt.Printf("教師姓名: %s\n", teacherName)
				fmt.Printf("上課地點: %s\n", location)
				fmt.Printf("上課節次: %s\n", timeNumber)
				fmt.Printf("額外備註: %s\n", "N/A") // 這裡可以根據需要添加更多資訊
			} else {
				fmt.Println(ASNIColor.Yellow + "沒有找到下一節課資訊。" + ASNIColor.Reset)
			}
		case "/status":
			globalSchedulerStatus.mu.Lock() // 鎖定互斥鎖以安全讀取狀態
			isRunning := globalSchedulerStatus.IsRunning
			nextTime := globalSchedulerStatus.NextPushTime
			globalSchedulerStatus.mu.Unlock() // 解鎖

			if isRunning {
				if !nextTime.IsZero() { // 檢查是否有排程時間
					// 實時計算剩餘時間
					remainingDuration := time.Until(nextTime)
					// 輸出格式：(排程器正常運行中，下一次 HH:MM:SS（剩餘 XhYmZs）)
					fmt.Printf(ASNIColor.BrightGreen+"排程器正常運行中，下一次 %s（剩餘 %s）\n"+ASNIColor.Reset,
						nextTime.Format("15:04:05"), remainingDuration.Round(time.Second))
				} else {
					fmt.Println(ASNIColor.BrightGreen + "排程器正常運行中，但目前沒有找到下一次觸發時間。" + ASNIColor.Reset)
				}
			} else {
				fmt.Println(ASNIColor.Yellow + "排程器尚未啟動或已停止。" + ASNIColor.Reset)
			}
		case "/clear": // 新增：處理 /clear 命令
			// ANSI escape code to clear the screen and move cursor to home
			fmt.Print("\033[H\033[2J")
			printBanner() // 清除後重新打印橫幅
			fmt.Println(ASNIColor.BrightGreen + "控制台已清除。輸入 /nextcourse 查看下一節課，或輸入 /status 檢查狀態，輸入 /clear 清除控制台。" + ASNIColor.Reset)
		case "": // 如果用戶只按了 Enter
			globalSchedulerStatus.mu.Lock() // 鎖定互斥鎖以安全讀取狀態
			isRunning := globalSchedulerStatus.IsRunning
			nextTime := globalSchedulerStatus.NextPushTime
			globalSchedulerStatus.mu.Unlock() // 解鎖

			if isRunning {
				if !nextTime.IsZero() { // 檢查是否有排程時間
					// 實時計算剩餘時間
					remainingDuration := time.Until(nextTime)
					fmt.Printf(ASNIColor.BrightGreen+"排程器正常運行中，下一次 %s（剩餘 %s）\n"+ASNIColor.Reset,
						nextTime.Format("15:04:05"), remainingDuration.Round(time.Second))
				} else {
					fmt.Println(ASNIColor.BrightGreen + "排程器正常運行中，但目前沒有找到下一次觸發時間。" + ASNIColor.Reset)
				}
			} else {
				fmt.Println(ASNIColor.Yellow + "排程器尚未啟動或已停止。" + ASNIColor.Reset)
			}
		default:
			fmt.Printf(ASNIColor.Yellow+"未知指令: %s\n"+ASNIColor.Reset, command)
		}
		fmt.Print(ASNIColor.BrightBlue + "> " + ASNIColor.Reset) // 每次處理完畢後再次顯示提示符
	}
}

func main() {
	printBanner()

	session, err := initializeSession()
	if err != nil {
		log.Fatalf("%v", err) // 如果會話無法初始化，則終止程式
	}

	// 在一個新的 Goroutine 中啟動排程器，讓它在背景運行
	go runScheduler(session)

	// 在主 Goroutine 中處理用戶輸入
	handleUserInput(session)
}
