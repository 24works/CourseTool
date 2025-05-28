package main

import (
	ASNIColor "CourseTool/asnicolor"
	_ "CourseTool/configloader" // Import for side effect: load .env
	"CourseTool/sdtbu"
	"CourseTool/wxpush"
	"fmt"
	"log"
	"os"
)

func main() {
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
説明：基於Golang的課程工具，提供課程提醒功能。
使用：請確保已正確配置環境變數，然後運行此程序。
網址：https://www.ric.moe
GitHub：https://github.com/RichardMiku/CourseTool
=============================================================
	` + ASNIColor.Reset)

	// The .env file loading is now handled by the configloader package's init()

	sdtbu.Init() // 初始化您的套件

	// 創建一個新的客戶端會話
	session, err := sdtbu.NewClientSession()
	if err != nil {
		log.Fatalf("Failed to create client session: %v", err)
	}

	// 使用這個會話進行登入
	username := os.Getenv("SDTBU_USERNAME")
	password := os.Getenv("SDTBU_PASSWORD")

	if username == "" || password == "" {
		log.Fatalf(ASNIColor.Red + "錯誤: 環境變數 SDTBU_USERNAME 或 SDTBU_PASSWORD 未設定。" + ASNIColor.Reset)
	}

	err = session.Login(username, password)
	if err != nil {
		// 登入失敗時，也使用帶有顏色的日誌
		log.Fatalf(ASNIColor.Red+"登入失敗: %v"+ASNIColor.Reset, err)
	}

	//fmt.Println(ASNIColor.BrightWhite + ASNIColor.BgCyan + "Login successful!" + ASNIColor.Reset)

	// 現在您可以使用 `session` 對象在其他函數中執行需要登入狀態的操作
	// 例如：
	// session.FetchCourseList()
	// session.SubmitAssignment()

	session.GetClassbyUserInfo()
	err = session.GetClassbyTime()
	if err != nil {
		log.Fatalf(ASNIColor.Red+"獲取本周課程失敗: %v"+ASNIColor.Reset, err)
	}

	CLIST, err := session.ParseClassList(session.ClassListbyTimeString)
	if err != nil {
		log.Fatalf(ASNIColor.Red+"解析課程列表失敗: %v"+ASNIColor.Reset, err)
	}

	CLIST2, sortMsg := session.SortClass(CLIST)
	if len(CLIST2) == 0 {
		log.Println(ASNIColor.Yellow + "沒有課程可供排序或顯示: " + sortMsg + ASNIColor.Reset)
		// 根據需求決定是否在此處終止程序或執行其他操作
		// return
	}

	CLINFO, err := session.NextClass(CLIST2) // CLINFO 是 map[string]interface{}
	if err != nil {
		log.Printf(ASNIColor.Yellow+"獲取下一節課失敗: %v"+ASNIColor.Reset, err)
		// 在這裡處理沒有下一節課的情況，例如不發送微信推送或發送特定提示
		return // 如果沒有下一節課，則不繼續執行後續的微信推送
	}

	if CLINFO == nil {
		log.Println(ASNIColor.Yellow + "今天沒有更多課程了，或未找到下一節課資訊。" + ASNIColor.Reset)
		return // 如果 CLINFO 為 nil，則不繼續執行後續的微信推送
	}

	// fmt.Println(ASNIColor.BrightWhite + ASNIColor.BgCyan + "下節課程信息：" + ASNIColor.Reset)
	// fmt.Printf("%+v\n", CLINFO) // 使用 %+v 打印更詳細的 map 內容

	// 從 CLINFO 中安全地提取課程資訊
	courseNameVal, ok := CLINFO["KCMC"].(string)
	if !ok {
		log.Println(ASNIColor.Yellow + "警告: 課程名稱 (KCMC) 未在 CLINFO 中找到或其類型非字符串。" + ASNIColor.Reset)
		courseNameVal = "未知課程"
	}

	teacherNameVal, ok := CLINFO["JSXM"].(string) // 假設教師姓名鍵為 JSXM
	if !ok {
		// 嘗試 JSMC 作為備用鍵，因為 sdtbu.go 的註釋中提到過
		teacherNameVal, ok = CLINFO["JSMC"].(string)
		if !ok {
			log.Println(ASNIColor.Yellow + "警告: 教師姓名 (JSXM/JSMC) 未在 CLINFO 中找到或其類型非字符串。" + ASNIColor.Reset)
			teacherNameVal = "未知教師"
		}
	}

	locationVal, ok := CLINFO["JXDD"].(string) // 假設上課地點鍵為 JKDD
	if !ok {
		// 嘗試 JASMC 作為備用鍵
		locationVal, ok = CLINFO["JASMC"].(string)
		if !ok {
			log.Println(ASNIColor.Yellow + "警告: 上課地點 (JKDD/JASMC) 未在 CLINFO 中找到或其類型非字符串。" + ASNIColor.Reset)
			locationVal = "未知地點"
		}
	}

	var timeNumberStr string
	skjcVal, ok := CLINFO["SKJC"] // 上課節次
	if !ok {
		log.Println(ASNIColor.Yellow + "警告: 上課節次 (SKJC) 未在 CLINFO 中找到。" + ASNIColor.Reset)
		timeNumberStr = "未知時間"
	} else {
		skjcFloat, ok := skjcVal.(float64) // JSON 數字通常解析為 float64
		if !ok {
			log.Printf(ASNIColor.Yellow+"警告: 上課節次 (SKJC) 類型不是 float64，實際為 %T。"+ASNIColor.Reset, skjcVal)
			timeNumberStr = "未知時間"
		} else {
			timeNumberStr, err = sdtbu.GetFormattedClassTime(int(skjcFloat))
			if err != nil {
				log.Printf(ASNIColor.Yellow+"警告: 獲取格式化課程時間失敗: %v"+ASNIColor.Reset, err)
				timeNumberStr = "未知時間"
			}
		}
	}

	// 獲取微信 Access Token
	accessToken, err := wxpush.GetAccessToken()
	if err != nil {
		log.Fatalf(ASNIColor.Red+"獲取微信 Access Token 失敗: %v"+ASNIColor.Reset, err)
	}

	courseData := wxpush.CourseReminderData{
		CourseName:     courseNameVal,
		TeacherName:    teacherNameVal,
		CourseLocation: locationVal,
		TimeNumber:     timeNumberStr,
		Note:           "記得帶上好心情去上課哦！", // 可以自定義備註
	}

	err = wxpush.SendCourseReminder(accessToken, courseData)
	if err != nil {
		fmt.Printf(ASNIColor.Red+"發送課程提醒失敗: %v"+ASNIColor.Reset+"\n", err)
	}
}
